package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/models"
)

type UserHandler struct {
	users *models.UserRepo
}

func NewUserHandler(users *models.UserRepo) *UserHandler {
	return &UserHandler{users: users}
}

var validRoles = map[string]bool{
	"admin":      true,
	"instructor": true,
	"user":       true,
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

type userResponse struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	Phone            string  `json:"phone"`
	Role             string  `json:"role"`
	MembershipType   *string `json:"membership_type,omitempty"`
	MembershipStatus *string `json:"membership_status,omitempty"`
	EmergencyContact *string `json:"emergency_contact,omitempty"`
	JoinDate         *string `json:"join_date,omitempty"`
}

func toUserResponse(u *models.User) userResponse {
	return userResponse{
		ID:               u.ID,
		Name:             u.Name,
		Email:            u.Email,
		Phone:            u.Phone,
		Role:             u.Role,
		MembershipType:   u.MembershipType,
		MembershipStatus: u.MembershipStatus,
		EmergencyContact: u.EmergencyContact,
		JoinDate:         u.JoinDate,
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]userResponse, len(users))
	for i := range users {
		resp[i] = toUserResponse(&users[i])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Phone == "" {
		http.Error(w, "name, email, and phone are required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}
	if !validRoles[req.Role] {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Role:         req.Role,
		PasswordHash: hash,
	}

	if err := h.users.Create(user); err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toUserResponse(user)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
