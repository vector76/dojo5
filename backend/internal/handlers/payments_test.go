package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dojo-crm/backend/internal/models"
)

func TestPaymentHandler_Create(t *testing.T) {
	setup := func(t *testing.T) (http.Handler, *models.PaymentRepo, *models.User, *models.User) {
		t.Helper()
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewPaymentHandler(paymentRepo)
		admin := createTestUser(t, userRepo, "pay-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "pay-user@example.com", "password", "user")
		return wrapWithMiddleware(handler.Create, "admin"), paymentRepo, user, admin
	}

	t.Run("admin records payment", func(t *testing.T) {
		h, _, user, admin := setup(t)

		body, _ := json.Marshal(map[string]any{
			"user_id": user.ID,
			"amount":  50.00,
			"date":    "2026-03-01",
			"note":    "Monthly fee",
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/payments", body, admin.ID, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp paymentResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if resp.UserID != user.ID {
			t.Errorf("expected user_id %d, got %d", user.ID, resp.UserID)
		}
		if resp.Amount != 50.00 {
			t.Errorf("expected amount 50, got %f", resp.Amount)
		}
		if resp.Date != "2026-03-01" {
			t.Errorf("expected date '2026-03-01', got %q", resp.Date)
		}
		if resp.Note == nil || *resp.Note != "Monthly fee" {
			t.Errorf("expected note 'Monthly fee', got %v", resp.Note)
		}
		if resp.RecordedBy != admin.ID {
			t.Errorf("expected recorded_by %d, got %d", admin.ID, resp.RecordedBy)
		}
	})

	t.Run("admin records payment without note", func(t *testing.T) {
		h, _, user, admin := setup(t)

		body, _ := json.Marshal(map[string]any{
			"user_id": user.ID,
			"amount":  25.50,
			"date":    "2026-03-15",
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/payments", body, admin.ID, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp paymentResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Note != nil {
			t.Errorf("expected nil note, got %v", resp.Note)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		h, _, _, admin := setup(t)

		body, _ := json.Marshal(map[string]any{"user_id": 1})
		req, _ := authedRequest(t, http.MethodPost, "/api/payments", body, admin.ID, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot create", func(t *testing.T) {
		h, _, user, _ := setup(t)

		body, _ := json.Marshal(map[string]any{
			"user_id": user.ID,
			"amount":  50.00,
			"date":    "2026-03-01",
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/payments", body, 1, "instructor")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("user cannot create", func(t *testing.T) {
		h, _, user, _ := setup(t)

		body, _ := json.Marshal(map[string]any{
			"user_id": user.ID,
			"amount":  50.00,
			"date":    "2026-03-01",
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/payments", body, 1, "user")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		h, _, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/payments", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestPaymentHandler_ListByUser(t *testing.T) {
	t.Run("admin can view any user payments", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewPaymentHandler(paymentRepo)
		admin := createTestUser(t, userRepo, "pay-list-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "pay-list-user@example.com", "password", "user")

		// Record two payments
		p1 := &models.Payment{UserID: user.ID, Amount: 50.00, Date: "2026-03-01", RecordedBy: admin.ID}
		if err := paymentRepo.RecordPayment(p1); err != nil {
			t.Fatalf("failed to record payment: %v", err)
		}
		note := "Second payment"
		p2 := &models.Payment{UserID: user.ID, Amount: 75.00, Date: "2026-03-15", Note: &note, RecordedBy: admin.ID}
		if err := paymentRepo.RecordPayment(p2); err != nil {
			t.Fatalf("failed to record payment: %v", err)
		}

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(user.ID)+"/payments", nil, admin.ID, "admin")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp []paymentResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 payments, got %d", len(resp))
		}
	})

	t.Run("user can view own payments", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewPaymentHandler(paymentRepo)
		admin := createTestUser(t, userRepo, "pay-self-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "pay-self@example.com", "password", "user")

		p := &models.Payment{UserID: user.ID, Amount: 50.00, Date: "2026-03-01", RecordedBy: admin.ID}
		if err := paymentRepo.RecordPayment(p); err != nil {
			t.Fatalf("failed to record payment: %v", err)
		}

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(user.ID)+"/payments", nil, user.ID, "user")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []paymentResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 1 {
			t.Errorf("expected 1 payment, got %d", len(resp))
		}
	})

	t.Run("user cannot view other user payments", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewPaymentHandler(paymentRepo)
		admin := createTestUser(t, userRepo, "pay-other-admin@example.com", "password", "admin")
		user1 := createTestUser(t, userRepo, "pay-other1@example.com", "password", "user")
		user2 := createTestUser(t, userRepo, "pay-other2@example.com", "password", "user")

		p := &models.Payment{UserID: user1.ID, Amount: 50.00, Date: "2026-03-01", RecordedBy: admin.ID}
		if err := paymentRepo.RecordPayment(p); err != nil {
			t.Fatalf("failed to record payment: %v", err)
		}

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "user")

		// user2 trying to view user1's payments
		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(user1.ID)+"/payments", nil, user2.ID, "user")
		req.SetPathValue("id", itoa(user1.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("empty payment history", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewPaymentHandler(paymentRepo)
		admin := createTestUser(t, userRepo, "pay-empty-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "pay-empty@example.com", "password", "user")

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(user.ID)+"/payments", nil, admin.ID, "admin")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []paymentResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 0 {
			t.Errorf("expected 0 payments, got %d", len(resp))
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		db := setupTestDB(t)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewPaymentHandler(paymentRepo)

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "user")

		req := httptest.NewRequest(http.MethodGet, "/api/users/1/payments", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}
