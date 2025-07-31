package ui

import (
	"errors"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quamejnr/addae/internal/service"
)

// CoreCommand represents commands that can be sent between layers
type CoreCommand int

const (
	NoCoreCmd CoreCommand = iota
	CoreRefreshProjects
	CoreRefreshProjectView
	CoreQuit
	CoreShowError
)

// viewState represents the current view of the application.
type viewState int

const (
	listView viewState = iota
	projectView
	updateView
	createView
	deleteView
	createTaskView
	createLogView
	deleteTaskView
	fullscreenLogEditView
	updateLogView
	deleteLogView
)

// detailTab represents the active tab in the detail view.
type detailTab int

const (
	projectDetailTab detailTab = iota
	tasksTab
	logsTab
)

// focusState represents the focus state of a view.
type focusState int

const (
	focusList focusState = iota
	focusForm
)

// dialogType represents the type of dialog to display.
type dialogType int

const (
	noDialog dialogType = iota
	projectDeleteDialog
	taskDeleteDialog
	logDeleteDialog
)

// taskDetailMode represents the mode of the task detail view.
type taskDetailMode int

const (
	taskDetailNone taskDetailMode = iota
	taskDetailReadonly
	taskDetailEdit
)

// logDetailMode represents the mode of the log detail view.
type logDetailMode int

const (
	logDetailNone logDetailMode = iota
	logDetailReadonly
)

// CoreModel handles all business logic without UI concerns
type CoreModel struct {
	service         Service
	state           viewState
	selectedProject *service.Project
	selectedTask    *service.Task
	selectedLog     *service.Log
	projects        []service.Project
	tasks           []service.Task
	logs            []service.Log
	err             error
}

// ProjectFormData represents the data structure for project forms
type ProjectFormData struct {
	Name    string
	Summary string
	Desc    string
	Status  string
}

// TaskFormData represents the data structure for task forms
type TaskFormData struct {
	Title string
	Desc  string
}

// LogFormData represents the data structure for log forms
type LogFormData struct {
	Title string
	Desc  string
}

// NewCoreModel creates a new business logic model
func NewCoreModel(svc Service) (*CoreModel, error) {
	projects, err := svc.ListProjects()
	if err != nil {
		return nil, err
	}

	return &CoreModel{
		service:  svc,
		state:    listView,
		projects: projects,
	}, nil
}

// GetProjects returns the current projects list
func (m *CoreModel) GetProjects() []service.Project {
	return m.projects
}

// GetSelectedProject returns the currently selected project
func (m *CoreModel) GetSelectedProject() *service.Project {
	return m.selectedProject
}

// GetSelectedTask returns the currently selected task
func (m *CoreModel) GetSelectedTask() *service.Task {
	return m.selectedTask
}

// GetSelectedLog returns the currently selected log
func (m *CoreModel) GetSelectedLog() *service.Log {
	return m.selectedLog
}

// GetTasks returns tasks for the selected project
func (m *CoreModel) GetTasks() []service.Task {
	return m.tasks
}

// GetLogs returns logs for the selected project
func (m *CoreModel) GetLogs() []service.Log {
	return m.logs
}

// GetError returns the current error
func (m *CoreModel) GetError() error {
	return m.err
}

// GetState returns the current view state
func (m *CoreModel) GetState() viewState {
	return m.state
}

// RefreshProjects reloads the projects list
func (m *CoreModel) RefreshProjects() error {
	projects, err := m.service.ListProjects()
	if err != nil {
		m.err = err
		return err
	}
	m.projects = projects
	m.err = nil
	return nil
}

// SelectProject selects a project by index and loads its related data
func (m *CoreModel) SelectProject(index int) CoreCommand {
	if index < 0 || index >= len(m.projects) {
		m.err = errors.New("invalid project index")
		return CoreShowError
	}

	project := m.projects[index]
	m.selectedProject = &project
	m.state = projectView

	// Load tasks and logs
	tasks, err := m.service.ListProjectTasks(project.ID)
	if err != nil {
		m.err = err
		return CoreShowError
	}
	m.tasks = tasks

	logs, err := m.service.ListProjectLogs(project.ID)
	if err != nil {
		m.err = err
		return CoreShowError
	}
	m.logs = logs

	return NoCoreCmd
}

// GoToCreateView switches to create view
func (m *CoreModel) GoToCreateView() CoreCommand {
	m.state = createView
	return NoCoreCmd
}

