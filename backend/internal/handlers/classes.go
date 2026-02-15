package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"dojo-crm/backend/internal/models"
)

type ClassHandler struct {
	classes *models.ClassRepo
}

func NewClassHandler(classes *models.ClassRepo) *ClassHandler {
	return &ClassHandler{classes: classes}
}

type classResponse struct {
	ID              int64  `json:"id"`
	ClassTypeID     int64  `json:"class_type_id"`
	InstructorID    int64  `json:"instructor_id"`
	StartTime       string `json:"start_time"`
	DurationMinutes int    `json:"duration_minutes"`
	Capacity        int    `json:"capacity"`
}

func toClassResponse(c *models.Class) classResponse {
	return classResponse{
		ID:              c.ID,
		ClassTypeID:     c.ClassTypeID,
		InstructorID:    c.InstructorID,
		StartTime:       c.StartTime.Format(time.RFC3339),
		DurationMinutes: c.DurationMinutes,
		Capacity:        c.Capacity,
	}
}

func (h *ClassHandler) List(w http.ResponseWriter, r *http.Request) {
	var filter models.ClassFilter

	if v := r.URL.Query().Get("class_type_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, "invalid class_type_id", http.StatusBadRequest)
			return
		}
		filter.ClassTypeID = &id
	}
	if v := r.URL.Query().Get("instructor_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, "invalid instructor_id", http.StatusBadRequest)
			return
		}
		filter.InstructorID = &id
	}
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, "invalid from date (use RFC3339 format)", http.StatusBadRequest)
			return
		}
		filter.From = &t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, "invalid to date (use RFC3339 format)", http.StatusBadRequest)
			return
		}
		filter.To = &t
	}

	classes, err := h.classes.List(filter)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]classResponse, len(classes))
	for i := range classes {
		resp[i] = toClassResponse(&classes[i])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *ClassHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid class id", http.StatusBadRequest)
		return
	}

	c, err := h.classes.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "class not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toClassResponse(c)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

type createClassRequest struct {
	ClassTypeID     int64  `json:"class_type_id"`
	InstructorID    int64  `json:"instructor_id"`
	StartTime       string `json:"start_time"`
	DurationMinutes int    `json:"duration_minutes"`
	Capacity        int    `json:"capacity"`
}

func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ClassTypeID == 0 || req.InstructorID == 0 || req.StartTime == "" || req.DurationMinutes <= 0 || req.Capacity <= 0 {
		writeError(w, "class_type_id, instructor_id, start_time, duration_minutes, and capacity are required", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, "invalid start_time (use RFC3339 format)", http.StatusBadRequest)
		return
	}

	c := &models.Class{
		ClassTypeID:     req.ClassTypeID,
		InstructorID:    req.InstructorID,
		StartTime:       startTime,
		DurationMinutes: req.DurationMinutes,
		Capacity:        req.Capacity,
	}

	if err := h.classes.Create(c); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toClassResponse(c)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

type updateClassRequest struct {
	ClassTypeID     *int64  `json:"class_type_id"`
	InstructorID    *int64  `json:"instructor_id"`
	StartTime       *string `json:"start_time"`
	DurationMinutes *int    `json:"duration_minutes"`
	Capacity        *int    `json:"capacity"`
}

func (h *ClassHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid class id", http.StatusBadRequest)
		return
	}

	var req updateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, err := h.classes.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "class not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if req.ClassTypeID != nil {
		if *req.ClassTypeID == 0 {
			writeError(w, "class_type_id cannot be zero", http.StatusBadRequest)
			return
		}
		c.ClassTypeID = *req.ClassTypeID
	}
	if req.InstructorID != nil {
		if *req.InstructorID == 0 {
			writeError(w, "instructor_id cannot be zero", http.StatusBadRequest)
			return
		}
		c.InstructorID = *req.InstructorID
	}
	if req.StartTime != nil {
		t, err := time.Parse(time.RFC3339, *req.StartTime)
		if err != nil {
			writeError(w, "invalid start_time (use RFC3339 format)", http.StatusBadRequest)
			return
		}
		c.StartTime = t
	}
	if req.DurationMinutes != nil {
		if *req.DurationMinutes <= 0 {
			writeError(w, "duration_minutes must be positive", http.StatusBadRequest)
			return
		}
		c.DurationMinutes = *req.DurationMinutes
	}
	if req.Capacity != nil {
		if *req.Capacity <= 0 {
			writeError(w, "capacity must be positive", http.StatusBadRequest)
			return
		}
		c.Capacity = *req.Capacity
	}

	if err := h.classes.Update(c); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toClassResponse(c)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid class id", http.StatusBadRequest)
		return
	}

	if err := h.classes.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "class not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
