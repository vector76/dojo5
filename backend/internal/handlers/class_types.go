package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"dojo-crm/backend/internal/models"
)

type ClassTypeHandler struct {
	classTypes *models.ClassTypeRepo
}

func NewClassTypeHandler(classTypes *models.ClassTypeRepo) *ClassTypeHandler {
	return &ClassTypeHandler{classTypes: classTypes}
}

type classTypeResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func toClassTypeResponse(ct *models.ClassType) classTypeResponse {
	return classTypeResponse{
		ID:          ct.ID,
		Name:        ct.Name,
		Description: ct.Description,
	}
}

func (h *ClassTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	types, err := h.classTypes.List()
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]classTypeResponse, len(types))
	for i := range types {
		resp[i] = toClassTypeResponse(&types[i])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *ClassTypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid class type id", http.StatusBadRequest)
		return
	}

	ct, err := h.classTypes.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "class type not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toClassTypeResponse(ct)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

type createClassTypeRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *ClassTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createClassTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}

	ct := &models.ClassType{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.classTypes.Create(ct); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toClassTypeResponse(ct)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

type updateClassTypeRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *ClassTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid class type id", http.StatusBadRequest)
		return
	}

	var req updateClassTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ct, err := h.classTypes.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "class type not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if req.Name != nil {
		if *req.Name == "" {
			writeError(w, "name cannot be empty", http.StatusBadRequest)
			return
		}
		ct.Name = *req.Name
	}
	if req.Description != nil {
		ct.Description = req.Description
	}

	if err := h.classTypes.Update(ct); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toClassTypeResponse(ct)); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *ClassTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, "invalid class type id", http.StatusBadRequest)
		return
	}

	if err := h.classTypes.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, "class type not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
