package db

import (
	"database/sql"
	"log"
)

func InitDB(file string) *sql.DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}

	err = createTables(db)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func createTables(db *sql.DB) error {

	// Create projects table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT check(length(name) <= 100) NOT NULL,
      summary TEXT CHECK(length(summary) <= 255),
			desc TEXT DEFAULT '',
			status TEXT CHECK(status IN ('todo', 'in progress', 'completed', 'archived')) NOT NULL DEFAULT 'todo',
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_updated DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create tasks table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER,
			title TEXT check(length(title) <= 100) NOT NULL,
			desc TEXT DEFAULT '',
			status TEXT CHECK(status IN ('todo', 'in progress', 'completed', 'archived')) NOT NULL DEFAULT 'todo',
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER,
			title TEXT check(length(title) <= 100),
			desc TEXT DEFAULT '',
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	return nil
}
