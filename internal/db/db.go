package db

import (
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Create the database file
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("error creating database file: %w", err)
		}
		file.Close()
	}

	// Open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Enable foreign key support
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	return db, nil
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
