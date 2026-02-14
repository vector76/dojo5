package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/models"
)

// withAuth wraps a handler with AuthMiddleware and RequireRole, then creates
// a request with a valid JWT for the given role.
func authedRequest(t *testing.T, method, url string, body []byte, userID int64, role string) (*http.Request, string) {
	t.Helper()
	token, err := auth.GenerateToken(testJWTSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req, token
}

func wrapWithMiddleware(handler http.HandlerFunc, roles ...string) http.Handler {
	h := http.Handler(handler)
	h = auth.RequireRole(roles...)(h)
	h = auth.AuthMiddleware(testJWTSecret)(h)
	return h
}

func TestUserHandler_List(t *testing.T) {
	db := setupTestDB(t)
	repo := models.NewUserRepo(db)
	handler := NewUserHandler(repo)

	// Create some users
	admin := createTestUser(t, repo, "list-admin@example.com", "password", "admin")
	createTestUser(t, repo, "list-inst@example.com", "password", "instructor")
	createTestUser(t, repo, "list-user@example.com", "password", "user")

	// Soft-delete one user
	deleted := createTestUser(t, repo, "list-deleted@example.com", "password", "user")
	if err := repo.SoftDelete(deleted.ID); err != nil {
		t.Fatalf("setup: %v", err)
	}

	wrapped := wrapWithMiddleware(handler.List, "admin", "instructor")

	tests := []struct {
		name       string
		role       string
		userID     int64
		wantStatus int
		wantCount  int
	}{
		{
			name:       "admin can list users",
			role:       "admin",
			userID:     admin.ID,
			wantStatus: http.StatusOK,
			wantCount:  3, // excludes soft-deleted
		},
		{
			name:       "instructor can list users",
			role:       "instructor",
			userID:     2,
			wantStatus: http.StatusOK,
			wantCount:  3,
		},
		{
			name:       "regular user is forbidden",
			role:       "user",
			userID:     3,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := authedRequest(t, http.MethodGet, "/api/users", nil, tt.userID, tt.role)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var users []userResponse
				if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(users) != tt.wantCount {
					t.Errorf("expected %d users, got %d", tt.wantCount, len(users))
				}
			}
		})
	}

	// Unauthenticated request
	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestUserHandler_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := models.NewUserRepo(db)
	handler := NewUserHandler(repo)

	admin := createTestUser(t, repo, "create-admin@example.com", "password", "admin")
	createTestUser(t, repo, "create-inst@example.com", "password", "instructor")

	wrapped := wrapWithMiddleware(handler.Create, "admin")

	tests := []struct {
		name       string
		role       string
		userID     int64
		body       map[string]string
		wantStatus int
	}{
		{
			name:   "admin creates user",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"name": "New User", "email": "new@example.com",
				"phone": "555-0000", "password": "secret123",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "admin creates instructor",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"name": "New Instructor", "email": "newinst@example.com",
				"phone": "555-0001", "password": "secret123", "role": "instructor",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "duplicate email",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"name": "Another User", "email": "new@example.com",
				"phone": "555-0002", "password": "secret123",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "missing name",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"email": "noname@example.com",
				"phone": "555-0003", "password": "secret123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "missing email",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"name": "No Email",
				"phone": "555-0004", "password": "secret123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "missing phone",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"name": "No Phone", "email": "nophone@example.com",
				"password": "secret123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "missing password",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"name": "No Pass", "email": "nopass@example.com",
				"phone": "555-0005",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid role",
			role:   "admin",
			userID: admin.ID,
			body: map[string]string{
				"name": "Bad Role", "email": "badrole@example.com",
				"phone": "555-0006", "password": "secret123", "role": "superadmin",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "instructor cannot create",
			role:   "instructor",
			userID: 2,
			body: map[string]string{
				"name": "Blocked", "email": "blocked@example.com",
				"phone": "555-0007", "password": "secret123",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "regular user cannot create",
			role:   "user",
			userID: 99,
			body: map[string]string{
				"name": "Blocked", "email": "blocked2@example.com",
				"phone": "555-0008", "password": "secret123",
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := authedRequest(t, http.MethodPost, "/api/users", bodyBytes, tt.userID, tt.role)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tt.wantStatus, rec.Code, rec.Body.String())
			}

			if tt.wantStatus == http.StatusCreated {
				var resp userResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.ID == 0 {
					t.Error("expected non-zero ID")
				}
				if resp.Email != tt.body["email"] {
					t.Errorf("expected email %q, got %q", tt.body["email"], resp.Email)
				}
			}
		})
	}

	// Verify default role is "user" when not specified
	t.Run("default role is user", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name": "Default Role", "email": "defaultrole@example.com",
			"phone": "555-9999", "password": "secret123",
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/users", body, admin.ID, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
		var resp userResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Role != "user" {
			t.Errorf("expected default role 'user', got %q", resp.Role)
		}
	})

	// Response should not include password hash
	t.Run("response excludes password hash", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name": "No Hash", "email": "nohash@example.com",
			"phone": "555-8888", "password": "secret123",
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/users", body, admin.ID, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
		// Decode as raw map to check no password_hash field
		var raw map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if _, ok := raw["password_hash"]; ok {
			t.Error("response should not include password_hash")
		}
		if _, ok := raw["password"]; ok {
			t.Error("response should not include password")
		}
	})
}