// GoToUpdateView switches to update view
func (m *CoreModel) GoToUpdateView() CoreCommand {
	if m.selectedProject == nil {
		m.err = errors.New("no project selected")
		return CoreShowError
	}
	m.state = updateView
	return NoCoreCmd
}

// GoToDeleteView switches to delete view
func (m *CoreModel) GoToDeleteView() CoreCommand {
	m.state = deleteView
	return NoCoreCmd
}

// GoToListView switches to list view
func (m *CoreModel) GoToListView() CoreCommand {
	m.state = listView
	return NoCoreCmd
}

// GoToProjectView switches to project view
func (m *CoreModel) GoToProjectView() CoreCommand {
	if m.selectedProject == nil {
		m.err = errors.New("no project selected")
		return CoreShowError
	}
	m.state = projectView
	return NoCoreCmd
}

// GoToCreateTaskView switches to create task view
func (m *CoreModel) GoToCreateTaskView() CoreCommand {
	m.state = createTaskView
	return NoCoreCmd
}

// GoToCreateLogView switches to create log view
func (m *CoreModel) GoToCreateLogView() CoreCommand {
	m.state = createLogView
	return NoCoreCmd
}

// GoToUpdateLogView switches to update log view
func (m *CoreModel) GoToUpdateLogView() CoreCommand {
	if m.selectedLog == nil {
		m.err = errors.New("no log selected")
		return CoreShowError
	}
	m.state = updateLogView
	return NoCoreCmd
}

// GoToDeleteLogView switches to delete log view
func (m *CoreModel) GoToDeleteLogView() CoreCommand {
	m.state = deleteLogView
	return NoCoreCmd
}

// CreateProject creates a new project
func (m *CoreModel) CreateProject(data ProjectFormData) CoreCommand {
	p := service.Project{
		Name:    data.Name,
		Summary: data.Summary,
		Desc:    data.Desc,
		Status:  data.Status,
	}

	if err := m.service.CreateProject(&p); err != nil {
		m.err = err
		return CoreShowError
	}

	m.state = listView
	return CoreRefreshProjects
}

// UpdateProject updates the selected project
func (m *CoreModel) UpdateProject(data ProjectFormData) CoreCommand {
	if m.selectedProject == nil {
		m.err = errors.New("no project selected")
		return CoreShowError
	}

	p := *m.selectedProject
	p.Name = data.Name
	p.Summary = data.Summary
	p.Desc = data.Desc
	p.Status = data.Status

	if err := m.service.UpdateProject(&p); err != nil {
		m.err = err
		return CoreShowError
	}

	m.selectedProject = &p
	m.state = projectView
	return CoreRefreshProjects
}

// CreateTask creates a new task for the selected project
func (m *CoreModel) CreateTask(data TaskFormData) CoreCommand {
	if m.selectedProject == nil {
		m.err = errors.New("no project selected")
		return CoreShowError
	}

	if err := m.service.CreateTask(m.selectedProject.ID, data.Title, data.Desc); err != nil {
		m.err = err
		return CoreShowError
	}

	m.state = projectView
	return CoreRefreshProjectView
}

// CreateLog creates a new log for the selected project
func (m *CoreModel) CreateLog(data LogFormData) CoreCommand {
	if m.selectedProject == nil {
		m.err = errors.New("no project selected")
		return CoreShowError
	}

	if err := m.service.CreateLog(m.selectedProject.ID, data.Title, data.Desc); err != nil {
		m.err = err
		return CoreShowError
	}

	m.state = projectView
	return CoreRefreshProjectView
}

// UpdateLog updates the selected log
func (m *CoreModel) UpdateLog(data LogFormData) CoreCommand {
	if m.selectedLog == nil {
		m.err = errors.New("no log selected")
		return CoreShowError
	}

	if err := m.service.UpdateLog(m.selectedLog.ID, data.Title, data.Desc); err != nil {
		m.err = err
		return CoreShowError
	}

	// Update the log in memory
	m.selectedLog.Title = data.Title
	m.selectedLog.Desc = data.Desc

	// Update the log in the logs slice
	for i, log := range m.logs {
		if log.ID == m.selectedLog.ID {
			m.logs[i] = *m.selectedLog
			break
		}
	}

	m.state = projectView
	return CoreRefreshProjectView
}

// DeleteLog deletes a log by its ID
func (m *CoreModel) DeleteLog(logID int) CoreCommand {
	if err := m.service.DeleteLog(logID); err != nil {
		m.err = err
		return CoreShowError
	}

	// Clear selected log since it's been deleted
	m.selectedLog = nil

	// Refresh logs for the current project
	logs, err := m.service.ListProjectLogs(m.selectedProject.ID)
	if err != nil {
		m.err = err
		return CoreShowError
	}
	m.logs = logs

	m.state = projectView
	return CoreRefreshProjectView
}

