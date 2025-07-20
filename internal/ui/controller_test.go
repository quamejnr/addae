package ui

import (
	"errors"
	"testing"

	"github.com/quamejnr/addae/internal/service"
)

// MockService is a mock implementation of the Service interface for testing.
type MockService struct {
	projects []service.Project
	tasks    []service.Task
	logs     []service.Log
	err      error
}

func (m *MockService) ListProjects() ([]service.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.projects, nil
}

func (m *MockService) DeleteProject(id int) error {
	if m.err != nil {
		return m.err
	}
	for i, p := range m.projects {
		if p.ID == id {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			return nil
		}
	}
	return errors.New("project not found")
}

func (m *MockService) CreateProject(p *service.Project) error {
	if m.err != nil {
		return m.err
	}
	p.ID = len(m.projects) + 1
	m.projects = append(m.projects, *p)
	return nil
}

func (m *MockService) UpdateProject(p *service.Project) error {
	if m.err != nil {
		return m.err
	}
	for i, proj := range m.projects {
		if proj.ID == p.ID {
			m.projects[i] = *p
			return nil
		}
	}
	return errors.New("project not found")
}

func (m *MockService) ListProjectTasks(projectID int) ([]service.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasks, nil
}

func (m *MockService) ListProjectLogs(projectID int) ([]service.Log, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.logs, nil
}

func (m *MockService) CreateTask(projectID int, title, desc string) error {
	if m.err != nil {
		return m.err
	}
	task := service.Task{
		ID:        len(m.tasks) + 1,
		ProjectID: projectID,
		Title:     title,
		Desc:      desc,
	}
	m.tasks = append(m.tasks, task)
	return nil
}

func (m *MockService) CreateLog(projectID int, title, desc string) error {
	if m.err != nil {
		return m.err
	}
	log := service.Log{
		ID:        len(m.logs) + 1,
		ProjectID: projectID,
		Title:     title,
		Desc:      desc,
	}
	m.logs = append(m.logs, log)
	return nil
}

func TestNewCoreModel(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, err := NewCoreModel(mockService)
	if err != nil {
		t.Fatalf("NewCoreModel failed: %v", err)
	}
	if len(coreModel.GetProjects()) != 1 {
		t.Errorf("expected 1 project, got %d", len(coreModel.GetProjects()))
	}
}

func TestSelectProject(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		tasks:    []service.Task{{ID: 1, Title: "Test Task"}},
		logs:     []service.Log{{ID: 1, Title: "Test Log"}},
	}
	coreModel, _ := NewCoreModel(mockService)

	cmd := coreModel.SelectProject(0)
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if coreModel.GetSelectedProject().Name != "Test Project" {
		t.Errorf("expected selected project to be 'Test Project', got %s", coreModel.GetSelectedProject().Name)
	}
	if len(coreModel.GetTasks()) != 1 {
		t.Errorf("expected 1 task, got %d", len(coreModel.GetTasks()))
	}
	if len(coreModel.GetLogs()) != 1 {
		t.Errorf("expected 1 log, got %d", len(coreModel.GetLogs()))
	}
}

func TestCreateProject(t *testing.T) {
	mockService := &MockService{}
	coreModel, _ := NewCoreModel(mockService)

	formData := ProjectFormData{
		Name:    "New Project",
		Summary: "Summary",
		Desc:    "Description",
		Status:  "todo",
	}
	cmd := coreModel.CreateProject(formData)

	if cmd != CoreRefreshProjects {
		t.Errorf("expected CoreRefreshProjects, got %v", cmd)
	}
	if coreModel.GetState() != listView {
		t.Errorf("expected state to be listView, got %v", coreModel.GetState())
	}
	if len(mockService.projects) != 1 {
		t.Errorf("expected 1 project in mock service, got %d", len(mockService.projects))
	}
	if mockService.projects[0].Name != "New Project" {
		t.Errorf("expected project name to be 'New Project', got %s", mockService.projects[0].Name)
	}
}

func TestUpdateProject(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	formData := ProjectFormData{
		Name:    "Updated Project",
		Summary: "Updated Summary",
		Desc:    "Updated Description",
		Status:  "in progress",
	}
	cmd := coreModel.UpdateProject(formData)

	if cmd != CoreRefreshProjects {
		t.Errorf("expected CoreRefreshProjects, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if mockService.projects[0].Name != "Updated Project" {
		t.Errorf("expected project name to be 'Updated Project', got %s", mockService.projects[0].Name)
	}
}

func TestDeleteProject(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, _ := NewCoreModel(mockService)

	cmd := coreModel.DeleteProject(0)

	if cmd != CoreRefreshProjects {
		t.Errorf("expected CoreRefreshProjects, got %v", cmd)
	}
	if coreModel.GetState() != listView {
		t.Errorf("expected state to be listView, got %v", coreModel.GetState())
	}
	if len(mockService.projects) != 0 {
		t.Errorf("expected 0 projects in mock service, got %d", len(mockService.projects))
	}
}

func TestCreateTask(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	formData := TaskFormData{
		Title: "New Task",
		Desc:  "Description",
	}
	cmd := coreModel.CreateTask(formData)

	if cmd != CoreRefreshProjectView {
		t.Errorf("expected CoreRefreshProjectView, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if len(mockService.tasks) != 1 {
		t.Errorf("expected 1 task in mock service, got %d", len(mockService.tasks))
	}
	if mockService.tasks[0].Title != "New Task" {
		t.Errorf("expected task title to be 'New Task', got %s", mockService.tasks[0].Title)
	}
}

func TestCreateLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	formData := LogFormData{
		Title: "New Log",
		Desc:  "Description",
	}
	cmd := coreModel.CreateLog(formData)

	if cmd != CoreRefreshProjectView {
		t.Errorf("expected CoreRefreshProjectView, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if len(mockService.logs) != 1 {
		t.Errorf("expected 1 log in mock service, got %d", len(mockService.logs))
	}
	if mockService.logs[0].Title != "New Log" {
		t.Errorf("expected log title to be 'New Log', got %s", mockService.logs[0].Title)
	}
}