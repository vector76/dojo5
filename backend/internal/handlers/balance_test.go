package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dojo-crm/backend/internal/models"
)

func TestBalanceHandler_Get(t *testing.T) {
	t.Run("admin can view any user balance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		admin := createTestUser(t, userRepo, "bal-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "bal-user@example.com", "password", "user")

		// Set expected balance
		user.ExpectedBalance = 200.00
		if err := userRepo.Update(user); err != nil {
			t.Fatalf("failed to update user: %v", err)
		}

		// Record a payment
		p := &models.Payment{UserID: user.ID, Amount: 50.00, Date: "2026-03-01", RecordedBy: admin.ID}
		if err := paymentRepo.RecordPayment(p); err != nil {
			t.Fatalf("failed to record payment: %v", err)
		}

		wrapped := wrapWithMiddleware(handler.Get, "admin", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(user.ID)+"/balance", nil, admin.ID, "admin")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp balanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.UserID != user.ID {
			t.Errorf("expected user_id %d, got %d", user.ID, resp.UserID)
		}
		if resp.ExpectedBalance != 200.00 {
			t.Errorf("expected expected_balance 200, got %f", resp.ExpectedBalance)
		}
		if resp.TotalPaid != 50.00 {
			t.Errorf("expected total_paid 50, got %f", resp.TotalPaid)
		}
		if resp.Balance != 150.00 {
			t.Errorf("expected balance 150, got %f", resp.Balance)
		}
	})

	t.Run("user can view own balance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		user := createTestUser(t, userRepo, "bal-self@example.com", "password", "user")

		user.ExpectedBalance = 100.00
		if err := userRepo.Update(user); err != nil {
			t.Fatalf("failed to update user: %v", err)
		}

		wrapped := wrapWithMiddleware(handler.Get, "admin", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(user.ID)+"/balance", nil, user.ID, "user")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp balanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Balance != 100.00 {
			t.Errorf("expected balance 100 (no payments), got %f", resp.Balance)
		}
	})

	t.Run("user cannot view other user balance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		user1 := createTestUser(t, userRepo, "bal-other1@example.com", "password", "user")
		user2 := createTestUser(t, userRepo, "bal-other2@example.com", "password", "user")

		wrapped := wrapWithMiddleware(handler.Get, "admin", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(user1.ID)+"/balance", nil, user2.ID, "user")
		req.SetPathValue("id", itoa(user1.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("nonexistent user returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		createTestUser(t, userRepo, "bal-admin2@example.com", "password", "admin")

		wrapped := wrapWithMiddleware(handler.Get, "admin", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/9999/balance", nil, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)

		wrapped := wrapWithMiddleware(handler.Get, "admin", "user")

		req := httptest.NewRequest(http.MethodGet, "/api/users/1/balance", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestBalanceHandler_Set(t *testing.T) {
	t.Run("admin sets expected balance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		admin := createTestUser(t, userRepo, "bal-set-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "bal-set-user@example.com", "password", "user")

		// Record a payment first
		p := &models.Payment{UserID: user.ID, Amount: 30.00, Date: "2026-03-01", RecordedBy: admin.ID}
		if err := paymentRepo.RecordPayment(p); err != nil {
			t.Fatalf("failed to record payment: %v", err)
		}

		wrapped := wrapWithMiddleware(handler.Set, "admin")

		body, _ := json.Marshal(map[string]any{"expected_balance": 300.00})
		req, _ := authedRequest(t, http.MethodPut, "/api/users/"+itoa(user.ID)+"/balance", body, admin.ID, "admin")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp balanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ExpectedBalance != 300.00 {
			t.Errorf("expected expected_balance 300, got %f", resp.ExpectedBalance)
		}
		if resp.TotalPaid != 30.00 {
			t.Errorf("expected total_paid 30, got %f", resp.TotalPaid)
		}
		if resp.Balance != 270.00 {
			t.Errorf("expected balance 270, got %f", resp.Balance)
		}
	})

	t.Run("admin can set balance to zero", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		createTestUser(t, userRepo, "bal-zero-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "bal-zero-user@example.com", "password", "user")

		// Set initial balance
		user.ExpectedBalance = 100.00
		if err := userRepo.Update(user); err != nil {
			t.Fatalf("failed to update user: %v", err)
		}

		wrapped := wrapWithMiddleware(handler.Set, "admin")

		body, _ := json.Marshal(map[string]any{"expected_balance": 0})
		req, _ := authedRequest(t, http.MethodPut, "/api/users/"+itoa(user.ID)+"/balance", body, 1, "admin")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp balanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ExpectedBalance != 0 {
			t.Errorf("expected expected_balance 0, got %f", resp.ExpectedBalance)
		}
	})

	t.Run("missing expected_balance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		createTestUser(t, userRepo, "bal-miss-admin@example.com", "password", "admin")
		user := createTestUser(t, userRepo, "bal-miss-user@example.com", "password", "user")

		wrapped := wrapWithMiddleware(handler.Set, "admin")

		body, _ := json.Marshal(map[string]any{})
		req, _ := authedRequest(t, http.MethodPut, "/api/users/"+itoa(user.ID)+"/balance", body, 1, "admin")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("nonexistent user returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		createTestUser(t, userRepo, "bal-set404-admin@example.com", "password", "admin")

		wrapped := wrapWithMiddleware(handler.Set, "admin")

		body, _ := json.Marshal(map[string]any{"expected_balance": 100})
		req, _ := authedRequest(t, http.MethodPut, "/api/users/9999/balance", body, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("user cannot set balance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		user := createTestUser(t, userRepo, "bal-set-forbid@example.com", "password", "user")

		wrapped := wrapWithMiddleware(handler.Set, "admin")

		body, _ := json.Marshal(map[string]any{"expected_balance": 100})
		req, _ := authedRequest(t, http.MethodPut, "/api/users/"+itoa(user.ID)+"/balance", body, user.ID, "user")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot set balance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)
		user := createTestUser(t, userRepo, "bal-set-inst@example.com", "password", "user")

		wrapped := wrapWithMiddleware(handler.Set, "admin")

		body, _ := json.Marshal(map[string]any{"expected_balance": 100})
		req, _ := authedRequest(t, http.MethodPut, "/api/users/"+itoa(user.ID)+"/balance", body, 1, "instructor")
		req.SetPathValue("id", itoa(user.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		paymentRepo := models.NewPaymentRepo(db)
		handler := NewBalanceHandler(paymentRepo, userRepo)

		wrapped := wrapWithMiddleware(handler.Set, "admin")

		req := httptest.NewRequest(http.MethodPut, "/api/users/1/balance", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}
