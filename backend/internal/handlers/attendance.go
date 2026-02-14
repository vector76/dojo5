package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"dojo-crm/backend/internal/auth"
	"dojo-crm/backend/internal/models"
)

type AttendanceHandler struct {
	attendance *models.AttendanceRepo
}

func NewAttendanceHandler(attendance *models.AttendanceRepo) *AttendanceHandler {
	return &AttendanceHandler{attendance: attendance}
}

type attendanceResponse struct {
	ID          int64  `json:"id"`
	ClassID     int64  `json:"class_id"`
	UserID      int64  `json:"user_id"`
	CheckedInAt string `json:"checked_in_at"`
}

func toAttendanceResponse(a *models.Attendance) attendanceResponse {
	return attendanceResponse{
		ID:          a.ID,
		ClassID:     a.ClassID,
		UserID:      a.UserID,
		CheckedInAt: a.CheckedInAt.Format(time.RFC3339),
	}
}

type recordAttendanceRequest struct {
	UserID int64 `json:"user_id"`
}

func (h *AttendanceHandler) Record(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}

	var req recordAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	a := &models.Attendance{
		ClassID: classID,
		UserID:  req.UserID,
	}

	if err := h.attendance.RecordAttendance(a); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toAttendanceResponse(a)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *AttendanceHandler) ListByClass(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}

	records, err := h.attendance.ListByClass(classID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]attendanceResponse, len(records))
	for i := range records {
		resp[i] = toAttendanceResponse(&records[i])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *AttendanceHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	claims := auth.UserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Admin and instructor can view any user; users can only view their own
	if claims.Role != "admin" && claims.Role != "instructor" && claims.UserID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	records, err := h.attendance.ListByUser(userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]attendanceResponse, len(records))
	for i := range records {
		resp[i] = toAttendanceResponse(&records[i])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
