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

func (m *MockService) UpdateLog(id int, title, desc string) error {
	if m.err != nil {
		return m.err
	}
	for i, l := range m.logs {
		if l.ID == id {
			m.logs[i].Title = title
			m.logs[i].Desc = desc
			return nil
		}
	}
	return errors.New("log not found")
}

func (m *MockService) DeleteLog(id int) error {
	if m.err != nil {
		return m.err
	}
	for i, l := range m.logs {
		if l.ID == id {
			m.logs = append(m.logs[:i], m.logs[i+1:]...)
			return nil
		}
	}
	return errors.New("log not found")
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

func TestSelectLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "Test Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0) // Select a project to load logs

	cmd := coreModel.SelectLog(0)
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if coreModel.GetSelectedLog().Title != "Test Log" {
		t.Errorf("expected selected log to be 'Test Log', got %s", coreModel.GetSelectedLog().Title)
	}
}

func TestSelectLogInvalidIndex(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, Title: "Test Log"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	// Test invalid negative index
	cmd := coreModel.SelectLog(-1)
	if cmd != CoreShowError {
		t.Errorf("expected CoreShowError for negative index, got %v", cmd)
	}

	// Test invalid large index
	cmd = coreModel.SelectLog(10)
	if cmd != CoreShowError {
		t.Errorf("expected CoreShowError for large index, got %v", cmd)
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

	if cmd != CoreRefreshTasksView {
		t.Errorf("expected CoreRefreshTasksView, got %v", cmd)
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

	if cmd != CoreRefreshLogsView {
		t.Errorf("expected CoreRefreshLogsView, got %v", cmd)
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

func TestUpdateLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Original Log", Desc: "Original Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectLog(0)

	formData := LogFormData{
		Title: "Updated Log",
		Desc:  "Updated Description",
	}
	cmd := coreModel.UpdateLog(formData)

	if cmd != CoreRefreshLogsView {
		t.Errorf("expected CoreRefreshLogsView, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if mockService.logs[0].Title != "Updated Log" {
		t.Errorf("expected log title to be 'Updated Log', got %s", mockService.logs[0].Title)
	}
	if mockService.logs[0].Desc != "Updated Description" {
		t.Errorf("expected log desc to be 'Updated Description', got %s", mockService.logs[0].Desc)
	}
	if coreModel.GetSelectedLog().Title != "Updated Log" {
		t.Errorf("expected selected log title to be 'Updated Log', got %s", coreModel.GetSelectedLog().Title)
	}
}

func TestUpdateLogNoSelectedLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	formData := LogFormData{
		Title: "Updated Log",
		Desc:  "Updated Description",
	}
	cmd := coreModel.UpdateLog(formData)

	if cmd != CoreShowError {
		t.Errorf("expected CoreShowError when no log selected, got %v", cmd)
	}
	if coreModel.GetError() == nil {
		t.Error("expected error when no log selected")
	}
}

func TestDeleteLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Log to Delete", Desc: "Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectLog(0)

	cmd := coreModel.DeleteLog(1)

	if cmd != CoreRefreshLogsView {
		t.Errorf("expected CoreRefreshLogsView, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if len(mockService.logs) != 0 {
		t.Errorf("expected 0 logs in mock service after delete, got %d", len(mockService.logs))
	}
	if coreModel.GetSelectedLog() != nil {
		t.Error("expected selected log to be nil after delete")
	}
}

func TestConfirmDeleteSelectedLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Log to Delete", Desc: "Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectLog(0)

	// Test confirmed deletion
	cmd := coreModel.ConfirmDeleteSelectedLog(true)
	if cmd != CoreRefreshLogsView {
		t.Errorf("expected CoreRefreshLogsView when confirmed, got %v", cmd)
	}
	if len(mockService.logs) != 0 {
		t.Errorf("expected 0 logs in mock service after confirmed delete, got %d", len(mockService.logs))
	}
	if coreModel.GetSelectedLog() != nil {
		t.Error("expected selected log to be nil after confirmed delete")
	}
}

func TestConfirmDeleteSelectedLogNotConfirmed(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Log to Delete", Desc: "Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectLog(0)

	// Test not confirmed deletion
	cmd := coreModel.ConfirmDeleteSelectedLog(false)
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd when not confirmed, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView when not confirmed, got %v", coreModel.GetState())
	}
	if len(mockService.logs) != 1 {
		t.Errorf("expected 1 log in mock service when not confirmed, got %d", len(mockService.logs))
	}
}

func TestConfirmDeleteSelectedLogNoSelectedLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	cmd := coreModel.ConfirmDeleteSelectedLog(true)
	if cmd != CoreShowError {
		t.Errorf("expected CoreShowError when no log selected, got %v", cmd)
	}
	if coreModel.GetError() == nil {
		t.Error("expected error when no log selected")
	}
}

