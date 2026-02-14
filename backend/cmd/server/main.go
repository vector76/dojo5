package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"dojo-crm/backend/internal/database"
	"dojo-crm/backend/internal/handlers"
	"dojo-crm/backend/internal/models"
)

func newMux(authHandler *handlers.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Fprintln(os.Stderr, "JWT_SECRET environment variable is required")
		os.Exit(1)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dojo.db"
	}

	db, err := database.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find executable path: %v\n", err)
		os.Exit(1)
	}
	migrationsDir := filepath.Join(filepath.Dir(exe), "migrations")
	if err := database.Migrate(db, migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	userRepo := models.NewUserRepo(db)
	authHandler := handlers.NewAuthHandler(userRepo, jwtSecret)

	fmt.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, newMux(authHandler)); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
