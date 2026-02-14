package models

import (
	"database/sql"
	"testing"
)

func TestClassTypeRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewClassTypeRepo(db)

	tests := []struct {
		name    string
		ct      *ClassType
		wantErr bool
	}{
		{
			name:    "valid class type",
			ct:      &ClassType{Name: "Yoga"},
			wantErr: false,
		},
		{
			name: "with description",
			ct: func() *ClassType {
				desc := "A beginner-friendly yoga class"
				return &ClassType{Name: "Beginner Yoga", Description: &desc}
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.ct)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.ct.ID == 0 {
				t.Error("expected non-zero ID after create")
			}
		})
	}
}

func TestClassTypeRepo_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewClassTypeRepo(db)

	desc := "Martial arts class"
	ct := &ClassType{Name: "Karate", Description: &desc}
	if err := repo.Create(ct); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"existing class type", ct.ID, false},
		{"non-existent class type", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if got.ID != ct.ID {
					t.Errorf("expected ID %d, got %d", ct.ID, got.ID)
				}
				if got.Name != "Karate" {
					t.Errorf("expected name Karate, got %q", got.Name)
				}
				if got.Description == nil || *got.Description != "Martial arts class" {
					t.Errorf("expected description 'Martial arts class', got %v", got.Description)
				}
			}
		})
	}
}

func TestClassTypeRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewClassTypeRepo(db)

	names := []string{"Yoga", "Karate", "Pilates"}
	for _, name := range names {
		if err := repo.Create(&ClassType{Name: name}); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	types, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(types) != 3 {
		t.Fatalf("expected 3 class types, got %d", len(types))
	}

	// Verify sorted by name
	if types[0].Name != "Karate" {
		t.Errorf("expected first type Karate, got %q", types[0].Name)
	}
	if types[1].Name != "Pilates" {
		t.Errorf("expected second type Pilates, got %q", types[1].Name)
	}
	if types[2].Name != "Yoga" {
		t.Errorf("expected third type Yoga, got %q", types[2].Name)
	}
}

func TestClassTypeRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewClassTypeRepo(db)

	ct := &ClassType{Name: "Yoga"}
	if err := repo.Create(ct); err != nil {
		t.Fatalf("setup: %v", err)
	}

	ct.Name = "Hot Yoga"
	desc := "Heated yoga class"
	ct.Description = &desc

	if err := repo.Update(ct); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := repo.GetByID(ct.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if got.Name != "Hot Yoga" {
		t.Errorf("expected name 'Hot Yoga', got %q", got.Name)
	}
	if got.Description == nil || *got.Description != "Heated yoga class" {
		t.Errorf("expected description 'Heated yoga class', got %v", got.Description)
	}
}

func TestClassTypeRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewClassTypeRepo(db)

	ct := &ClassType{Name: "Yoga"}
	if err := repo.Create(ct); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"existing class type", ct.ID, false},
		{"already deleted", ct.ID, true},
		{"non-existent class type", 99999, true},
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
	_, err := repo.GetByID(ct.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}
}
