package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/database"
	"dojo-crm/backend/internal/models"
)

const testJWTSecret = "test-secret-for-handlers"

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

func createTestUser(t *testing.T, repo *models.UserRepo, email, password, role string) *models.User {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	u := &models.User{
		Name:         "Test User",
		Email:        email,
		Phone:        "555-1234",
		Role:         role,
		PasswordHash: hash,
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return u
}

func TestMe(t *testing.T) {
	db := setupTestDB(t)
	repo := models.NewUserRepo(db)
	handler := NewAuthHandler(repo, testJWTSecret)

	user := createTestUser(t, repo, "me@example.com", "password", "instructor")

	wrapped := wrapWithMiddleware(handler.Me, "admin", "instructor", "user")

	t.Run("returns current user profile", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/auth/me", nil, user.ID, "instructor")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp userResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.ID != user.ID {
			t.Errorf("expected id %d, got %d", user.ID, resp.ID)
		}
		if resp.Email != "me@example.com" {
			t.Errorf("expected email me@example.com, got %q", resp.Email)
		}
		if resp.Role != "instructor" {
			t.Errorf("expected role instructor, got %q", resp.Role)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestSetupStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := models.NewUserRepo(db)
	handler := NewAuthHandler(repo, testJWTSecret)

	t.Run("needs setup when no users exist", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/setup-status", nil)
		rec := httptest.NewRecorder()
		handler.SetupStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp setupStatusResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if !resp.NeedsSetup {
			t.Error("expected needs_setup to be true")
		}
	})

	t.Run("does not need setup when users exist", func(t *testing.T) {
		createTestUser(t, repo, "existing@example.com", "password", "admin")

		req := httptest.NewRequest(http.MethodGet, "/api/auth/setup-status", nil)
		rec := httptest.NewRecorder()
		handler.SetupStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp setupStatusResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.NeedsSetup {
			t.Error("expected needs_setup to be false")
		}
	})
}

func TestSetup(t *testing.T) {
	t.Run("creates first admin", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewUserRepo(db)
		handler := NewAuthHandler(repo, testJWTSecret)

		body, _ := json.Marshal(map[string]string{
			"name": "Admin", "email": "admin@example.com",
			"phone": "555-0000", "password": "secret123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.Setup(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var resp loginResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Token == "" {
			t.Error("expected non-empty token")
		}

		// Verify token has admin role
		claims, err := auth.ValidateToken(testJWTSecret, resp.Token)
		if err != nil {
			t.Fatalf("token validation failed: %v", err)
		}
		if claims.Role != "admin" {
			t.Errorf("expected role admin, got %q", claims.Role)
		}

		// Verify user was created as admin
		user, err := repo.GetByEmail("admin@example.com")
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if user.Role != "admin" {
			t.Errorf("expected user role admin, got %q", user.Role)
		}
	})

	t.Run("blocked when users exist", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewUserRepo(db)
		handler := NewAuthHandler(repo, testJWTSecret)

		createTestUser(t, repo, "existing@example.com", "password", "admin")

		body, _ := json.Marshal(map[string]string{
			"name": "Another Admin", "email": "another@example.com",
			"phone": "555-0001", "password": "secret123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.Setup(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewUserRepo(db)
		handler := NewAuthHandler(repo, testJWTSecret)

		body, _ := json.Marshal(map[string]string{
			"email": "admin@example.com", "phone": "555-0000", "password": "secret123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.Setup(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing password", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewUserRepo(db)
		handler := NewAuthHandler(repo, testJWTSecret)

		body, _ := json.Marshal(map[string]string{
			"name": "Admin", "email": "admin@example.com", "phone": "555-0000",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.Setup(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	repo := models.NewUserRepo(db)
	handler := NewAuthHandler(repo, testJWTSecret)

	createTestUser(t, repo, "admin@example.com", "correct-password", "admin")

	// Create and soft-delete a user
	deleted := createTestUser(t, repo, "deleted@example.com", "password123", "user")
	if err := repo.SoftDelete(deleted.ID); err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantToken  bool
	}{
		{
			name:       "successful login",
			body:       map[string]string{"email": "admin@example.com", "password": "correct-password"},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:       "wrong password",
			body:       map[string]string{"email": "admin@example.com", "password": "wrong-password"},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name:       "nonexistent user",
			body:       map[string]string{"email": "nobody@example.com", "password": "anything"},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name:       "soft-deleted user",
			body:       map[string]string{"email": "deleted@example.com", "password": "password123"},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name:       "missing email",
			body:       map[string]string{"password": "anything"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "missing password",
			body:       map[string]string{"email": "admin@example.com"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantToken {
				var resp loginResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Token == "" {
					t.Error("expected non-empty token")
				}

				// Verify the token is valid and contains correct claims
				claims, err := auth.ValidateToken(testJWTSecret, resp.Token)
				if err != nil {
					t.Fatalf("token validation failed: %v", err)
				}
				if claims.Role != "admin" {
					t.Errorf("expected role admin, got %q", claims.Role)
				}
			}
		})
	}
}
