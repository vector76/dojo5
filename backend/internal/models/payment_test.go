package models

import (
	"database/sql"
	"fmt"
	"testing"
)

func createTestPayer(t *testing.T, repo *UserRepo, email string, expectedBalance float64) *User {
	t.Helper()
	u := &User{
		Name:            "Payer",
		Email:           email,
		Phone:           "555-0000",
		Role:            "user",
		PasswordHash:    "hashed",
		ExpectedBalance: expectedBalance,
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("failed to create payer: %v", err)
	}
	return u
}

func createTestAdmin(t *testing.T, repo *UserRepo, email string) *User {
	t.Helper()
	u := &User{
		Name:         "Admin",
		Email:        email,
		Phone:        "555-9999",
		Role:         "admin",
		PasswordHash: "hashed",
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}
	return u
}

func TestPaymentRepo_RecordPayment(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	repo := NewPaymentRepo(db)

	payer := createTestPayer(t, userRepo, "rp-payer@example.com", 100.0)
	admin := createTestAdmin(t, userRepo, "rp-admin@example.com")

	tests := []struct {
		name    string
		payment *Payment
		wantErr bool
	}{
		{
			name: "valid payment",
			payment: &Payment{
				UserID: payer.ID, Amount: 50.0, Date: "2026-03-01",
				RecordedBy: admin.ID,
			},
			wantErr: false,
		},
		{
			name: "payment with note",
			payment: func() *Payment {
				note := "Monthly fee"
				return &Payment{
					UserID: payer.ID, Amount: 25.0, Date: "2026-03-15",
					Note: &note, RecordedBy: admin.ID,
				}
			}(),
			wantErr: false,
		},
		{
			name: "invalid user",
			payment: &Payment{
				UserID: 99999, Amount: 10.0, Date: "2026-03-01",
				RecordedBy: admin.ID,
			},
			wantErr: true,
		},
		{
			name: "invalid recorded_by",
			payment: &Payment{
				UserID: payer.ID, Amount: 10.0, Date: "2026-03-01",
				RecordedBy: 99999,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RecordPayment(tt.payment)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordPayment() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.payment.ID == 0 {
				t.Error("expected non-zero ID after record")
			}
		})
	}
}

func TestPaymentRepo_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	repo := NewPaymentRepo(db)

	user1 := createTestPayer(t, userRepo, "lbu-pay1@example.com", 200.0)
	user2 := createTestPayer(t, userRepo, "lbu-pay2@example.com", 100.0)
	admin := createTestAdmin(t, userRepo, "lbu-admin@example.com")

	// user1 has 3 payments, user2 has 1
	payments := []Payment{
		{UserID: user1.ID, Amount: 50.0, Date: "2026-01-15", RecordedBy: admin.ID},
		{UserID: user1.ID, Amount: 75.0, Date: "2026-02-15", RecordedBy: admin.ID},
		{UserID: user1.ID, Amount: 25.0, Date: "2026-03-15", RecordedBy: admin.ID},
		{UserID: user2.ID, Amount: 100.0, Date: "2026-01-15", RecordedBy: admin.ID},
	}
	for i := range payments {
		if err := repo.RecordPayment(&payments[i]); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    int64
		wantCount int
	}{
		{"user with 3 payments", user1.ID, 3},
		{"user with 1 payment", user2.ID, 1},
		{"user with no payments", 99999, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ListByUser(tt.userID)
			if err != nil {
				t.Fatalf("ListByUser() error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("expected %d records, got %d", tt.wantCount, len(got))
			}
			for _, p := range got {
				if p.UserID != tt.userID {
					t.Errorf("expected user_id %d, got %d", tt.userID, p.UserID)
				}
			}
		})
	}

	// Verify ordering by date DESC (most recent first)
	got, err := repo.ListByUser(user1.ID)
	if err != nil {
		t.Fatalf("ListByUser() error: %v", err)
	}
	if len(got) >= 2 {
		if got[0].Date < got[1].Date {
			t.Errorf("payments not ordered by date DESC: %s before %s", got[0].Date, got[1].Date)
		}
	}
}

func TestPaymentRepo_GetBalance(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	repo := NewPaymentRepo(db)

	admin := createTestAdmin(t, userRepo, "bal-admin@example.com")

	tests := []struct {
		name            string
		expectedBalance float64
		payments        []float64
		wantBalance     float64
	}{
		{
			name:            "no payments",
			expectedBalance: 200.0,
			payments:        nil,
			wantBalance:     200.0,
		},
		{
			name:            "partial payment",
			expectedBalance: 200.0,
			payments:        []float64{50.0},
			wantBalance:     150.0,
		},
		{
			name:            "multiple payments",
			expectedBalance: 300.0,
			payments:        []float64{100.0, 75.0, 50.0},
			wantBalance:     75.0,
		},
		{
			name:            "fully paid",
			expectedBalance: 100.0,
			payments:        []float64{100.0},
			wantBalance:     0.0,
		},
		{
			name:            "overpaid",
			expectedBalance: 100.0,
			payments:        []float64{150.0},
			wantBalance:     -50.0,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestPayer(t, userRepo, fmt.Sprintf("bal-user%d@example.com", i), tt.expectedBalance)
			for _, amount := range tt.payments {
				p := &Payment{UserID: user.ID, Amount: amount, Date: "2026-03-01", RecordedBy: admin.ID}
				if err := repo.RecordPayment(p); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}

			balance, err := repo.GetBalance(user.ID)
			if err != nil {
				t.Fatalf("GetBalance() error: %v", err)
			}
			if balance != tt.wantBalance {
				t.Errorf("expected balance %.2f, got %.2f", tt.wantBalance, balance)
			}
		})
	}

	// Non-existent user
	_, err := repo.GetBalance(99999)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows for non-existent user, got %v", err)
	}
}
