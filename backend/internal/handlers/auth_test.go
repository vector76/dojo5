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
