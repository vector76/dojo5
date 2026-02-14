package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dojo-crm/backend/internal/models"
)

func createTestClassType(t *testing.T, repo *models.ClassTypeRepo, name string, description *string) *models.ClassType {
	t.Helper()
	ct := &models.ClassType{Name: name, Description: description}
	if err := repo.Create(ct); err != nil {
		t.Fatalf("failed to create class type: %v", err)
	}
	return ct
}

func TestClassTypeHandler_List(t *testing.T) {
	db := setupTestDB(t)
	repo := models.NewClassTypeRepo(db)
	handler := NewClassTypeHandler(repo)

	desc := "A beginner class"
	createTestClassType(t, repo, "Yoga", nil)
	createTestClassType(t, repo, "Karate", &desc)

	wrapped := wrapWithMiddleware(handler.List, "admin", "instructor", "user")

	t.Run("admin can list", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/class-types", nil, 1, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []classTypeResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 class types, got %d", len(resp))
		}
	})

	t.Run("instructor can list", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/class-types", nil, 1, "instructor")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("user can list", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/class-types", nil, 1, "user")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/class-types", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestClassTypeHandler_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := models.NewClassTypeRepo(db)
	handler := NewClassTypeHandler(repo)

	desc := "Martial arts"
	ct := createTestClassType(t, repo, "Karate", &desc)

	wrapped := wrapWithMiddleware(handler.Get, "admin", "instructor", "user")

	t.Run("admin can get", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/class-types/"+itoa(ct.ID), nil, 1, "admin")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp classTypeResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ID != ct.ID {
			t.Errorf("expected id %d, got %d", ct.ID, resp.ID)
		}
		if resp.Name != "Karate" {
			t.Errorf("expected name 'Karate', got %q", resp.Name)
		}
		if resp.Description == nil || *resp.Description != "Martial arts" {
			t.Errorf("expected description 'Martial arts', got %v", resp.Description)
		}
	})

	t.Run("user can get", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/class-types/"+itoa(ct.ID), nil, 1, "user")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("nonexistent returns 404", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/class-types/9999", nil, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/class-types/1", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestClassTypeHandler_Create(t *testing.T) {
	wrapped := func(t *testing.T) (http.Handler, *models.ClassTypeRepo) {
		t.Helper()
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		return wrapWithMiddleware(handler.Create, "admin"), repo
	}

	t.Run("admin creates class type", func(t *testing.T) {
		h, _ := wrapped(t)

		body, _ := json.Marshal(map[string]string{"name": "Yoga", "description": "Relaxing yoga class"})
		req, _ := authedRequest(t, http.MethodPost, "/api/class-types", body, 1, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp classTypeResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if resp.Name != "Yoga" {
			t.Errorf("expected name 'Yoga', got %q", resp.Name)
		}
		if resp.Description == nil || *resp.Description != "Relaxing yoga class" {
			t.Errorf("expected description, got %v", resp.Description)
		}
	})

	t.Run("admin creates without description", func(t *testing.T) {
		h, _ := wrapped(t)

		body, _ := json.Marshal(map[string]string{"name": "Pilates"})
		req, _ := authedRequest(t, http.MethodPost, "/api/class-types", body, 1, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp classTypeResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Description != nil {
			t.Errorf("expected nil description, got %v", resp.Description)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		h, _ := wrapped(t)

		body, _ := json.Marshal(map[string]string{"description": "some desc"})
		req, _ := authedRequest(t, http.MethodPost, "/api/class-types", body, 1, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot create", func(t *testing.T) {
		h, _ := wrapped(t)

		body, _ := json.Marshal(map[string]string{"name": "Yoga"})
		req, _ := authedRequest(t, http.MethodPost, "/api/class-types", body, 1, "instructor")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("user cannot create", func(t *testing.T) {
		h, _ := wrapped(t)

		body, _ := json.Marshal(map[string]string{"name": "Yoga"})
		req, _ := authedRequest(t, http.MethodPost, "/api/class-types", body, 1, "user")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		h, _ := wrapped(t)

		req := httptest.NewRequest(http.MethodPost, "/api/class-types", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestClassTypeHandler_Update(t *testing.T) {
	t.Run("admin updates class type", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		ct := createTestClassType(t, repo, "Yoga", nil)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]string{"name": "Hot Yoga", "description": "Heated yoga"})
		req, _ := authedRequest(t, http.MethodPut, "/api/class-types/"+itoa(ct.ID), body, 1, "admin")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp classTypeResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Name != "Hot Yoga" {
			t.Errorf("expected name 'Hot Yoga', got %q", resp.Name)
		}
		if resp.Description == nil || *resp.Description != "Heated yoga" {
			t.Errorf("expected description 'Heated yoga', got %v", resp.Description)
		}
	})

	t.Run("partial update name only", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		desc := "Original desc"
		ct := createTestClassType(t, repo, "Yoga", &desc)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]string{"name": "New Yoga"})
		req, _ := authedRequest(t, http.MethodPut, "/api/class-types/"+itoa(ct.ID), body, 1, "admin")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp classTypeResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Name != "New Yoga" {
			t.Errorf("expected 'New Yoga', got %q", resp.Name)
		}
		// Description should be preserved
		if resp.Description == nil || *resp.Description != "Original desc" {
			t.Errorf("expected description preserved, got %v", resp.Description)
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		ct := createTestClassType(t, repo, "Yoga", nil)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]string{"name": ""})
		req, _ := authedRequest(t, http.MethodPut, "/api/class-types/"+itoa(ct.ID), body, 1, "admin")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("nonexistent returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]string{"name": "Ghost"})
		req, _ := authedRequest(t, http.MethodPut, "/api/class-types/9999", body, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot update", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		ct := createTestClassType(t, repo, "Yoga", nil)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]string{"name": "Hacked"})
		req, _ := authedRequest(t, http.MethodPut, "/api/class-types/"+itoa(ct.ID), body, 1, "instructor")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})
}

func TestClassTypeHandler_Delete(t *testing.T) {
	t.Run("admin deletes class type", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		ct := createTestClassType(t, repo, "Yoga", nil)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/class-types/"+itoa(ct.ID), nil, 1, "admin")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify deleted
		_, err := repo.GetByID(ct.ID)
		if err == nil {
			t.Error("expected error after delete")
		}
	})

	t.Run("nonexistent returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/class-types/9999", nil, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot delete", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		ct := createTestClassType(t, repo, "Yoga", nil)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/class-types/"+itoa(ct.ID), nil, 1, "instructor")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("user cannot delete", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)
		ct := createTestClassType(t, repo, "Yoga", nil)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/class-types/"+itoa(ct.ID), nil, 1, "user")
		req.SetPathValue("id", itoa(ct.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		db := setupTestDB(t)
		repo := models.NewClassTypeRepo(db)
		handler := NewClassTypeHandler(repo)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req := httptest.NewRequest(http.MethodDelete, "/api/class-types/1", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}
