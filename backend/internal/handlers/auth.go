package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/models"
)

type AuthHandler struct {
	users     *models.UserRepo
	jwtSecret string
}

func NewAuthHandler(users *models.UserRepo, jwtSecret string) *AuthHandler {
	return &AuthHandler{users: users, jwtSecret: jwtSecret}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Reject soft-deleted users
	if user.DeletedAt != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Role, 24*time.Hour)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(loginResponse{Token: token}); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
