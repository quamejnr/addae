package service

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type Project struct {
	ID          int
	Name        string
	Summary     string
	Desc        string
	Status      string
	DateCreated time.Time
	DateUpdated time.Time
}

func (p Project) Title() string       { return p.Name }
func (p Project) Description() string { return p.Status }
func (p Project) FilterValue() string { return p.Name }

type Task struct {
	ID          int
	ProjectID   int
	Title       string
	Desc        string
	Status      string
	DateCreated time.Time
	DateUpdated time.Time
}

type Log struct {
	ID          int
	ProjectID   int
	Title       string
	Desc        string
	DateCreated time.Time
	DateUpdated time.Time
}

// Project CRUD operations
func (s *Service) CreateProject(p *Project) error {
	_, err := s.db.Exec(`
		INSERT INTO projects (name, summary, desc, status, date_created, date_updated)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, p.Name, p.Summary, p.Desc, p.Status)
	return err
}

func (s *Service) GetProject(id int) (*Project, error) {
	project := &Project{}
	err := s.db.QueryRow(`
		SELECT id, name, summary, desc, status, date_created, date_updated 
		FROM projects WHERE id = ?
	`, id).Scan(&project.ID, &project.Name, &project.Summary, &project.Desc, &project.Status,
		&project.DateCreated, &project.DateUpdated)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	return project, err
}

func (s *Service) UpdateProject(p *Project) error {
	result, err := s.db.Exec(`
		UPDATE projects 
		SET name = ?, summary = ?, desc = ?, status = ?, date_updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`, p.Name, p.Summary, p.Desc, p.Status, p.ID)
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

func (s *Service) DeleteProject(id int) error {
	result, err := s.db.Exec("DELETE FROM projects WHERE id = ?", id)
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
func (s *Service) CreateTask(projectID int, title, desc string) error {
	_, err := s.db.Exec(`
		INSERT INTO tasks (project_id, title, desc, date_created, date_updated)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, projectID, title, desc)
	return err
}

func (s *Service) UpdateTask(id int, title, desc, status string) error {
	result, err := s.db.Exec(`
		UPDATE tasks 
		SET title = ?, desc = ?, status = ?, date_updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, desc, status, id)
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

func (s *Service) DeleteTask(id int) error {
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
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
func (s *Service) CreateLog(projectID int, title, desc string) error {
	_, err := s.db.Exec(`
		INSERT INTO logs (project_id, title, desc, date_created, date_updated)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, projectID, title, desc)
	return err
}

func (s *Service) UpdateLog(id int, title, desc string) error {
	result, err := s.db.Exec(`
		UPDATE logs 
		SET title = ?, desc = ?, date_updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, desc, id)
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

func (s *Service) DeleteLog(id int) error {
	result, err := s.db.Exec("DELETE FROM logs WHERE id = ?", id)
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
func (s *Service) ListProjects() ([]Project, error) {
	rows, err := s.db.Query(`
		SELECT id, name, summary, desc, status, date_created, date_updated 
		FROM projects
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Name, &p.Summary, &p.Desc, &p.Status,
			&p.DateCreated, &p.DateUpdated)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *Service) ListProjectTasks(projectID int) ([]Task, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, title, desc, status, date_created, date_updated 
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
		err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Desc, &t.Status,
			&t.DateCreated, &t.DateUpdated)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *Service) ListProjectLogs(projectID int) ([]Log, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, title, desc, date_created, date_updated 
		FROM logs 
		WHERE project_id = ?
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		err := rows.Scan(&l.ID, &l.ProjectID, &l.Title, &l.Desc, &l.DateCreated, &l.DateUpdated)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