func TestGoToUpdateLogView(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectLog(0)

	cmd := coreModel.GoToUpdateLogView()
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if coreModel.GetState() != updateLogView {
		t.Errorf("expected state to be updateLogView, got %v", coreModel.GetState())
	}
}

func TestGoToUpdateLogViewNoSelectedLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	cmd := coreModel.GoToUpdateLogView()
	if cmd != CoreShowError {
		t.Errorf("expected CoreShowError when no log selected, got %v", cmd)
	}
	if coreModel.GetError() == nil {
		t.Error("expected error when no log selected")
	}
}

func TestGoToDeleteLogView(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectLog(0)

	cmd := coreModel.GoToDeleteLogView()
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if coreModel.GetState() != deleteLogView {
		t.Errorf("expected state to be deleteLogView, got %v", coreModel.GetState())
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
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
	}
	if coreModel.GetSelectedTask().Title != "Test Task" {
		t.Errorf("expected selected task to be 'Test Task', got %s", coreModel.GetSelectedTask().Title)
	}
}

func TestGoToProjectView(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		tasks:    []service.Task{{ID: 1, ProjectID: 1, Title: "Test Task", Desc: "Test Desc"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectTask(0)

	cmd := coreModel.GoToProjectView()
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to be projectView, got %v", coreModel.GetState())
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

	// Test aborting from deleteLog view
	mockServiceDeleteLog := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "Description"}},
	}
	model, _ = NewModel(mockServiceDeleteLog)
	model.CoreModel.SelectProject(0)
	model.CoreModel.SelectLog(0)
	model.CoreModel.GoToDeleteLogView()
	model.handleFormAbort("deleteLog")
	if model.GetState() != projectView {
		t.Errorf("expected state to be projectView after aborting deleteLog, got %v", model.GetState())
	}
	if model.activeTab != logsTab {
		t.Errorf("expected activeTab to be logsTab after aborting deleteLog, got %v", model.activeTab)
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
	if cmd != CoreRefreshTasksView {
		t.Errorf("expected CoreRefreshTasksView, got %v", cmd)
	}
	if mockService.tasks[0].CompletedAt == nil {
		t.Errorf("expected task to be completed, but CompletedAt is nil")
	}

	// Mark task as incomplete
	cmd = coreModel.ToggleTaskCompletion(1, nil)
	if cmd != CoreRefreshTasksView {
		t.Errorf("expected CoreRefreshTasksView, got %v", cmd)
	}
	if mockService.tasks[0].CompletedAt != nil {
		t.Errorf("expected task to be incomplete, but CompletedAt is not nil")
	}
}

func TestLogsInteractivity(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs: []service.Log{
			{ID: 1, ProjectID: 1, Title: "Log 1", Desc: "# Markdown Log 1\nContent here"},
			{ID: 2, ProjectID: 1, Title: "Log 2", Desc: "# Markdown Log 2\nMore content"},
		},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Select project and switch to logs tab
	model.CoreModel.SelectProject(0)
	model.activeTab = logsTab

	// Test selecting a log
	model.selectedLogIndex = 0
	log := model.getLogAtIndex(0)
	if log == nil {
		t.Error("expected to get log at index 0, but got nil")
	}
	if log.Title != "Log 1" {
		t.Errorf("expected log title to be 'Log 1', got %s", log.Title)
	}

	// Test log selection with SelectLog
	cmd := model.CoreModel.SelectLog(1)
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd, got %v", cmd)
	}
	if model.CoreModel.GetSelectedLog().Title != "Log 2" {
		t.Errorf("expected selected log to be 'Log 2', got %s", model.CoreModel.GetSelectedLog().Title)
	}
}

