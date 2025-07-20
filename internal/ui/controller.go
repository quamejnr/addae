package ui

import (
	"errors"

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

// CoreModel handles all business logic without UI concerns
type CoreModel struct {
	service         Service
	state           viewState
	selectedProject *service.Project
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
	Title  string
	Desc   string
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
