package frontend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_ServesIndexHTML(t *testing.T) {
	handler := Handler()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<html") && !strings.Contains(body, "<!DOCTYPE") && !strings.Contains(body, "<!doctype") {
		t.Errorf("expected HTML content, got: %s", body[:min(len(body), 200)])
	}
}

func TestHandler_ServesStaticAssets(t *testing.T) {
	handler := Handler()
	req := httptest.NewRequest("GET", "/vite.svg", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for static asset, got %d", rec.Code)
	}
}

func TestHandler_SPAFallback(t *testing.T) {
	handler := Handler()

	// SPA routes should return index.html
	routes := []string{"/login", "/members", "/classes", "/some/deep/route"}
	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("route %s: expected 200, got %d", route, rec.Code)
			continue
		}

		body := rec.Body.String()
		if !strings.Contains(body, "<html") && !strings.Contains(body, "<!DOCTYPE") && !strings.Contains(body, "<!doctype") {
			t.Errorf("route %s: expected HTML fallback, got: %s", route, body[:min(len(body), 200)])
		}
	}
}

func TestHandler_DoesNotHandleAPIRoutes(t *testing.T) {
	handler := Handler()
	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for API route, got %d", rec.Code)
	}
}
