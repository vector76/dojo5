package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dojo-crm/backend/internal/handlers"
	"dojo-crm/backend/internal/models"

	"dojo-crm/backend/internal/database"
	"path/filepath"
)

func TestHealthEndpoint(t *testing.T) {
	dir := t.TempDir()
	db, err := database.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()
	authHandler := handlers.NewAuthHandler(models.NewUserRepo(db), "test-secret")
	handler := newMux(authHandler)
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
