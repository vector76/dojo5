package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func requestWithRole(role string) *http.Request {
	claims := &Claims{UserID: 1, Role: role}
	ctx := context.WithValue(context.Background(), contextKey{}, claims)
	return httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
}

func TestRequireRole(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		allowedRoles []string
		requestRole  string
		wantStatus   int
	}{
		{
			name:         "admin allowed for admin-only",
			allowedRoles: []string{"admin"},
			requestRole:  "admin",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "user denied for admin-only",
			allowedRoles: []string{"admin"},
			requestRole:  "user",
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "instructor denied for admin-only",
			allowedRoles: []string{"admin"},
			requestRole:  "instructor",
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "admin allowed for admin+instructor",
			allowedRoles: []string{"admin", "instructor"},
			requestRole:  "admin",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "instructor allowed for admin+instructor",
			allowedRoles: []string{"admin", "instructor"},
			requestRole:  "instructor",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "user denied for admin+instructor",
			allowedRoles: []string{"admin", "instructor"},
			requestRole:  "user",
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "admin allowed for any role",
			allowedRoles: []string{"admin", "instructor", "user"},
			requestRole:  "admin",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "user allowed for any role",
			allowedRoles: []string{"admin", "instructor", "user"},
			requestRole:  "user",
			wantStatus:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireRole(tt.allowedRoles...)(okHandler)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, requestWithRole(tt.requestRole))

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestRequireRoleNoClaimsInContext(t *testing.T) {
	handler := RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no claims, got %d", rec.Code)
	}
}
