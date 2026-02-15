package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dojo-crm/backend/internal/database"
	"dojo-crm/backend/internal/handlers"
	"dojo-crm/backend/internal/models"
	"path/filepath"
)

func TestHealthEndpoint(t *testing.T) {
	dir := t.TempDir()
	db, err := database.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()
	userRepo := models.NewUserRepo(db)
	classTypeRepo := models.NewClassTypeRepo(db)
	classRepo := models.NewClassRepo(db)
	paymentRepo := models.NewPaymentRepo(db)
	authHandler := handlers.NewAuthHandler(userRepo, "test-secret")
	userHandler := handlers.NewUserHandler(userRepo)
	classTypeHandler := handlers.NewClassTypeHandler(classTypeRepo)
	classHandler := handlers.NewClassHandler(classRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentRepo)
	balanceHandler := handlers.NewBalanceHandler(paymentRepo, userRepo)
	attendanceRepo := models.NewAttendanceRepo(db)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo)
	handler := newMux("test-secret", authHandler, userHandler, classTypeHandler, classHandler, paymentHandler, balanceHandler, attendanceHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status field to be \"ok\", got %q", body["status"])
	}
}

func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	dir := t.TempDir()
	db, err := database.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	userRepo := models.NewUserRepo(db)
	classTypeRepo := models.NewClassTypeRepo(db)
	classRepo := models.NewClassRepo(db)
	paymentRepo := models.NewPaymentRepo(db)
	authHandler := handlers.NewAuthHandler(userRepo, "test-secret")
	userHandler := handlers.NewUserHandler(userRepo)
	classTypeHandler := handlers.NewClassTypeHandler(classTypeRepo)
	classHandler := handlers.NewClassHandler(classRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentRepo)
	balanceHandler := handlers.NewBalanceHandler(paymentRepo, userRepo)
	attendanceRepo := models.NewAttendanceRepo(db)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo)
	return newMux("test-secret", authHandler, userHandler, classTypeHandler, classHandler, paymentHandler, balanceHandler, attendanceHandler)
}

func TestFrontendServedAtRoot(t *testing.T) {
	mux := newTestMux(t)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<html") && !strings.Contains(string(body), "<!DOCTYPE") && !strings.Contains(string(body), "<!doctype") {
		t.Errorf("expected HTML content at root, got: %s", string(body)[:min(len(body), 200)])
	}
}

func TestSPAFallbackDoesNotOverrideAPI(t *testing.T) {
	mux := newTestMux(t)
	server := httptest.NewServer(mux)
	defer server.Close()

	// API route should still work
	resp, err := http.Get(server.URL + "/api/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for API route, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestSPAFallbackForClientRoutes(t *testing.T) {
	mux := newTestMux(t)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Client-side route should serve index.html
	resp, err := http.Get(server.URL + "/members")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for SPA route, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<html") && !strings.Contains(string(body), "<!DOCTYPE") && !strings.Contains(string(body), "<!doctype") {
		t.Errorf("expected HTML fallback for SPA route, got: %s", string(body)[:min(len(body), 200)])
	}
}