// ConfirmDeleteSelectedLog deletes the selected log if confirmed
func (m *CoreModel) ConfirmDeleteSelectedLog(confirmed bool) CoreCommand {
	if !confirmed {
		m.state = projectView
		return NoCoreCmd
	}

	if m.selectedLog == nil {
		m.err = errors.New("no log selected")
		return CoreShowError
	}

	if err := m.service.DeleteLog(m.selectedLog.ID); err != nil {
		m.err = err
		return CoreShowError
	}

	// Clear selected log since it's been deleted
	m.selectedLog = nil

	// Refresh logs for the current project
	logs, err := m.service.ListProjectLogs(m.selectedProject.ID)
	if err != nil {
		m.err = err
		return CoreShowError
	}
	m.logs = logs

	m.state = projectView
	return CoreRefreshProjectView
}

// ToggleTaskCompletion toggles the completion status of a task
func (m *CoreModel) ToggleTaskCompletion(taskID int, completedAt *time.Time) CoreCommand {
	if m.selectedProject == nil {
		m.err = errors.New("no project selected")
		return CoreShowError
	}

	// Find the task in the current list to get its title and description
	var taskToUpdate service.Task
	found := false
	for _, t := range m.tasks {
		if t.ID == taskID {
			taskToUpdate = t
			found = true
			break
		}
	}

	if !found {
		m.err = errors.New("task not found")
		return CoreShowError
	}

	if err := m.service.UpdateTask(taskID, taskToUpdate.Title, taskToUpdate.Desc, completedAt); err != nil {
		m.err = err
		return CoreShowError
	}

	// Refresh tasks for the current project
	tasks, err := m.service.ListProjectTasks(m.selectedProject.ID)
	if err != nil {
		m.err = err
		return CoreShowError
	}
	m.tasks = tasks

	return CoreRefreshProjectView
}

// SelectTask selects a task by index
func (m *CoreModel) SelectTask(index int) CoreCommand {
	if index < 0 || index >= len(m.tasks) {
		m.err = errors.New("invalid task index")
		return CoreShowError
	}

	task := m.tasks[index]
	m.selectedTask = &task
	m.state = projectView

	return NoCoreCmd
}

// SelectLog selects a log by index
func (m *CoreModel) SelectLog(index int) CoreCommand {
	if index < 0 || index >= len(m.logs) {
		m.err = errors.New("invalid log index")
		return CoreShowError
	}

	log := m.logs[index]
	m.selectedLog = &log
	m.state = projectView

	return NoCoreCmd
}

// GoToEditTaskView switches to edit task view

// DeleteProject deletes a project by index
func (m *CoreModel) DeleteProject(index int) CoreCommand {
	if index < 0 || index >= len(m.projects) {
		m.err = errors.New("invalid project index")
		return CoreShowError
	}

	project := m.projects[index]
	if err := m.service.DeleteProject(project.ID); err != nil {
		m.err = err
		return CoreShowError
	}
	m.state = listView
	return CoreRefreshProjects
}

// ConfirmDeleteSelected deletes the selected project if confirmed
func (m *CoreModel) ConfirmDeleteSelected(confirmed bool) CoreCommand {
	if !confirmed {
		return NoCoreCmd
	}

	if m.selectedProject == nil {
		m.err = errors.New("no project selected")
		return CoreShowError
	}

	if err := m.service.DeleteProject(m.selectedProject.ID); err != nil {
		m.err = err
		return CoreShowError
	}

	m.state = listView
	m.selectedProject = nil
	return CoreRefreshProjects
}

// GoToDeleteTaskView switches to delete task view
func (m *CoreModel) GoToDeleteTaskView() CoreCommand {
	m.state = deleteTaskView
	return NoCoreCmd
}

// DeleteTask deletes a task by its ID
func (m *CoreModel) DeleteTask(taskID int) CoreCommand {
	if err := m.service.DeleteTask(taskID); err != nil {
		m.err = err
		return CoreShowError
	}

	// Refresh tasks for the current project
	tasks, err := m.service.ListProjectTasks(m.selectedProject.ID)
	if err != nil {
		m.err = err
		return CoreShowError
	}
	m.tasks = tasks

	return CoreRefreshProjectView
}

func (m *CoreModel) RefreshProjectViewCmd() tea.Cmd {
	return func() tea.Msg {
		return CoreRefreshProjectView
	}
}