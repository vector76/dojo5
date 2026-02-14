package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dojo-crm/backend/internal/models"
)

func createTestClass(t *testing.T, classRepo *models.ClassRepo, classTypeID, instructorID int64, startTime time.Time, duration, capacity int) *models.Class {
	t.Helper()
	c := &models.Class{
		ClassTypeID:     classTypeID,
		InstructorID:    instructorID,
		StartTime:       startTime,
		DurationMinutes: duration,
		Capacity:        capacity,
	}
	if err := classRepo.Create(c); err != nil {
		t.Fatalf("failed to create class: %v", err)
	}
	return c
}

func TestClassHandler_List(t *testing.T) {
	db := setupTestDB(t)
	userRepo := models.NewUserRepo(db)
	classTypeRepo := models.NewClassTypeRepo(db)
	classRepo := models.NewClassRepo(db)
	handler := NewClassHandler(classRepo)

	instructor := createTestUser(t, userRepo, "cls-inst@example.com", "password", "instructor")
	ct := createTestClassType(t, classTypeRepo, "Yoga", nil)

	t1 := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	createTestClass(t, classRepo, ct.ID, instructor.ID, t1, 60, 20)
	createTestClass(t, classRepo, ct.ID, instructor.ID, t2, 90, 15)

	wrapped := wrapWithMiddleware(handler.List, "admin", "instructor", "user")

	t.Run("admin can list all", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes", nil, 1, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 classes, got %d", len(resp))
		}
	})

	t.Run("instructor can list", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes", nil, 1, "instructor")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("user can list", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes", nil, 1, "user")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("filter by date from", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes?from=2026-03-02T00:00:00Z", nil, 1, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 1 {
			t.Errorf("expected 1 class after date filter, got %d", len(resp))
		}
	})

	t.Run("filter by date to", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes?to=2026-03-02T00:00:00Z", nil, 1, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 1 {
			t.Errorf("expected 1 class before date filter, got %d", len(resp))
		}
	})

	t.Run("filter by date range", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes?from=2026-03-01T00:00:00Z&to=2026-03-03T00:00:00Z", nil, 1, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 classes in date range, got %d", len(resp))
		}
	})

	t.Run("invalid from date", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes?from=bad-date", nil, 1, "admin")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/classes", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestClassHandler_Get(t *testing.T) {
	db := setupTestDB(t)
	userRepo := models.NewUserRepo(db)
	classTypeRepo := models.NewClassTypeRepo(db)
	classRepo := models.NewClassRepo(db)
	handler := NewClassHandler(classRepo)

	instructor := createTestUser(t, userRepo, "cls-get-inst@example.com", "password", "instructor")
	ct := createTestClassType(t, classTypeRepo, "Karate", nil)
	cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

	wrapped := wrapWithMiddleware(handler.Get, "admin", "instructor", "user")

	t.Run("admin can get", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes/"+itoa(cls.ID), nil, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ID != cls.ID {
			t.Errorf("expected id %d, got %d", cls.ID, resp.ID)
		}
		if resp.ClassTypeID != ct.ID {
			t.Errorf("expected class_type_id %d, got %d", ct.ID, resp.ClassTypeID)
		}
		if resp.InstructorID != instructor.ID {
			t.Errorf("expected instructor_id %d, got %d", instructor.ID, resp.InstructorID)
		}
		if resp.DurationMinutes != 60 {
			t.Errorf("expected duration 60, got %d", resp.DurationMinutes)
		}
		if resp.Capacity != 20 {
			t.Errorf("expected capacity 20, got %d", resp.Capacity)
		}
	})

	t.Run("user can get", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes/"+itoa(cls.ID), nil, 1, "user")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("nonexistent returns 404", func(t *testing.T) {
		req, _ := authedRequest(t, http.MethodGet, "/api/classes/9999", nil, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/classes/1", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestClassHandler_Create(t *testing.T) {
	setup := func(t *testing.T) (http.Handler, *models.ClassRepo, *models.ClassType, *models.User) {
		t.Helper()
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-create-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		return wrapWithMiddleware(handler.Create, "admin"), classRepo, ct, instructor
	}

	t.Run("admin creates class", func(t *testing.T) {
		h, _, ct, inst := setup(t)

		body, _ := json.Marshal(map[string]any{
			"class_type_id":    ct.ID,
			"instructor_id":   inst.ID,
			"start_time":      "2026-03-01T10:00:00Z",
			"duration_minutes": 60,
			"capacity":        20,
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes", body, 1, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if resp.ClassTypeID != ct.ID {
			t.Errorf("expected class_type_id %d, got %d", ct.ID, resp.ClassTypeID)
		}
		if resp.DurationMinutes != 60 {
			t.Errorf("expected duration 60, got %d", resp.DurationMinutes)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		h, _, _, _ := setup(t)

		body, _ := json.Marshal(map[string]any{"class_type_id": 1})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes", body, 1, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("invalid start_time", func(t *testing.T) {
		h, _, ct, inst := setup(t)

		body, _ := json.Marshal(map[string]any{
			"class_type_id":    ct.ID,
			"instructor_id":   inst.ID,
			"start_time":      "not-a-date",
			"duration_minutes": 60,
			"capacity":        20,
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes", body, 1, "admin")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot create", func(t *testing.T) {
		h, _, ct, inst := setup(t)

		body, _ := json.Marshal(map[string]any{
			"class_type_id":    ct.ID,
			"instructor_id":   inst.ID,
			"start_time":      "2026-03-01T10:00:00Z",
			"duration_minutes": 60,
			"capacity":        20,
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes", body, 1, "instructor")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("user cannot create", func(t *testing.T) {
		h, _, ct, inst := setup(t)

		body, _ := json.Marshal(map[string]any{
			"class_type_id":    ct.ID,
			"instructor_id":   inst.ID,
			"start_time":      "2026-03-01T10:00:00Z",
			"duration_minutes": 60,
			"capacity":        20,
		})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes", body, 1, "user")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		h, _, _, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/classes", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestClassHandler_Update(t *testing.T) {
	t.Run("admin updates class", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-upd-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]any{"duration_minutes": 90, "capacity": 25})
		req, _ := authedRequest(t, http.MethodPut, "/api/classes/"+itoa(cls.ID), body, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.DurationMinutes != 90 {
			t.Errorf("expected duration 90, got %d", resp.DurationMinutes)
		}
		if resp.Capacity != 25 {
			t.Errorf("expected capacity 25, got %d", resp.Capacity)
		}
		// Verify unchanged fields preserved
		if resp.ClassTypeID != ct.ID {
			t.Errorf("expected class_type_id preserved as %d, got %d", ct.ID, resp.ClassTypeID)
		}
	})

	t.Run("update start_time", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-upd2-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]any{"start_time": "2026-04-01T14:00:00Z"})
		req, _ := authedRequest(t, http.MethodPut, "/api/classes/"+itoa(cls.ID), body, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp classResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.StartTime != "2026-04-01T14:00:00Z" {
			t.Errorf("expected updated start_time, got %q", resp.StartTime)
		}
	})

	t.Run("invalid duration rejected", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-upd3-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]any{"duration_minutes": 0})
		req, _ := authedRequest(t, http.MethodPut, "/api/classes/"+itoa(cls.ID), body, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("nonexistent returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]any{"capacity": 30})
		req, _ := authedRequest(t, http.MethodPut, "/api/classes/9999", body, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot update", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-upd4-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.Update, "admin")

		body, _ := json.Marshal(map[string]any{"capacity": 30})
		req, _ := authedRequest(t, http.MethodPut, "/api/classes/"+itoa(cls.ID), body, 1, "instructor")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})
}

func TestClassHandler_Delete(t *testing.T) {
	t.Run("admin deletes class", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-del-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/classes/"+itoa(cls.ID), nil, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify deleted
		_, err := classRepo.GetByID(cls.ID)
		if err == nil {
			t.Error("expected error after delete")
		}
	})

	t.Run("nonexistent returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/classes/9999", nil, 1, "admin")
		req.SetPathValue("id", "9999")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("instructor cannot delete", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-del2-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/classes/"+itoa(cls.ID), nil, 1, "instructor")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("user cannot delete", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)
		instructor := createTestUser(t, userRepo, "cls-del3-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req, _ := authedRequest(t, http.MethodDelete, "/api/classes/"+itoa(cls.ID), nil, 1, "user")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		db := setupTestDB(t)
		classRepo := models.NewClassRepo(db)
		handler := NewClassHandler(classRepo)

		wrapped := wrapWithMiddleware(handler.Delete, "admin")

		req := httptest.NewRequest(http.MethodDelete, "/api/classes/1", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}
