package models

import (
	"fmt"
	"testing"
	"time"
)

func TestAttendanceRepo_RecordAttendance(t *testing.T) {
	db := setupTestDB(t)
	ctRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	classRepo := NewClassRepo(db)
	repo := NewAttendanceRepo(db)

	ct := createTestClassType(t, ctRepo, "Yoga")
	inst := createTestInstructor(t, userRepo, "att-inst@example.com")
	student := &User{Name: "Student", Email: "student@example.com", Phone: "555-1111", Role: "user", PasswordHash: "hashed"}
	if err := userRepo.Create(student); err != nil {
		t.Fatalf("setup: %v", err)
	}
	class := &Class{
		ClassTypeID: ct.ID, InstructorID: inst.ID,
		StartTime: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), DurationMinutes: 60, Capacity: 20,
	}
	if err := classRepo.Create(class); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		att     *Attendance
		wantErr bool
	}{
		{
			name:    "valid attendance",
			att:     &Attendance{ClassID: class.ID, UserID: student.ID},
			wantErr: false,
		},
		{
			name:    "invalid class",
			att:     &Attendance{ClassID: 99999, UserID: student.ID},
			wantErr: true,
		},
		{
			name:    "invalid user",
			att:     &Attendance{ClassID: class.ID, UserID: 99999},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RecordAttendance(tt.att)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordAttendance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.att.ID == 0 {
				t.Error("expected non-zero ID after record")
			}
		})
	}
}

func TestAttendanceRepo_ListByClass(t *testing.T) {
	db := setupTestDB(t)
	ctRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	classRepo := NewClassRepo(db)
	repo := NewAttendanceRepo(db)

	ct := createTestClassType(t, ctRepo, "Karate")
	inst := createTestInstructor(t, userRepo, "lbc-inst@example.com")

	class1 := &Class{
		ClassTypeID: ct.ID, InstructorID: inst.ID,
		StartTime: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), DurationMinutes: 60, Capacity: 20,
	}
	class2 := &Class{
		ClassTypeID: ct.ID, InstructorID: inst.ID,
		StartTime: time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC), DurationMinutes: 60, Capacity: 20,
	}
	if err := classRepo.Create(class1); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := classRepo.Create(class2); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Create students
	students := make([]*User, 3)
	for i := range students {
		students[i] = &User{
			Name: "Student", Email: fmt.Sprintf("lbc-student%d@example.com", i),
			Phone: "555-0000", Role: "user", PasswordHash: "hashed",
		}
		if err := userRepo.Create(students[i]); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	// Record attendance: 2 students in class1, 1 in class2
	for _, s := range students[:2] {
		if err := repo.RecordAttendance(&Attendance{ClassID: class1.ID, UserID: s.ID}); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	if err := repo.RecordAttendance(&Attendance{ClassID: class2.ID, UserID: students[2].ID}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name      string
		classID   int64
		wantCount int
	}{
		{"class with 2 attendees", class1.ID, 2},
		{"class with 1 attendee", class2.ID, 1},
		{"class with no attendees", 99999, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ListByClass(tt.classID)
			if err != nil {
				t.Fatalf("ListByClass() error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("expected %d records, got %d", tt.wantCount, len(got))
			}
			for _, a := range got {
				if a.ClassID != tt.classID {
					t.Errorf("expected class_id %d, got %d", tt.classID, a.ClassID)
				}
			}
		})
	}
}

func TestAttendanceRepo_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	ctRepo := NewClassTypeRepo(db)
	userRepo := NewUserRepo(db)
	classRepo := NewClassRepo(db)
	repo := NewAttendanceRepo(db)

	ct := createTestClassType(t, ctRepo, "Pilates")
	inst := createTestInstructor(t, userRepo, "lbu-inst@example.com")

	// Create 2 classes
	classes := make([]*Class, 2)
	for i := range classes {
		classes[i] = &Class{
			ClassTypeID: ct.ID, InstructorID: inst.ID,
			StartTime:       time.Date(2026, 3, 1+i, 10, 0, 0, 0, time.UTC),
			DurationMinutes: 60, Capacity: 20,
		}
		if err := classRepo.Create(classes[i]); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	user1 := &User{Name: "User1", Email: "lbu-user1@example.com", Phone: "555-0000", Role: "user", PasswordHash: "hashed"}
	user2 := &User{Name: "User2", Email: "lbu-user2@example.com", Phone: "555-0001", Role: "user", PasswordHash: "hashed"}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// user1 attends both classes, user2 attends 1
	for _, c := range classes {
		if err := repo.RecordAttendance(&Attendance{ClassID: c.ID, UserID: user1.ID}); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	if err := repo.RecordAttendance(&Attendance{ClassID: classes[0].ID, UserID: user2.ID}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name      string
		userID    int64
		wantCount int
	}{
		{"user with 2 classes", user1.ID, 2},
		{"user with 1 class", user2.ID, 1},
		{"user with no attendance", 99999, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.ListByUser(tt.userID)
			if err != nil {
				t.Fatalf("ListByUser() error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("expected %d records, got %d", tt.wantCount, len(got))
			}
			for _, a := range got {
				if a.UserID != tt.userID {
					t.Errorf("expected user_id %d, got %d", tt.userID, a.UserID)
				}
			}
		})
	}
}
