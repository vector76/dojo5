package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/models"
)

type BalanceHandler struct {
	payments *models.PaymentRepo
	users    *models.UserRepo
}

func NewBalanceHandler(payments *models.PaymentRepo, users *models.UserRepo) *BalanceHandler {
	return &BalanceHandler{payments: payments, users: users}
}

type balanceResponse struct {
	UserID          int64   `json:"user_id"`
	ExpectedBalance float64 `json:"expected_balance"`
	TotalPaid       float64 `json:"total_paid"`
	Balance         float64 `json:"balance"`
}

func (h *BalanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Same permissions as payment history: admin or self
	if claims.Role != "admin" && claims.UserID != userID {
		writeError(w, "forbidden", http.StatusForbidden)
		return
	}

	user, err := h.users.GetByID(userID)
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

	balance, err := h.payments.GetBalance(userID)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	totalPaid := user.ExpectedBalance - balance

	resp := balanceResponse{
		UserID:          userID,
		ExpectedBalance: user.ExpectedBalance,
		TotalPaid:       totalPaid,
		Balance:         balance,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

type setBalanceRequest struct {
	ExpectedBalance *float64 `json:"expected_balance"`
}

func (h *BalanceHandler) Set(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req setBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ExpectedBalance == nil {
		writeError(w, "expected_balance is required", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByID(userID)
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

	user.ExpectedBalance = *req.ExpectedBalance
	if err := h.users.Update(user); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Return the updated balance info
	balance, err := h.payments.GetBalance(userID)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	totalPaid := user.ExpectedBalance - balance

	resp := balanceResponse{
		UserID:          userID,
		ExpectedBalance: user.ExpectedBalance,
		TotalPaid:       totalPaid,
		Balance:         balance,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}
