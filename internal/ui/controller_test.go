package ui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

func (m *MockService) UpdateTask(id int, title, desc string, completedAt *time.Time) error {
	if m.err != nil {
		return m.err
	}
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks[i].Title = title
			m.tasks[i].Desc = desc
			m.tasks[i].CompletedAt = completedAt
			return nil
		}
	}
	return errors.New("task not found")
}

func (m *MockService) DeleteTask(id int) error {
	if m.err != nil {
		return m.err
	}
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return errors.New("task not found")
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

func TestSelectTask(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		tasks:    []service.Task{{ID: 1, ProjectID: 1, Title: "Test Task", Desc: "Test Desc"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0) // Select a project to load tasks

	cmd := coreModel.SelectTask(0)
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if coreModel.GetState() != taskSplitView {
		t.Errorf("expected state to be taskSplitView, got %v", coreModel.GetState())
	}
	if coreModel.GetSelectedTask().Title != "Test Task" {
		t.Errorf("expected selected task to be 'Test Task', got %s", coreModel.GetSelectedTask().Title)
	}
}

func TestGoToTaskDetailView(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		tasks:    []service.Task{{ID: 1, ProjectID: 1, Title: "Test Task", Desc: "Test Desc"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectTask(0)

	cmd := coreModel.GoToTaskSplitView()
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if coreModel.GetState() != taskSplitView {
		t.Errorf("expected state to be taskSplitView, got %v", coreModel.GetState())
	}
}



func TestEditTask(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		tasks:    []service.Task{{ID: 1, ProjectID: 1, Title: "Original Title", Desc: "Original Desc"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectTask(0)

	updatedTitle := "Updated Title"
	updatedDesc := "Updated Desc"

	cmd := coreModel.service.UpdateTask(coreModel.GetSelectedTask().ID, updatedTitle, updatedDesc, coreModel.GetSelectedTask().CompletedAt)

	if cmd != nil { // UpdateTask returns an error, not a CoreCommand
		t.Errorf("expected no error, got %v", cmd)
	}

	// Manually update the selected task in the coreModel to reflect the service change
	coreModel.GetSelectedTask().Title = updatedTitle
	coreModel.GetSelectedTask().Desc = updatedDesc

	// After updating, the state should remain in taskDetailView or return to projectView
	// depending on the UI flow. For this test, we're just verifying the service call.
	// The UI state transition is handled by handleFormCompletion in ui.go, which is not part of this unit test.

	if mockService.tasks[0].Title != "Updated Title" {
		t.Errorf("expected task title to be 'Updated Title', got %s", mockService.tasks[0].Title)
	}
	if mockService.tasks[0].Desc != "Updated Desc" {
		t.Errorf("expected task description to be 'Updated Desc', got %s", mockService.tasks[0].Desc)
	}
}

func TestHandleFormAbort(t *testing.T) {
	// Test aborting from create view
	mockServiceCreate := &MockService{}
	model, _ := NewModel(mockServiceCreate)
	model.CoreModel.GoToCreateView()
	model.handleFormAbort("create")
	if model.GetState() != listView {
		t.Errorf("expected state to be listView after aborting create, got %v", model.GetState())
	}

	// Test aborting from update view
	mockServiceUpdate := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, _ = NewModel(mockServiceUpdate)
	model.CoreModel.SelectProject(0)
	model.CoreModel.GoToUpdateView()
	model.handleFormAbort("update")
	if model.GetState() != projectView {
		t.Errorf("expected state to be projectView after aborting update, got %v", model.GetState())
	}

	// Test aborting from createTask view
	mockServiceCreateTask := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, _ = NewModel(mockServiceCreateTask)
	model.CoreModel.SelectProject(0)
	model.CoreModel.GoToCreateTaskView()
	model.handleFormAbort("createTask")
	if model.GetState() != projectView {
		t.Errorf("expected state to be projectView after aborting createTask, got %v", model.GetState())
	}

	// Test aborting from createLog view
	mockServiceCreateLog := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, _ = NewModel(mockServiceCreateLog)
	model.CoreModel.SelectProject(0)
	model.CoreModel.GoToCreateLogView()
	model.handleFormAbort("createLog")
	if model.GetState() != projectView {
		t.Errorf("expected state to be projectView after aborting createLog, got %v", model.GetState())
	}
}

func TestUpdateProjectView(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Select a project to enter project view
	model.CoreModel.SelectProject(0)

	// Test navigating to the update view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	if model.GetState() != updateView {
		t.Errorf("expected state to be updateView, got %v", model.GetState())
	}
	if model.form == nil {
		t.Error("expected form to be initialized, but it was nil")
	}

	// Test navigating to the create task view
	model.CoreModel.GoToProjectView() // Go back to project view first
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	newModel, _ = model.Update(msg)
	model = newModel.(*Model)

	if model.GetState() != createTaskView {
		t.Errorf("expected state to be createTaskView, got %v", model.GetState())
	}
	if model.form == nil {
		t.Error("expected form to be initialized, but it was nil")
	}
}

func TestToggleTaskCompletion(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		tasks:    []service.Task{{ID: 1, ProjectID: 1, Title: "Test Task", Desc: "Desc"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	// Mark task as complete
	now := time.Now()
	cmd := coreModel.ToggleTaskCompletion(1, &now)
	if cmd != CoreRefreshProjectView {
		t.Errorf("expected CoreRefreshProjectView, got %v", cmd)
	}
	if mockService.tasks[0].CompletedAt == nil {
		t.Errorf("expected task to be completed, but CompletedAt is nil")
	}

	// Mark task as incomplete
	cmd = coreModel.ToggleTaskCompletion(1, nil)
	if cmd != CoreRefreshProjectView {
		t.Errorf("expected CoreRefreshProjectView, got %v", cmd)
	}
	if mockService.tasks[0].CompletedAt != nil {
		t.Errorf("expected task to be incomplete, but CompletedAt is not nil")
	}
}