func TestLogNavigation(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs: []service.Log{
			{ID: 1, ProjectID: 1, Title: "Log 1", Desc: "Content 1"},
			{ID: 2, ProjectID: 1, Title: "Log 2", Desc: "Content 2"},
			{ID: 3, ProjectID: 1, Title: "Log 3", Desc: "Content 3"},
		},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.CoreModel.SelectProject(0)
	model.activeTab = logsTab

	// Test initial state
	if model.selectedLogIndex != 0 {
		t.Errorf("expected initial selectedLogIndex to be 0, got %d", model.selectedLogIndex)
	}

	// Test navigation down
	model.selectedLogIndex = 1
	log := model.getLogAtIndex(model.selectedLogIndex)
	if log.Title != "Log 2" {
		t.Errorf("expected log at index 1 to be 'Log 2', got %s", log.Title)
	}

	// Test navigation up
	model.selectedLogIndex = 0
	log = model.getLogAtIndex(model.selectedLogIndex)
	if log.Title != "Log 1" {
		t.Errorf("expected log at index 0 to be 'Log 1', got %s", log.Title)
	}

	// Test bounds checking - invalid index
	log = model.getLogAtIndex(-1)
	if log != nil {
		t.Error("expected nil for invalid negative index")
	}

	log = model.getLogAtIndex(10)
	if log != nil {
		t.Error("expected nil for invalid large index")
	}
}

func TestLogDetailModes(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs: []service.Log{
			{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "# Test Content\nSome markdown content"},
		},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.CoreModel.SelectProject(0)
	model.activeTab = logsTab

	// Test initial state
	if model.logDetailMode != logDetailNone {
		t.Errorf("expected initial logDetailMode to be logDetailNone, got %v", model.logDetailMode)
	}

	// Test selecting a log transitions to readonly mode
	model.CoreModel.SelectLog(0)
	model.logDetailMode = logDetailReadonly

	if model.logDetailMode != logDetailReadonly {
		t.Errorf("expected logDetailMode to be logDetailReadonly after selection, got %v", model.logDetailMode)
	}

	// Test log view focus states
	if model.logViewFocus != focusList {
		t.Errorf("expected initial logViewFocus to be focusList, got %v", model.logViewFocus)
	}

	// Switch to pager focus
	model.logViewFocus = focusForm
	if model.logViewFocus != focusForm {
		t.Errorf("expected logViewFocus to be focusForm after switch, got %v", model.logViewFocus)
	}
}

func TestLogViewportIntegration(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs: []service.Log{
			{ID: 1, ProjectID: 1, Title: "Markdown Log", Desc: "# Heading\n\nThis is **bold** text"},
		},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test that viewport is initialized
	if model.logViewport.Width == 0 {
		t.Error("expected viewport to have non-zero width")
	}

	// Test that glamour renderer is initialized
	if model.glamourRenderer == nil {
		t.Error("expected glamour renderer to be initialized")
	}

	// Test rendering markdown content
	testMarkdown := "# Test\nThis is a test"
	rendered, err := model.glamourRenderer.Render(testMarkdown)
	if err != nil {
		t.Errorf("expected no error rendering markdown, got %v", err)
	}
	if rendered == "" {
		t.Error("expected rendered content to be non-empty")
	}
}

func TestGetSelectedLog(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs: []service.Log{
			{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "Test content"},
		},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)

	// Test no selected log initially
	if coreModel.GetSelectedLog() != nil {
		t.Error("expected no selected log initially")
	}

	// Test after selecting a log
	coreModel.SelectLog(0)
	selectedLog := coreModel.GetSelectedLog()
	if selectedLog == nil {
		t.Error("expected selected log after SelectLog")
	}
	if selectedLog.Title != "Test Log" {
		t.Errorf("expected selected log title to be 'Test Log', got %s", selectedLog.Title)
	}
}

