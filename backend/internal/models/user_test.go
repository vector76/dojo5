package models

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	"dojo-crm/backend/internal/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := database.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, f, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine caller")
	}
	migrationsDir := filepath.Join(filepath.Dir(f), "..", "..", "migrations")
	if err := database.Migrate(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	return db
}

func newTestUser(email string) *User {
	return &User{
		Name:         "Test User",
		Email:        email,
		Phone:        "555-1234",
		Role:         "user",
		PasswordHash: "hashed_password",
	}
}

func TestUserRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name:    "valid user",
			user:    newTestUser("create@example.com"),
			wantErr: false,
		},
		{
			name: "with optional fields",
			user: func() *User {
				u := newTestUser("optional@example.com")
				mt := "monthly"
				ms := "active"
				ec := "Jane Doe 555-5678"
				jd := "2025-01-15"
				u.MembershipType = &mt
				u.MembershipStatus = &ms
				u.EmergencyContact = &ec
				u.JoinDate = &jd
				u.ExpectedBalance = 100.50
				return u
			}(),
			wantErr: false,
		},
		{
			name:    "duplicate email",
			user:    newTestUser("create@example.com"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.user.ID == 0 {
				t.Error("expected non-zero ID after create")
			}
		})
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	u := newTestUser("getbyid@example.com")
	if err := repo.Create(u); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"existing user", u.ID, false},
		{"non-existent user", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if got.ID != u.ID {
					t.Errorf("expected ID %d, got %d", u.ID, got.ID)
				}
				if got.Email != "getbyid@example.com" {
					t.Errorf("expected email getbyid@example.com, got %q", got.Email)
				}
				if got.Role != "user" {
					t.Errorf("expected role user, got %q", got.Role)
				}
			}
		})
	}
}

func TestUserRepo_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	u := newTestUser("byemail@example.com")
	if err := repo.Create(u); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"existing email", "byemail@example.com", false},
		{"non-existent email", "nosuch@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if got.Email != tt.email {
					t.Errorf("expected email %q, got %q", tt.email, got.Email)
				}
			}
		})
	}
}

func TestUserRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	// Create some users
	for _, email := range []string{"alice@example.com", "bob@example.com", "charlie@example.com"} {
		u := newTestUser(email)
		u.Name = email[:len(email)-len("@example.com")]
		if err := repo.Create(u); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	// Soft-delete one
	if err := repo.SoftDelete(2); err != nil {
		t.Fatalf("setup soft delete: %v", err)
	}

	users, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users (excluding soft-deleted), got %d", len(users))
	}

	// Verify sorted by name
	if len(users) == 2 {
		if users[0].Name != "alice" {
			t.Errorf("expected first user alice, got %q", users[0].Name)
		}
		if users[1].Name != "charlie" {
			t.Errorf("expected second user charlie, got %q", users[1].Name)
		}
	}
}

func TestUserRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	u := newTestUser("update@example.com")
	if err := repo.Create(u); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Update fields
	u.Name = "Updated Name"
	u.Phone = "555-9999"
	mt := "annual"
	u.MembershipType = &mt

	if err := repo.Update(u); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	// Fetch and verify
	got, err := repo.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if got.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", got.Name)
	}
	if got.Phone != "555-9999" {
		t.Errorf("expected phone '555-9999', got %q", got.Phone)
	}
	if got.MembershipType == nil || *got.MembershipType != "annual" {
		t.Errorf("expected membership_type 'annual', got %v", got.MembershipType)
	}
}

func TestUserRepo_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	u := newTestUser("delete@example.com")
	if err := repo.Create(u); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"existing user", u.ID, false},
		{"already deleted", u.ID, true},
		{"non-existent user", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SoftDelete(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SoftDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify user still exists but has deleted_at set
	got, err := repo.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID after soft delete: %v", err)
	}
	if got.DeletedAt == nil {
		t.Error("expected deleted_at to be set after soft delete")
	}

	// Verify not in list
	users, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	for _, lu := range users {
		if lu.ID == u.ID {
			t.Error("soft-deleted user should not appear in List()")
		}
	}
}
