package service

import (
	"database/sql"
	"testing"

	adb "github.com/quamejnr/addae/internal/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Run migrations
	if err := adb.RunMigrations(db, "../db/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func TestCreateProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}

	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Verify the project was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE name = ?", "Test Project").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query for project: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 project, got %d", count)
	}
}

func TestGetProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to fetch
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project
	// To get the ID, we need to query the database
	var id int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&id)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	fetchedProject, err := service.GetProject(id)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}

	if fetchedProject.Name != project.Name {
		t.Errorf("expected project name %s, got %s", project.Name, fetchedProject.Name)
	}

	if fetchedProject.Summary != project.Summary {
		t.Errorf("expected project summary %s, got %s", project.Summary, fetchedProject.Summary)
	}

	if fetchedProject.Desc != project.Desc {
		t.Errorf("expected project desc %s, got %s", project.Desc, fetchedProject.Desc)
	}

	if fetchedProject.Status != project.Status {
		t.Errorf("expected project status %s, got %s", project.Status, fetchedProject.Status)
	}
}

func TestUpdateProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to update
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var id int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&id)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Update the project
	updatedProject := &Project{
		ID:      id,
		Name:    "Updated Project",
		Summary: "Updated Summary",
		Desc:    "Updated Description",
		Status:  "in progress",
	}

	err = service.UpdateProject(updatedProject)
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}

	// Get the project again to verify the update
	fetchedProject, err := service.GetProject(id)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}

	if fetchedProject.Name != updatedProject.Name {
		t.Errorf("expected project name %s, got %s", updatedProject.Name, fetchedProject.Name)
	}

	if fetchedProject.Summary != updatedProject.Summary {
		t.Errorf("expected project summary %s, got %s", updatedProject.Summary, fetchedProject.Summary)
	}

	if fetchedProject.Desc != updatedProject.Desc {
		t.Errorf("expected project desc %s, got %s", updatedProject.Desc, fetchedProject.Desc)
	}

	if fetchedProject.Status != updatedProject.Status {
		t.Errorf("expected project status %s, got %s", updatedProject.Status, fetchedProject.Status)
	}
}

func TestDeleteProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to delete
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var id int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&id)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Delete the project
	err = service.DeleteProject(id)
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	// Verify the project was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query for project: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 projects, got %d", count)
	}
}

func TestCreateTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to associate the task with
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a task
	err = service.CreateTask(projectID, "Test Task", "Test Description")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Verify the task was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ?", "Test Task").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query for task: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 task, got %d", count)
	}
}

func TestUpdateTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to associate the task with
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a task to update
	err = service.CreateTask(projectID, "Test Task", "Test Description")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Get the task id
	var taskID int
	err = db.QueryRow("SELECT id FROM tasks WHERE title = ?", "Test Task").Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to query for task id: %v", err)
	}

	// Update the task
	err = service.UpdateTask(taskID, "Updated Task", "Updated Description", "in progress")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// Verify the task was updated
	var title, desc, status string
	err = db.QueryRow("SELECT title, desc, status FROM tasks WHERE id = ?", taskID).Scan(&title, &desc, &status)
	if err != nil {
		t.Fatalf("failed to query for updated task: %v", err)
	}

	if title != "Updated Task" {
		t.Errorf("expected task title to be 'Updated Task', got %s", title)
	}

	if desc != "Updated Description" {
		t.Errorf("expected task description to be 'Updated Description', got %s", desc)
	}

	if status != "in progress" {
		t.Errorf("expected task status to be 'in progress', got %s", status)
	}
}

func TestDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to associate the task with
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a task to delete
	err = service.CreateTask(projectID, "Test Task", "Test Description")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Get the task id
	var taskID int
	err = db.QueryRow("SELECT id FROM tasks WHERE title = ?", "Test Task").Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to query for task id: %v", err)
	}

	// Delete the task
	err = service.DeleteTask(taskID)
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Verify the task was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query for task: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 tasks, got %d", count)
	}
}

func TestCreateLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to associate the log with
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a log
	err = service.CreateLog(projectID, "Test Log", "Test Description")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	// Verify the log was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE title = ?", "Test Log").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query for log: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 log, got %d", count)
	}
}

func TestUpdateLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to associate the log with
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a log to update
	err = service.CreateLog(projectID, "Test Log", "Test Description")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	// Get the log id
	var logID int
	err = db.QueryRow("SELECT id FROM logs WHERE title = ?", "Test Log").Scan(&logID)
	if err != nil {
		t.Fatalf("failed to query for log id: %v", err)
	}

	// Update the log
	err = service.UpdateLog(logID, "Updated Log", "Updated Description")
	if err != nil {
		t.Fatalf("UpdateLog failed: %v", err)
	}

	// Verify the log was updated
	var title, desc string
	err = db.QueryRow("SELECT title, desc FROM logs WHERE id = ?", logID).Scan(&title, &desc)
	if err != nil {
		t.Fatalf("failed to query for updated log: %v", err)
	}

	if title != "Updated Log" {
		t.Errorf("expected log title to be 'Updated Log', got %s", title)
	}

	if desc != "Updated Description" {
		t.Errorf("expected log description to be 'Updated Description', got %s", desc)
	}
}

func TestDeleteLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project to associate the log with
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a log to delete
	err = service.CreateLog(projectID, "Test Log", "Test Description")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	// Get the log id
	var logID int
	err = db.QueryRow("SELECT id FROM logs WHERE title = ?", "Test Log").Scan(&logID)
	if err != nil {
		t.Fatalf("failed to query for log id: %v", err)
	}

	// Delete the log
	err = service.DeleteLog(logID)
	if err != nil {
		t.Fatalf("DeleteLog failed: %v", err)
	}

	// Verify the log was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE id = ?", logID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query for log: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 logs, got %d", count)
	}
}

func TestListProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a few projects
	project1 := &Project{
		Name:    "Test Project 1",
		Summary: "Test Summary 1",
		Desc:    "Test Description 1",
		Status:  "todo",
	}
	project2 := &Project{
		Name:    "Test Project 2",
		Summary: "Test Summary 2",
		Desc:    "Test Description 2",
		Status:  "in progress",
	}
	err := service.CreateProject(project1)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	err = service.CreateProject(project2)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// List the projects
	projects, err := service.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	// Verify the number of projects
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestListProjectTasks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a few tasks for the project
	err = service.CreateTask(projectID, "Test Task 1", "Description 1")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	err = service.CreateTask(projectID, "Test Task 2", "Description 2")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// List the tasks for the project
	tasks, err := service.ListProjectTasks(projectID)
	if err != nil {
		t.Fatalf("ListProjectTasks failed: %v", err)
	}

	// Verify the number of tasks
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}

	// Verify task details
	if tasks[0].Title != "Test Task 1" || tasks[1].Title != "Test Task 2" {
		t.Errorf("expected tasks to be 'Test Task 1' and 'Test Task 2', got %s and %s", tasks[0].Title, tasks[1].Title)
	}
}

func TestListProjectLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewService(db)

	// Create a project
	project := &Project{
		Name:    "Test Project",
		Summary: "Test Summary",
		Desc:    "Test Description",
		Status:  "todo",
	}
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Get the project id
	var projectID int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to query for project id: %v", err)
	}

	// Create a few logs for the project
	err = service.CreateLog(projectID, "Test Log 1", "Description 1")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}
	err = service.CreateLog(projectID, "Test Log 2", "Description 2")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	// List the logs for the project
	logs, err := service.ListProjectLogs(projectID)
	if err != nil {
		t.Fatalf("ListProjectLogs failed: %v", err)
	}

	// Verify the number of logs
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}

	// Verify log details
	if logs[0].Title != "Test Log 1" || logs[1].Title != "Test Log 2" {
		t.Errorf("expected logs to be 'Test Log 1' and 'Test Log 2', got %s and %s", logs[0].Title, logs[1].Title)
	}
}
