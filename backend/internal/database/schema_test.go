package database

import (
	"path/filepath"
	"runtime"
	"testing"
)

// migrationsDir returns the path to the backend/migrations directory.
func migrationsDir(t *testing.T) string {
	t.Helper()
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine caller")
	}
	return filepath.Join(filepath.Dir(f), "..", "..", "migrations")
}

func TestUsersMigration(t *testing.T) {
	db := tempDB(t)

	if err := Migrate(db, migrationsDir(t)); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Verify users table exists
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	if err != nil {
		t.Fatalf("expected users table to exist: %v", err)
	}

	// Verify we can insert a user with all fields
	_, err = db.Exec(`INSERT INTO users (name, email, phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance)
		VALUES ('Test User', 'test@example.com', '555-1234', 'admin', 'hash123', 'monthly', 'active', 'Jane Doe 555-5678', '2025-01-15', 0)`)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Verify we can read back the user with auto-generated fields
	var id int64
	var createdAt, updatedAt string
	err = db.QueryRow("SELECT id, created_at, updated_at FROM users WHERE email='test@example.com'").Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		t.Fatalf("failed to read user: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero auto-generated id")
	}
	if createdAt == "" {
		t.Error("expected non-empty created_at")
	}
	if updatedAt == "" {
		t.Error("expected non-empty updated_at")
	}

	// Verify email uniqueness constraint
	_, err = db.Exec(`INSERT INTO users (name, email, phone, password_hash)
		VALUES ('Duplicate', 'test@example.com', '555-9999', 'hash456')`)
	if err == nil {
		t.Error("expected unique constraint violation for duplicate email")
	}

	// Verify role CHECK constraint
	_, err = db.Exec(`INSERT INTO users (name, email, phone, password_hash, role)
		VALUES ('Bad Role', 'bad@example.com', '555-0000', 'hash789', 'superadmin')`)
	if err == nil {
		t.Error("expected CHECK constraint violation for invalid role")
	}

	// Verify default role is 'user'
	_, err = db.Exec(`INSERT INTO users (name, email, phone, password_hash)
		VALUES ('Default Role', 'default@example.com', '555-1111', 'hash000')`)
	if err != nil {
		t.Fatalf("failed to insert user with default role: %v", err)
	}
	var role string
	err = db.QueryRow("SELECT role FROM users WHERE email='default@example.com'").Scan(&role)
	if err != nil {
		t.Fatalf("failed to read role: %v", err)
	}
	if role != "user" {
		t.Errorf("expected default role 'user', got %q", role)
	}

	// Verify soft delete field is nullable
	var deletedAt *string
	err = db.QueryRow("SELECT deleted_at FROM users WHERE email='test@example.com'").Scan(&deletedAt)
	if err != nil {
		t.Fatalf("failed to read deleted_at: %v", err)
	}
	if deletedAt != nil {
		t.Error("expected deleted_at to be NULL for active user")
	}
}

func TestClassTypesMigration(t *testing.T) {
	db := tempDB(t)

	if err := Migrate(db, migrationsDir(t)); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Insert a class type
	res, err := db.Exec(`INSERT INTO class_types (name, description) VALUES ('Beginner Yoga', 'Intro-level yoga class')`)
	if err != nil {
		t.Fatalf("failed to insert class type: %v", err)
	}

	id, _ := res.LastInsertId()
	if id == 0 {
		t.Error("expected non-zero id")
	}

	// Verify timestamps are auto-populated
	var createdAt, updatedAt string
	err = db.QueryRow("SELECT created_at, updated_at FROM class_types WHERE id=?", id).Scan(&createdAt, &updatedAt)
	if err != nil {
		t.Fatalf("failed to read class type: %v", err)
	}
	if createdAt == "" || updatedAt == "" {
		t.Error("expected non-empty timestamps")
	}
}

