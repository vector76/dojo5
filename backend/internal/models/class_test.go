package models

import (
	"database/sql"
	"testing"
	"time"
)

// createTestClassType creates a class type for use in class tests.
func createTestClassType(t *testing.T, repo *ClassTypeRepo, name string) *ClassType {
	t.Helper()
	ct := &ClassType{Name: name}
	if err := repo.Create(ct); err != nil {
		t.Fatalf("failed to create class type: %v", err)
	}
	return ct
}

// createTestInstructor creates a user with the instructor role for use in class tests.
func createTestInstructor(t *testing.T, repo *UserRepo, email string) *User {
	t.Helper()
	u := &User{
		Name:         "Instructor",
		Email:        email,
		Phone:        "555-0000",
		Role:         "instructor",
		PasswordHash: "hashed",
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("failed to create instructor: %v", err)
	}
	return u
}

func TestClassRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	classTypeRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	repo := NewClassRepo(db)

	ct := createTestClassType(t, classTypeRepo, "Yoga")
	inst := createTestInstructor(t, userRepo, "inst@example.com")

	tests := []struct {
		name    string
		class   *Class
		wantErr bool
	}{
		{
			name: "valid class",
			class: &Class{
				ClassTypeID:     ct.ID,
				InstructorID:    inst.ID,
				StartTime:       time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
				DurationMinutes: 60,
				Capacity:        20,
			},
			wantErr: false,
		},
		{
			name: "invalid class type",
			class: &Class{
				ClassTypeID:     99999,
				InstructorID:    inst.ID,
				StartTime:       time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC),
				DurationMinutes: 45,
				Capacity:        15,
			},
			wantErr: true,
		},
		{
			name: "invalid instructor",
			class: &Class{
				ClassTypeID:     ct.ID,
				InstructorID:    99999,
				StartTime:       time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
				DurationMinutes: 30,
				Capacity:        10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.class)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.class.ID == 0 {
				t.Error("expected non-zero ID after create")
			}
		})
	}
}

func TestClassRepo_GetByID(t *testing.T) {
	db := setupTestDB(t)
	classTypeRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	repo := NewClassRepo(db)

	ct := createTestClassType(t, classTypeRepo, "Pilates")
	inst := createTestInstructor(t, userRepo, "pilates-inst@example.com")

	startTime := time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC)
	c := &Class{
		ClassTypeID:     ct.ID,
		InstructorID:    inst.ID,
		StartTime:       startTime,
		DurationMinutes: 45,
		Capacity:        15,
	}
	if err := repo.Create(c); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"existing class", c.ID, false},
		{"non-existent class", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if got.ID != c.ID {
					t.Errorf("expected ID %d, got %d", c.ID, got.ID)
				}
				if got.ClassTypeID != ct.ID {
					t.Errorf("expected class_type_id %d, got %d", ct.ID, got.ClassTypeID)
				}
				if got.InstructorID != inst.ID {
					t.Errorf("expected instructor_id %d, got %d", inst.ID, got.InstructorID)
				}
				if got.DurationMinutes != 45 {
					t.Errorf("expected duration 45, got %d", got.DurationMinutes)
				}
				if got.Capacity != 15 {
					t.Errorf("expected capacity 15, got %d", got.Capacity)
				}
			}
		})
	}
}

func TestClassRepo_List(t *testing.T) {
	db := setupTestDB(t)
	classTypeRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	repo := NewClassRepo(db)

	yoga := createTestClassType(t, classTypeRepo, "Yoga")
	karate := createTestClassType(t, classTypeRepo, "Karate")
	inst1 := createTestInstructor(t, userRepo, "inst1@example.com")
	inst2 := createTestInstructor(t, userRepo, "inst2@example.com")

	// Create classes with different types and instructors
	classes := []Class{
		{ClassTypeID: yoga.ID, InstructorID: inst1.ID, StartTime: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), DurationMinutes: 60, Capacity: 20},
		{ClassTypeID: yoga.ID, InstructorID: inst2.ID, StartTime: time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC), DurationMinutes: 60, Capacity: 20},
		{ClassTypeID: karate.ID, InstructorID: inst1.ID, StartTime: time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC), DurationMinutes: 90, Capacity: 15},
	}
	for i := range classes {
		if err := repo.Create(&classes[i]); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	tests := []struct {
		name      string
		filter    ClassFilter
		wantCount int
	}{
		{
			name:      "no filter",
			filter:    ClassFilter{},
			wantCount: 3,
		},
		{
			name:      "filter by class type",
			filter:    ClassFilter{ClassTypeID: &yoga.ID},
			wantCount: 2,
		},
		{
			name:      "filter by instructor",
			filter:    ClassFilter{InstructorID: &inst1.ID},
			wantCount: 2,
		},
		{
			name:      "filter by both",
			filter:    ClassFilter{ClassTypeID: &yoga.ID, InstructorID: &inst1.ID},
			wantCount: 1,
		},
		{
			name: "filter with no results",
			filter: func() ClassFilter {
				id := int64(99999)
				return ClassFilter{ClassTypeID: &id}
			}(),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.List(tt.filter)
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("expected %d classes, got %d", tt.wantCount, len(got))
			}
		})
	}

	// Verify ordering by start_time
	all, err := repo.List(ClassFilter{})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(all) >= 2 {
		for i := 1; i < len(all); i++ {
			if all[i].StartTime.Before(all[i-1].StartTime) {
				t.Errorf("classes not ordered by start_time: %v before %v", all[i].StartTime, all[i-1].StartTime)
			}
		}
	}
}

func TestClassRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	classTypeRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	repo := NewClassRepo(db)

	ct := createTestClassType(t, classTypeRepo, "Yoga")
	inst := createTestInstructor(t, userRepo, "update-inst@example.com")

	c := &Class{
		ClassTypeID:     ct.ID,
		InstructorID:    inst.ID,
		StartTime:       time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		DurationMinutes: 60,
		Capacity:        20,
	}
	if err := repo.Create(c); err != nil {
		t.Fatalf("setup: %v", err)
	}

	c.DurationMinutes = 90
	c.Capacity = 25
	c.StartTime = time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)

	if err := repo.Update(c); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := repo.GetByID(c.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if got.DurationMinutes != 90 {
		t.Errorf("expected duration 90, got %d", got.DurationMinutes)
	}
	if got.Capacity != 25 {
		t.Errorf("expected capacity 25, got %d", got.Capacity)
	}
}

func TestClassRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	classTypeRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	repo := NewClassRepo(db)

	ct := createTestClassType(t, classTypeRepo, "Yoga")
	inst := createTestInstructor(t, userRepo, "delete-inst@example.com")

	c := &Class{
		ClassTypeID:     ct.ID,
		InstructorID:    inst.ID,
		StartTime:       time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		DurationMinutes: 60,
		Capacity:        20,
	}
	if err := repo.Create(c); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"existing class", c.ID, false},
		{"already deleted", c.ID, true},
		{"non-existent class", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify it's gone
	_, err := repo.GetByID(c.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}
}
