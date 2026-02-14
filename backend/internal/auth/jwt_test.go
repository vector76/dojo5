package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testSecret = "test-secret-key-for-jwt-testing"

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(testSecret, 42, "admin", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateTokenSuccess(t *testing.T) {
	token, err := GenerateToken(testSecret, 42, "admin", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := ValidateToken(testSecret, token)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}

	if claims.Role != "admin" {
		t.Errorf("expected Role \"admin\", got %q", claims.Role)
	}
}

func TestValidateTokenExpired(t *testing.T) {
	token, err := GenerateToken(testSecret, 1, "user", -time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ValidateToken(testSecret, token)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestValidateTokenTampered(t *testing.T) {
	token, err := GenerateToken(testSecret, 1, "user", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tamper with the token by flipping a character
	tampered := []byte(token)
	if tampered[len(tampered)-1] == 'a' {
		tampered[len(tampered)-1] = 'b'
	} else {
		tampered[len(tampered)-1] = 'a'
	}

	_, err = ValidateToken(testSecret, string(tampered))
	if err == nil {
		t.Error("expected error for tampered token, got nil")
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	token, err := GenerateToken(testSecret, 1, "user", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ValidateToken("wrong-secret", token)
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestValidateTokenInvalidString(t *testing.T) {
	_, err := ValidateToken(testSecret, "not-a-jwt-token")
	if err == nil {
		t.Error("expected error for invalid token string, got nil")
	}
}

func TestAuthMiddleware(t *testing.T) {
	handler := AuthMiddleware(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := UserFromContext(r.Context())
		if claims == nil {
			t.Error("expected claims in context")
			http.Error(w, "no claims", http.StatusInternalServerError)
			return
		}
		if claims.UserID != 7 {
			t.Errorf("expected UserID 7, got %d", claims.UserID)
		}
		if claims.Role != "instructor" {
			t.Errorf("expected Role \"instructor\", got %q", claims.Role)
		}
		w.WriteHeader(http.StatusOK)
	}))

	token, err := GenerateToken(testSecret, 7, "instructor", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", rec.Code)
	}
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	handler := AuthMiddleware(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no token, got %d", rec.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	handler := AuthMiddleware(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", rec.Code)
	}
}

func TestUserFromContextNil(t *testing.T) {
	claims := UserFromContext(context.Background())
	if claims != nil {
		t.Error("expected nil claims from empty context")
	}
}
