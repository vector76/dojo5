package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/models"
)

type PaymentHandler struct {
	payments *models.PaymentRepo
}

func NewPaymentHandler(payments *models.PaymentRepo) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

type paymentResponse struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"user_id"`
	Amount     float64 `json:"amount"`
	Date       string  `json:"date"`
	Note       *string `json:"note,omitempty"`
	RecordedBy int64   `json:"recorded_by"`
}

func toPaymentResponse(p *models.Payment) paymentResponse {
	return paymentResponse{
		ID:         p.ID,
		UserID:     p.UserID,
		Amount:     p.Amount,
		Date:       p.Date,
		Note:       p.Note,
		RecordedBy: p.RecordedBy,
	}
}

type createPaymentRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
	Note   *string `json:"note"`
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.Amount == 0 || req.Date == "" {
		writeError(w, "user_id, amount, and date are required", http.StatusBadRequest)
		return
	}

	p := &models.Payment{
		UserID:     req.UserID,
		Amount:     req.Amount,
		Date:       req.Date,
		Note:       req.Note,
		RecordedBy: claims.UserID,
	}

	if err := h.payments.RecordPayment(p); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toPaymentResponse(p)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *PaymentHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
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

	// Admin can view any user's payments; users can only view their own
	if claims.Role != "admin" && claims.UserID != userID {
		writeError(w, "forbidden", http.StatusForbidden)
		return
	}

	payments, err := h.payments.ListByUser(userID)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]paymentResponse, len(payments))
	for i := range payments {
		resp[i] = toPaymentResponse(&payments[i])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}
