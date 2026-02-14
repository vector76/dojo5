package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func tempDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func writeMigration(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write migration file: %v", err)
	}
}

func TestMigrateAppliesInOrder(t *testing.T) {
	db := tempDB(t)
	dir := t.TempDir()

	writeMigration(t, dir, "001_create_foo.sql", "CREATE TABLE foo (id INTEGER PRIMARY KEY);")
	writeMigration(t, dir, "002_create_bar.sql", "CREATE TABLE bar (id INTEGER PRIMARY KEY);")

	if err := Migrate(db, dir); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Verify both tables exist
	for _, table := range []string{"foo", "bar"} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestMigrateSkipsAlreadyApplied(t *testing.T) {
	db := tempDB(t)
	dir := t.TempDir()

	writeMigration(t, dir, "001_create_foo.sql", "CREATE TABLE foo (id INTEGER PRIMARY KEY);")

	if err := Migrate(db, dir); err != nil {
		t.Fatalf("first migrate failed: %v", err)
	}

	// Add a second migration
	writeMigration(t, dir, "002_create_bar.sql", "CREATE TABLE bar (id INTEGER PRIMARY KEY);")

	// Run again — should apply only the new one, not fail on the first
	if err := Migrate(db, dir); err != nil {
		t.Fatalf("second migrate failed: %v", err)
	}

	// Verify both tables exist
	for _, table := range []string{"foo", "bar"} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestMigrateRecordsApplied(t *testing.T) {
	db := tempDB(t)
	dir := t.TempDir()

	writeMigration(t, dir, "001_create_foo.sql", "CREATE TABLE foo (id INTEGER PRIMARY KEY);")
	writeMigration(t, dir, "002_create_bar.sql", "CREATE TABLE bar (id INTEGER PRIMARY KEY);")

	if err := Migrate(db, dir); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("failed to query migrations table: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 recorded migrations, got %d", count)
	}
}

func TestMigrateNoFiles(t *testing.T) {
	db := tempDB(t)
	dir := t.TempDir()

	// Should succeed with no migration files
	if err := Migrate(db, dir); err != nil {
		t.Fatalf("migrate with no files should succeed: %v", err)
	}
}

func TestMigrateIgnoresNonSQL(t *testing.T) {
	db := tempDB(t)
	dir := t.TempDir()

	writeMigration(t, dir, "001_create_foo.sql", "CREATE TABLE foo (id INTEGER PRIMARY KEY);")
	writeMigration(t, dir, "README.md", "This is not a migration.")

	if err := Migrate(db, dir); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("failed to query migrations table: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 recorded migration, got %d", count)
	}
}

func TestOpenCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Verify the file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected database file to be created")
	}

	// Verify we can query
	if err := db.Ping(); err != nil {
		t.Errorf("expected db to be pingable: %v", err)
	}
}