func TestClassesMigration(t *testing.T) {
	db := tempDB(t)

	if err := Migrate(db, migrationsDir(t)); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Set up prerequisite data
	db.Exec(`INSERT INTO users (name, email, phone, password_hash, role) VALUES ('Instructor', 'inst@example.com', '555-0001', 'hash', 'instructor')`)
	db.Exec(`INSERT INTO class_types (name) VALUES ('Karate')`)

	// Insert a class
	_, err := db.Exec(`INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		VALUES (1, 1, '2025-06-15 10:00:00', 60, 20)`)
	if err != nil {
		t.Fatalf("failed to insert class: %v", err)
	}

	// Verify FK constraint — invalid class_type_id
	_, err = db.Exec(`INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		VALUES (999, 1, '2025-06-15 11:00:00', 60, 20)`)
	if err == nil {
		t.Error("expected FK constraint violation for invalid class_type_id")
	}

	// Verify FK constraint — invalid instructor_id
	_, err = db.Exec(`INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		VALUES (1, 999, '2025-06-15 12:00:00', 60, 20)`)
	if err == nil {
		t.Error("expected FK constraint violation for invalid instructor_id")
	}
}

func TestAttendanceMigration(t *testing.T) {
	db := tempDB(t)

	if err := Migrate(db, migrationsDir(t)); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Set up prerequisite data
	db.Exec(`INSERT INTO users (name, email, phone, password_hash, role) VALUES ('Instructor', 'inst@example.com', '555-0001', 'hash', 'instructor')`)
	db.Exec(`INSERT INTO users (name, email, phone, password_hash) VALUES ('Student', 'student@example.com', '555-0002', 'hash')`)
	db.Exec(`INSERT INTO class_types (name) VALUES ('Yoga')`)
	db.Exec(`INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity) VALUES (1, 1, '2025-06-15 10:00:00', 60, 20)`)

	// Record attendance
	_, err := db.Exec(`INSERT INTO attendance (class_id, user_id) VALUES (1, 2)`)
	if err != nil {
		t.Fatalf("failed to insert attendance: %v", err)
	}

	// Verify checked_in_at is auto-populated
	var checkedInAt string
	err = db.QueryRow("SELECT checked_in_at FROM attendance WHERE class_id=1 AND user_id=2").Scan(&checkedInAt)
	if err != nil {
		t.Fatalf("failed to read attendance: %v", err)
	}
	if checkedInAt == "" {
		t.Error("expected non-empty checked_in_at")
	}

	// Verify FK constraint — invalid class_id
	_, err = db.Exec(`INSERT INTO attendance (class_id, user_id) VALUES (999, 2)`)
	if err == nil {
		t.Error("expected FK constraint violation for invalid class_id")
	}
}

func TestPaymentsMigration(t *testing.T) {
	db := tempDB(t)

	if err := Migrate(db, migrationsDir(t)); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Set up prerequisite data
	db.Exec(`INSERT INTO users (name, email, phone, password_hash, role) VALUES ('Admin', 'admin@example.com', '555-0001', 'hash', 'admin')`)
	db.Exec(`INSERT INTO users (name, email, phone, password_hash) VALUES ('Student', 'student@example.com', '555-0002', 'hash')`)

	// Record a payment
	_, err := db.Exec(`INSERT INTO payments (user_id, amount, date, note, recorded_by)
		VALUES (2, 50.00, '2025-06-15', 'Monthly dues', 1)`)
	if err != nil {
		t.Fatalf("failed to insert payment: %v", err)
	}

	// Verify timestamps
	var createdAt, updatedAt string
	err = db.QueryRow("SELECT created_at, updated_at FROM payments WHERE user_id=2").Scan(&createdAt, &updatedAt)
	if err != nil {
		t.Fatalf("failed to read payment: %v", err)
	}
	if createdAt == "" || updatedAt == "" {
		t.Error("expected non-empty timestamps")
	}

	// Verify FK constraint — invalid user_id
	_, err = db.Exec(`INSERT INTO payments (user_id, amount, date, recorded_by) VALUES (999, 25.00, '2025-06-15', 1)`)
	if err == nil {
		t.Error("expected FK constraint violation for invalid user_id")
	}

	// Verify FK constraint — invalid recorded_by
	_, err = db.Exec(`INSERT INTO payments (user_id, amount, date, recorded_by) VALUES (2, 25.00, '2025-06-15', 999)`)
	if err == nil {
		t.Error("expected FK constraint violation for invalid recorded_by")
	}
}

func TestAllMigrationsApply(t *testing.T) {
	db := tempDB(t)

	if err := Migrate(db, migrationsDir(t)); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Verify all expected tables exist
	expectedTables := []string{"users", "class_types", "classes", "attendance", "payments"}
	for _, table := range expectedTables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}

	// Verify all 5 migrations were recorded
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("failed to query migrations: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 recorded migrations, got %d", count)
	}
}
