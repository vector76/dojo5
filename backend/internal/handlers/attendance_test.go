package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dojo-crm/backend/internal/models"
)

func TestAttendanceHandler_Record(t *testing.T) {
	setup := func(t *testing.T) (http.Handler, *models.AttendanceRepo, *models.Class, *models.User) {
		t.Helper()
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
		student := createTestUser(t, userRepo, "att-student@example.com", "password", "user")

		return wrapWithMiddleware(handler.Record, "admin", "instructor"), attendanceRepo, cls, student
	}

	t.Run("admin records attendance", func(t *testing.T) {
		h, _, cls, student := setup(t)

		body, _ := json.Marshal(map[string]any{"user_id": student.ID})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes/"+itoa(cls.ID)+"/attendance", body, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp attendanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if resp.ClassID != cls.ID {
			t.Errorf("expected class_id %d, got %d", cls.ID, resp.ClassID)
		}
		if resp.UserID != student.ID {
			t.Errorf("expected user_id %d, got %d", student.ID, resp.UserID)
		}
		if resp.CheckedInAt == "" {
			t.Error("expected non-empty checked_in_at")
		}
		parsed, err := time.Parse(time.RFC3339, resp.CheckedInAt)
		if err != nil {
			t.Fatalf("checked_in_at not valid RFC3339: %v", err)
		}
		if time.Since(parsed) > 10*time.Second {
			t.Errorf("checked_in_at too old: %s", resp.CheckedInAt)
		}
	})

	t.Run("instructor records attendance", func(t *testing.T) {
		h, _, cls, student := setup(t)

		body, _ := json.Marshal(map[string]any{"user_id": student.ID})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes/"+itoa(cls.ID)+"/attendance", body, 1, "instructor")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing user_id", func(t *testing.T) {
		h, _, cls, _ := setup(t)

		body, _ := json.Marshal(map[string]any{})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes/"+itoa(cls.ID)+"/attendance", body, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("user cannot record", func(t *testing.T) {
		h, _, cls, student := setup(t)

		body, _ := json.Marshal(map[string]any{"user_id": student.ID})
		req, _ := authedRequest(t, http.MethodPost, "/api/classes/"+itoa(cls.ID)+"/attendance", body, 1, "user")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		h, _, cls, _ := setup(t)

		req := httptest.NewRequest(http.MethodPost, "/api/classes/"+itoa(cls.ID)+"/attendance", nil)
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestAttendanceHandler_ListByClass(t *testing.T) {
	t.Run("admin can list class attendance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-lc-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
		s1 := createTestUser(t, userRepo, "att-lc-s1@example.com", "password", "user")
		s2 := createTestUser(t, userRepo, "att-lc-s2@example.com", "password", "user")

		attendanceRepo.RecordAttendance(&models.Attendance{ClassID: cls.ID, UserID: s1.ID})
		attendanceRepo.RecordAttendance(&models.Attendance{ClassID: cls.ID, UserID: s2.ID})

		wrapped := wrapWithMiddleware(handler.ListByClass, "admin", "instructor")

		req, _ := authedRequest(t, http.MethodGet, "/api/classes/"+itoa(cls.ID)+"/attendance", nil, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp []attendanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 records, got %d", len(resp))
		}
	})

	t.Run("instructor can list class attendance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-lc2-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.ListByClass, "admin", "instructor")

		req, _ := authedRequest(t, http.MethodGet, "/api/classes/"+itoa(cls.ID)+"/attendance", nil, 1, "instructor")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("user cannot list class attendance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-lc3-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.ListByClass, "admin", "instructor")

		req, _ := authedRequest(t, http.MethodGet, "/api/classes/"+itoa(cls.ID)+"/attendance", nil, 1, "user")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("empty class returns empty list", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-lc4-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)

		wrapped := wrapWithMiddleware(handler.ListByClass, "admin", "instructor")

		req, _ := authedRequest(t, http.MethodGet, "/api/classes/"+itoa(cls.ID)+"/attendance", nil, 1, "admin")
		req.SetPathValue("id", itoa(cls.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []attendanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 0 {
			t.Errorf("expected 0 records, got %d", len(resp))
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		db := setupTestDB(t)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		wrapped := wrapWithMiddleware(handler.ListByClass, "admin", "instructor")

		req := httptest.NewRequest(http.MethodGet, "/api/classes/1/attendance", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestAttendanceHandler_ListByUser(t *testing.T) {
	t.Run("admin can view any user attendance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-lu-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls1 := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
		cls2 := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC), 60, 20)
		student := createTestUser(t, userRepo, "att-lu-student@example.com", "password", "user")

		attendanceRepo.RecordAttendance(&models.Attendance{ClassID: cls1.ID, UserID: student.ID})
		attendanceRepo.RecordAttendance(&models.Attendance{ClassID: cls2.ID, UserID: student.ID})

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "instructor", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(student.ID)+"/attendance", nil, 1, "admin")
		req.SetPathValue("id", itoa(student.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp []attendanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 records, got %d", len(resp))
		}
	})

	t.Run("instructor can view any user attendance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-lu2-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
		student := createTestUser(t, userRepo, "att-lu2-student@example.com", "password", "user")

		attendanceRepo.RecordAttendance(&models.Attendance{ClassID: cls.ID, UserID: student.ID})

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "instructor", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(student.ID)+"/attendance", nil, instructor.ID, "instructor")
		req.SetPathValue("id", itoa(student.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp []attendanceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp) != 1 {
			t.Errorf("expected 1 record, got %d", len(resp))
		}
	})

	t.Run("user can view own attendance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		classTypeRepo := models.NewClassTypeRepo(db)
		classRepo := models.NewClassRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		instructor := createTestUser(t, userRepo, "att-lu3-inst@example.com", "password", "instructor")
		ct := createTestClassType(t, classTypeRepo, "Yoga", nil)
		cls := createTestClass(t, classRepo, ct.ID, instructor.ID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
		student := createTestUser(t, userRepo, "att-lu3-student@example.com", "password", "user")

		attendanceRepo.RecordAttendance(&models.Attendance{ClassID: cls.ID, UserID: student.ID})

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "instructor", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(student.ID)+"/attendance", nil, student.ID, "user")
		req.SetPathValue("id", itoa(student.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("user cannot view other user attendance", func(t *testing.T) {
		db := setupTestDB(t)
		userRepo := models.NewUserRepo(db)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		s1 := createTestUser(t, userRepo, "att-lu4-s1@example.com", "password", "user")
		s2 := createTestUser(t, userRepo, "att-lu4-s2@example.com", "password", "user")

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "instructor", "user")

		req, _ := authedRequest(t, http.MethodGet, "/api/users/"+itoa(s1.ID)+"/attendance", nil, s2.ID, "user")
		req.SetPathValue("id", itoa(s1.ID))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		db := setupTestDB(t)
		attendanceRepo := models.NewAttendanceRepo(db)
		handler := NewAttendanceHandler(attendanceRepo)

		wrapped := wrapWithMiddleware(handler.ListByUser, "admin", "instructor", "user")

		req := httptest.NewRequest(http.MethodGet, "/api/users/1/attendance", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}
