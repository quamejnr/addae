
package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitDB(t *testing.T) {
	dbPath := ":memory:"
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestRunMigrations(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create a temporary migrations directory for testing
	migrationsDir := t.TempDir()
	file, err := os.Create(migrationsDir + "/20250204221253_init_schema.sql")
	if err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}
	_, err = file.WriteString(`-- +goose Up
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    summary TEXT,
    desc TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    date_created TIMESTAMP NOT NULL,
    date_updated TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE projects;
`)
	if err != nil {
		t.Fatalf("failed to write to migration file: %v", err)
	}
	file.Close()

	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify that the table was created
	_, err = db.Exec("SELECT * FROM projects")
	if err != nil {
		t.Errorf("failed to query projects table: %v", err)
	}
}
