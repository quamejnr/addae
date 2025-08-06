package db

import (
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes and returns a database connection.
// If dbPath is empty, it will use a default OS-specific path.
func InitDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		var err error
		dbPath, err = getDefaultDBPath()
		if err != nil {
			return nil, fmt.Errorf("failed to get default DB path: %w", err)
		}
	}

	// Construct the DSN to automatically create the file and enable foreign keys
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=ON", dbPath)

	// Open the database connection
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Use defer to ensure the database connection is closed on error
	// If the function returns successfully, the caller is responsible for closing db.
	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return db, nil
}

// getDefaultDBPath returns the default OS-specific path for the database file.
func getDefaultDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	appDir := filepath.Join(configDir, "addae")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create application directory: %w", err)
	}

	return filepath.Join(appDir, "addae.db"), nil
}

func RunMigrations(db *sql.DB, migrationsDir string) error {
	// goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// temporarily disabling logs in the goose.Up function after migration
	log.SetOutput(io.Discard)

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// restore logs to stdout
	log.SetOutput(os.Stdout)
	return nil
}

func RunMigrationsFromFS(db *sql.DB, migrationsFS fs.FS) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	// temporarily disabling logs in the goose.Up function after migration
	log.SetOutput(io.Discard)

	goose.SetBaseFS(migrationsFS)
	return goose.Up(db, ".")
}
