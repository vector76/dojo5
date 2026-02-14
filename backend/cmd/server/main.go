package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/database"
	"dojo-crm/backend/internal/handlers"
	"dojo-crm/backend/internal/models"
)

func newMux(jwtSecret string, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, classTypeHandler *handlers.ClassTypeHandler, classHandler *handlers.ClassHandler, paymentHandler *handlers.PaymentHandler, balanceHandler *handlers.BalanceHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.Handle("GET /api/auth/me", chain(authHandler.Me, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor", "user")))
	mux.HandleFunc("GET /api/auth/setup-status", authHandler.SetupStatus)
	mux.HandleFunc("POST /api/auth/setup", authHandler.Setup)

	// User endpoints
	mux.Handle("GET /api/users", chain(userHandler.List, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor")))
	mux.Handle("POST /api/users", chain(userHandler.Create, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("GET /api/users/{id}", chain(userHandler.Get, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor", "user")))
	mux.Handle("PUT /api/users/{id}", chain(userHandler.Update, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor", "user")))
	mux.Handle("DELETE /api/users/{id}", chain(userHandler.Delete, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("PUT /api/users/{id}/role", chain(userHandler.ChangeRole, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("PUT /api/users/{id}/password", chain(userHandler.ResetPassword, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))

	// Class type endpoints
	mux.Handle("GET /api/class-types", chain(classTypeHandler.List, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor", "user")))
	mux.Handle("GET /api/class-types/{id}", chain(classTypeHandler.Get, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor", "user")))
	mux.Handle("POST /api/class-types", chain(classTypeHandler.Create, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("PUT /api/class-types/{id}", chain(classTypeHandler.Update, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("DELETE /api/class-types/{id}", chain(classTypeHandler.Delete, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))

	// Class endpoints
	mux.Handle("GET /api/classes", chain(classHandler.List, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor", "user")))
	mux.Handle("GET /api/classes/{id}", chain(classHandler.Get, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "instructor", "user")))
	mux.Handle("POST /api/classes", chain(classHandler.Create, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("PUT /api/classes/{id}", chain(classHandler.Update, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("DELETE /api/classes/{id}", chain(classHandler.Delete, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))

	// Payment endpoints
	mux.Handle("POST /api/payments", chain(paymentHandler.Create, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))
	mux.Handle("GET /api/users/{id}/payments", chain(paymentHandler.ListByUser, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "user")))

	// Balance endpoints
	mux.Handle("GET /api/users/{id}/balance", chain(balanceHandler.Get, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin", "user")))
	mux.Handle("PUT /api/users/{id}/balance", chain(balanceHandler.Set, auth.AuthMiddleware(jwtSecret), auth.RequireRole("admin")))

	return mux
}

// chain wraps an http.HandlerFunc with middleware (applied inside-out).
func chain(h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	var handler http.Handler = h
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
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
	classTypeRepo := models.NewClassTypeRepo(db)
	classRepo := models.NewClassRepo(db)
	paymentRepo := models.NewPaymentRepo(db)
	authHandler := handlers.NewAuthHandler(userRepo, jwtSecret)
	userHandler := handlers.NewUserHandler(userRepo)
	classTypeHandler := handlers.NewClassTypeHandler(classTypeRepo)
	classHandler := handlers.NewClassHandler(classRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentRepo)
	balanceHandler := handlers.NewBalanceHandler(paymentRepo, userRepo)

	fmt.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, newMux(jwtSecret, authHandler, userHandler, classTypeHandler, classHandler, paymentHandler, balanceHandler)); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