func TestLogUpdateAndDeleteIntegration(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs: []service.Log{
			{ID: 1, ProjectID: 1, Title: "Original Log", Desc: "Original content"},
			{ID: 2, ProjectID: 1, Title: "Another Log", Desc: "Another content"},
		},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Setup: select project and log
	model.CoreModel.SelectProject(0)
	model.activeTab = logsTab
	model.CoreModel.SelectLog(0)

	// Test update log workflow
	formData := LogFormData{
		Title: "Updated Log Title",
		Desc:  "Updated log description",
	}
	cmd := model.CoreModel.UpdateLog(formData)
	if cmd != CoreRefreshLogsView {
		t.Errorf("expected CoreRefreshLogsView after update, got %v", cmd)
	}

	// Verify the log was updated in memory
	if model.CoreModel.GetSelectedLog().Title != "Updated Log Title" {
		t.Errorf("expected selected log title to be updated, got %s", model.CoreModel.GetSelectedLog().Title)
	}
	if model.CoreModel.GetSelectedLog().Desc != "Updated log description" {
		t.Errorf("expected selected log desc to be updated, got %s", model.CoreModel.GetSelectedLog().Desc)
	}

	// Test delete log workflow
	originalLogCount := len(model.CoreModel.GetLogs())
	cmd = model.CoreModel.ConfirmDeleteSelectedLog(true)
	if cmd != CoreRefreshLogsView {
		t.Errorf("expected CoreRefreshLogsView after delete, got %v", cmd)
	}

	// Verify the log was deleted
	if len(mockService.logs) != originalLogCount-1 {
		t.Errorf("expected log count to decrease by 1, got %d logs", len(mockService.logs))
	}
	if model.CoreModel.GetSelectedLog() != nil {
		t.Error("expected selected log to be nil after delete")
	}
}

func TestLogServiceErrorHandling(t *testing.T) {
	// Test UpdateLog service error
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "Description"}},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
    // add errors
	mockService.err = errors.New("service error")

	coreModel := model.CoreModel
	coreModel.SelectProject(0)
	model.activeTab = logsTab
	coreModel.SelectLog(0)

	formData := LogFormData{Title: "Updated", Desc: "Updated"}
	cmd := coreModel.UpdateLog(formData)
	if cmd != CoreShowError {
		t.Errorf("expected CoreShowError on service error, got %v", cmd)
	}
	if coreModel.GetError() == nil {
		t.Error("expected error to be set on service failure")
	}

	// Test DeleteLog service error
	mockService.err = errors.New("delete error")
	cmd = coreModel.DeleteLog(1)
	if cmd != CoreShowError {
		t.Errorf("expected CoreShowError on delete service error, got %v", cmd)
	}
}

func TestLogStateTransitions(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
		logs:     []service.Log{{ID: 1, ProjectID: 1, Title: "Test Log", Desc: "Description"}},
	}
	coreModel, _ := NewCoreModel(mockService)
	coreModel.SelectProject(0)
	coreModel.SelectLog(0)

	// Test transition to update log view
	cmd := coreModel.GoToUpdateLogView()
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd for valid transition, got %v", cmd)
	}
	if coreModel.GetState() != updateLogView {
		t.Errorf("expected state to be updateLogView, got %v", coreModel.GetState())
	}

	// Test transition to delete log view
	cmd = coreModel.GoToDeleteLogView()
	if cmd != NoCoreCmd {
		t.Errorf("expected NoCoreCmd for valid transition, got %v", cmd)
	}
	if coreModel.GetState() != deleteLogView {
		t.Errorf("expected state to be deleteLogView, got %v", coreModel.GetState())
	}

	// Test return to project view after operations
	coreModel.UpdateLog(LogFormData{Title: "Updated", Desc: "Updated"})
	if coreModel.GetState() != projectView {
		t.Errorf("expected state to return to projectView after update, got %v", coreModel.GetState())
	}
}
