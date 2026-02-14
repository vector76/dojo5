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
