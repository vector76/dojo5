package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]userResponse, len(users))
	for i := range users {
		resp[i] = toUserResponse(&users[i])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Phone == "" {
		writeError(w, "name, email, and phone are required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		writeError(w, "password is required", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}
	if !validRoles[req.Role] {
		writeError(w, "invalid role", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
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
			writeError(w, "email already exists", http.StatusConflict)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toUserResponse(user)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin, instructor, or the user themselves can view
	if claims.Role != "admin" && claims.Role != "instructor" && claims.UserID != id {
		writeError(w, "forbidden", http.StatusForbidden)
		return
	}

	user, err := h.users.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user.DeletedAt != nil {
		writeError(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toUserResponse(user)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

type updateUserRequest struct {
	Name             *string `json:"name"`
	Email            *string `json:"email"`
	Phone            *string `json:"phone"`
	Role             *string `json:"role"`
	Password         *string `json:"password"`
	MembershipType   *string `json:"membership_type"`
	MembershipStatus *string `json:"membership_status"`
	EmergencyContact *string `json:"emergency_contact"`
	JoinDate         *string `json:"join_date"`
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	isSelf := claims.UserID == id
	isAdmin := claims.Role == "admin"

	if !isAdmin && !isSelf {
		writeError(w, "forbidden", http.StatusForbidden)
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Non-admin users cannot change role or membership fields
	if !isAdmin {
		if req.Role != nil {
			writeError(w, "only admins can change role", http.StatusForbidden)
			return
		}
		if req.MembershipType != nil || req.MembershipStatus != nil || req.JoinDate != nil {
			writeError(w, "only admins can change membership fields", http.StatusForbidden)
			return
		}
	}

	if req.Role != nil && !validRoles[*req.Role] {
		writeError(w, "invalid role", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user.DeletedAt != nil {
		writeError(w, "user not found", http.StatusNotFound)
		return
	}

	// Apply partial updates
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Password != nil {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		user.PasswordHash = hash
	}
	if req.MembershipType != nil {
		user.MembershipType = req.MembershipType
	}
	if req.MembershipStatus != nil {
		user.MembershipStatus = req.MembershipStatus
	}
	if req.EmergencyContact != nil {
		user.EmergencyContact = req.EmergencyContact
	}
	if req.JoinDate != nil {
		user.JoinDate = req.JoinDate
	}

	if err := h.users.Update(user); err != nil {
		if isUniqueViolation(err) {
			writeError(w, "email already exists", http.StatusConflict)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toUserResponse(user)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.users.SoftDelete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type changeRoleRequest struct {
	Role string `json:"role"`
}

func (h *UserHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req changeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		writeError(w, "role is required", http.StatusBadRequest)
		return
	}
	if !validRoles[req.Role] {
		writeError(w, "invalid role", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user.DeletedAt != nil {
		writeError(w, "user not found", http.StatusNotFound)
		return
	}

	// Prevent demoting the last admin
	if user.Role == "admin" && req.Role != "admin" {
		count, err := h.users.CountByRole("admin")
		if err != nil {
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		if count <= 1 {
			writeError(w, "cannot demote the last admin", http.StatusConflict)
			return
		}
	}

	user.Role = req.Role
	if err := h.users.Update(user); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toUserResponse(user)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		writeError(w, "password is required", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user.DeletedAt != nil {
		writeError(w, "user not found", http.StatusNotFound)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = hash
	if err := h.users.Update(user); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
