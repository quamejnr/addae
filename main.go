// main.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Project struct {
	ID          int
	Name        string
	Description string
	Status      string
	DateCreated time.Time
	DateUpdated time.Time
}

type Task struct {
	ID          int
	ProjectID   int
	Title       string
	Description string
	Status      string
	DateCreated time.Time
	DateUpdated time.Time
}

type Log struct {
	ID          int
	ProjectID   int
	Title       string
	DateCreated time.Time
	DateUpdated time.Time
}

func initDB(file string) *sql.DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}

	// Create projects table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT CHECK(status IN ('todo', 'in progress', 'completed', 'archived')) NOT NULL DEFAULT 'todo',
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_updated DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create tasks table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT CHECK(status IN ('todo', 'in progress', 'completed', 'archived')) NOT NULL DEFAULT 'todo',
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER,
			title TEXT NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// Project CRUD operations
func createProject(db *sql.DB, name, description string) error {
	_, err := db.Exec(`
		INSERT INTO projects (name, description, status, date_created, date_updated)
		VALUES (?, ?, 'todo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, name, description)
	return err
}

func getProject(db *sql.DB, id int) (*Project, error) {
	project := &Project{}
	err := db.QueryRow(`
		SELECT id, name, description, status, date_created, date_updated 
		FROM projects WHERE id = ?
	`, id).Scan(&project.ID, &project.Name, &project.Description, &project.Status,
		&project.DateCreated, &project.DateUpdated)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	return project, err
}

func updateProject(db *sql.DB, id int, name, description, status string) error {
	result, err := db.Exec(`
		UPDATE projects 
		SET name = ?, description = ?, status = ?, date_updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`, name, description, status, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

func deleteProject(db *sql.DB, id int) error {
	result, err := db.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

// Task CRUD operations
func createTask(db *sql.DB, projectID int, title, description string) error {
	_, err := db.Exec(`
		INSERT INTO tasks (project_id, title, description, date_created, date_updated)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, projectID, title, description)
	return err
}

func updateTask(db *sql.DB, id int, title, description, status string) error {
	result, err := db.Exec(`
		UPDATE tasks 
		SET title = ?, description = ?, status = ?, date_updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, description, status, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func deleteTask(db *sql.DB, id int) error {
	result, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// Log CRUD operations
func createLog(db *sql.DB, projectID int, title string) error {
	_, err := db.Exec(`
		INSERT INTO logs (project_id, title, date_created, date_updated)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, projectID, title)
	return err
}

func updateLog(db *sql.DB, id int, title string) error {
	result, err := db.Exec(`
		UPDATE logs 
		SET title = ?, date_updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("log not found")
	}
	return nil
}

func deleteLog(db *sql.DB, id int) error {
	result, err := db.Exec("DELETE FROM logs WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("log not found")
	}
	return nil
}

// List functions
func listProjects(db *sql.DB) ([]Project, error) {
	rows, err := db.Query(`
		SELECT id, name, description, status, date_created, date_updated 
		FROM projects
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status,
			&p.DateCreated, &p.DateUpdated)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func listProjectTasks(db *sql.DB, projectID int) ([]Task, error) {
	rows, err := db.Query(`
		SELECT id, project_id, title, description, status, date_created, date_updated 
		FROM tasks 
		WHERE project_id = ?
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status,
			&t.DateCreated, &t.DateUpdated)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func listProjectLogs(db *sql.DB, projectID int) ([]Log, error) {
	rows, err := db.Query(`
		SELECT id, project_id, title, date_created, date_updated 
		FROM logs 
		WHERE project_id = ?
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var n Log
		err := rows.Scan(&n.ID, &n.ProjectID, &n.Title, &n.DateCreated, &n.DateUpdated)
		if err != nil {
			return nil, err
		}
		logs = append(logs, n)
	}
	return logs, nil
}

func main() {
	db := initDB("./addae.db")
	defer db.Close()

	// Command line flags
	createProjectCmd := flag.NewFlagSet("create-project", flag.ExitOnError)
	projectName := createProjectCmd.String("name", "", "Project name")
	projectDesc := createProjectCmd.String("desc", "", "Project description")

	createTaskCmd := flag.NewFlagSet("create-task", flag.ExitOnError)
	taskProjectID := createTaskCmd.Int("project", 0, "Project ID")
	taskTitle := createTaskCmd.String("title", "", "Task title")
	taskDesc := createTaskCmd.String("desc", "", "Task description")

	createLogCmd := flag.NewFlagSet("create-log", flag.ExitOnError)
	logProjectID := createLogCmd.Int("project", 0, "Project ID")
	logTitle := createLogCmd.String("title", "", "Log title")

	if len(os.Args) < 2 {
		fmt.Println("Expected subcommands: create-project, create-task, create-log, list-projects")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create-project":
		createProjectCmd.Parse(os.Args[2:])
		if err := createProject(db, *projectName, *projectDesc); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Project created successfully")

	case "create-task":
		createTaskCmd.Parse(os.Args[2:])
		if err := createTask(db, *taskProjectID, *taskTitle, *taskDesc); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Task created successfully")

	case "create-log":
		createLogCmd.Parse(os.Args[2:])
		if err := createLog(db, *logProjectID, *logTitle); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Log created successfully")

	case "list-projects":
		projects, err := listProjects(db)
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range projects {
			fmt.Printf("Project %d: %s (%s)\n", p.ID, p.Name, p.Status)
			tasks, err := listProjectTasks(db, p.ID)
			if err != nil {
				log.Fatal(err)
			}
			for _, t := range tasks {
				fmt.Printf("  Task: %s (%s)\n", t.Title, t.Status)
			}
			logs, err := listProjectLogs(db, p.ID)
			if err != nil {
				log.Fatal(err)
			}
			for _, n := range logs {
				fmt.Printf("  Log: %s\n", n.Title)
			}
		}

	default:
		fmt.Println("Expected subcommands: create-project, create-task, create-log, list-projects")
		os.Exit(1)
	}
}
