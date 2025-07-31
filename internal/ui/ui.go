package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/quamejnr/addae/internal/service"
)

// Service defines the interface for interacting with the business logic.
type Service interface {
	ListProjects() ([]service.Project, error)
	DeleteProject(id int) error
	CreateProject(*service.Project) error
	UpdateProject(*service.Project) error
	ListProjectTasks(projectID int) ([]service.Task, error)
	CreateTask(projectID int, title, desc string) error
	UpdateTask(id int, title, desc string, completedAt *time.Time) error
	DeleteTask(id int) error
	ListProjectLogs(projectID int) ([]service.Log, error)
	CreateLog(projectID int, title, desc string) error
	UpdateLog(id int, title, desc string) error
	DeleteLog(id int) error
}

// Model represents the state of the UI.
type Model struct {
	*CoreModel
	list              list.Model
	form              *huh.Form
	width             int
	height            int
	activeTab         detailTab
	help              help.Model
	keys              ProjectKeyMap
	selectedTaskIndex int
	selectedLogIndex  int
	taskViewFocus     focusState
	logViewFocus      focusState
	taskDetailMode    taskDetailMode
	logDetailMode     logDetailMode
	taskEditForm      *TaskEditForm
	quickTaskInput    textinput.Model
	quickInputActive  bool
	showCompleted     bool
	logViewport       viewport.Model
	glamourRenderer   *glamour.TermRenderer
	logEditForm       *LogEditForm

	// State for the custom delete confirmation dialog
	deleteConfirmCursor int // 0 = cancel, 1 = delete
	deleteDialogType    dialogType
	deleteAction        func() CoreCommand
}

// NewModel creates a new UI model.
func NewModel(svc Service) (*Model, error) {
	coreModel, err := NewCoreModel(svc)
	if err != nil {
		return nil, err
	}

	quickInput := textinput.New()
	quickInput.Placeholder = "Add task"
	quickInput.Width = 40

	projects := coreModel.GetProjects()
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = p
	}

	delegate := list.NewDefaultDelegate()
	projectList := list.New(items, delegate, 40, 20)
	projectList.Title = "Addae"
	projectList.SetShowHelp(true)

	// Initialize viewport for log pager
	const width = 78
	vp := viewport.New(width, 20)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.HiddenBorder())
	vp.YPosition = 0

	const glamourGutter = 2
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter
	// Initialize glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dracula"),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return nil, err
	}

	return &Model{
		CoreModel:         coreModel,
		list:              projectList,
		width:             80,
		height:            24,
		activeTab:         projectDetailTab,
		help:              help.New(),
		keys:              projectKeys,
		selectedTaskIndex: 0,
		selectedLogIndex:  0,
		quickTaskInput:    quickInput,
		quickInputActive:  false,
		showCompleted:     false,
		logViewport:       vp,
		glamourRenderer:   renderer,
		logViewFocus:      focusList,
	}, nil
}

// Init initializes the UI model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

// Update handles messages and updates the UI model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle the delete confirmation dialog first if it's visible.
	if m.deleteDialogType != noDialog {
		return m.updateConfirmDeleteDialog(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport size
		rightWidth := m.width/2 - 4
		m.logViewport.Width = rightWidth
		m.logViewport.Height = m.height - 10

		// Update glamour renderer width based on the viewport's new content area
		const glamourGutter = 2
		// The content area is the viewport's total width minus its own border/padding.
		glamourRenderWidth := m.logViewport.Width - m.logViewport.Style.GetHorizontalFrameSize() - glamourGutter
		m.glamourRenderer, _ = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dracula"),
			glamour.WithWordWrap(glamourRenderWidth),
		)

		// Re-render log content if a log is selected
		if m.activeTab == logsTab && m.logDetailMode == logDetailReadonly {
			if log := m.CoreModel.GetSelectedLog(); log != nil {
				rendered, err := m.glamourRenderer.Render(log.Desc)
				if err != nil {
					rendered = log.Desc // fallback to plain text
				}
				m.logViewport.SetContent(rendered)
			}
		}

		// Update log edit form if active
		if m.logEditForm != nil {
			m.logEditForm.width = m.width
			m.logEditForm.height = m.height
			m.logEditForm.textarea.SetWidth(m.width - 6)
			m.logEditForm.textarea.SetHeight(m.height - 6)
		}
	}

	switch m.GetState() {
	case listView:
		return m.updateListView(msg)
	case projectView:
		model, cmd := m.updateProjectViewCommon(msg)
		if cmd != nil {
			return model, cmd
		}
		return m, cmd
	case updateView:
		return m.updateFormView(msg, "update")
	case createView:
		return m.updateFormView(msg, "create")
	case deleteView:
		return m.updateFormView(msg, "delete")
	case createTaskView:
		return m.updateFormView(msg, "createTask")
	case createLogView:
		return m.updateFormView(msg, "createLog")
	case deleteTaskView:
		return m.updateFormView(msg, "deleteTask")
	case fullscreenLogEditView:
		return m.updateFullscreenLogEdit(msg)
	case updateLogView:
		return m.updateFullscreenLogEdit(msg)
	case deleteLogView:
		return m.updateFormView(msg, "deleteLog")
	}

	return m, cmd
}

// updateListView handles updates for the list view.
func (m *Model) updateListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	if m.list.Index() >= 0 {
		m.loadProjectDetails(m.list.Index())
	}

	if m.list.FilterState().String() != "filtering" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if selected := m.list.Index(); selected >= 0 {
					coreCmd := m.CoreModel.SelectProject(selected)
					if coreCmd == CoreShowError {
						return m, nil
					}
					m.CoreModel.GoToProjectView()
				}
			case "n":
				m.CoreModel.GoToCreateView()
				m.form = createProjectForm()
				return m, m.form.Init()
			case "d":
				if selectedIndex := m.list.Index(); selectedIndex >= 0 {
					projects := m.CoreModel.GetProjects()
					if selectedIndex < len(projects) {
						m.deleteDialogType = projectDeleteDialog
						m.deleteConfirmCursor = 0 // Default to Cancel
						m.deleteAction = func() CoreCommand {
							return m.CoreModel.DeleteProject(selectedIndex)
						}
					}
				}
			case "u":
				if selectedIndex := m.list.Index(); selectedIndex >= 0 {
					projects := m.CoreModel.GetProjects()
					if selectedIndex < len(projects) {
						selectedProject := projects[selectedIndex]
						m.CoreModel.selectedProject = &selectedProject
						m.CoreModel.GoToUpdateView()
						m.form = updateProjectForm(selectedProject)
						return m, m.form.Init()
					}
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	return m, cmd
}

// updateFullscreenLogEdit handles updates for the fullscreen log edit view.
func (m *Model) updateFullscreenLogEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.logEditForm == nil {
		m.CoreModel.GoToProjectView()
		return m, nil
	}

	var cmd tea.Cmd
	m.logEditForm, cmd = m.logEditForm.Update(msg)

	if m.logEditForm.IsAborted() {
		m.logEditForm = nil
		m.CoreModel.GoToProjectView()
		m.activeTab = logsTab
		return m, nil
	}

	if m.logEditForm.IsCompleted() {
		title, desc := m.logEditForm.GetContent()
		if title != "" {
			data := LogFormData{
				Title: title,
				Desc:  desc,
			}
			var coreCmd CoreCommand
			if m.GetState() == updateLogView {
				coreCmd = m.CoreModel.UpdateLog(data)
			} else {
				coreCmd = m.CoreModel.CreateLog(data)
			}

			m.logEditForm = nil

			switch coreCmd {
			case CoreRefreshProjectView:
				m.loadProjectDetails(m.list.Index())
				// Re-render log content after editing
				if log := m.CoreModel.GetSelectedLog(); log != nil {
					rendered, err := m.glamourRenderer.Render(log.Desc)
					if err != nil {
						rendered = log.Desc // fallback to plain text
					}
					m.logViewport.SetContent(rendered)
				}
				m.activeTab = logsTab
			case CoreShowError:
				fmt.Printf("Something went wrong %s", m.err.Error())
			}
		}
		m.CoreModel.GoToProjectView()
		m.activeTab = logsTab
		return m, nil
	}

	return m, cmd
}

// updateProjectViewCommon handles common updates for the project view.
func (m *Model) updateProjectViewCommon(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Quick input handling
		if m.activeTab == tasksTab && m.quickInputActive {
			switch msg.String() {
			case "enter":
				title := strings.TrimSpace(m.quickTaskInput.Value())
				if title != "" {
					m.CoreModel.service.CreateTask(m.GetSelectedProject().ID, title, "")
					m.loadProjectDetails(m.list.Index())
					m.quickTaskInput.SetValue("")
				}
				m.quickInputActive = false
				return m, nil
			case "esc":
				m.quickInputActive = false
				m.quickTaskInput.SetValue("")
				return m, nil
			default:
				m.quickTaskInput, cmd = m.quickTaskInput.Update(msg)
				return m, cmd
			}
		}

		// Handle log viewport navigation when in pager focus
		if m.activeTab == logsTab && m.logDetailMode == logDetailReadonly && m.logViewFocus == focusForm {
			switch {
			case key.Matches(msg, m.keys.SwitchFocus):
				m.logViewFocus = focusList
				return m, nil
			case key.Matches(msg, m.keys.Edit):
				if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
					m.CoreModel.selectedLog = log
					m.CoreModel.state = updateLogView
					m.logEditForm = newLogEditFormWithData(m.width, m.height, log.Title, log.Desc)
					return m, m.logEditForm.Init()
				}
			default:
				// Let viewport handle scrolling
				m.logViewport, cmd = m.logViewport.Update(msg)
				return m, cmd
			}
		}

		// Handle ONLY the back key for taskDetailReadonly mode FIRST
		if m.activeTab == tasksTab && m.taskDetailMode == taskDetailReadonly {
			if key.Matches(msg, m.keys.Back) {
				m.CoreModel.selectedTask = nil
				m.taskDetailMode = taskDetailNone
				return m, nil
			}
		}

		// Handle ONLY the back key for logDetailReadonly mode FIRST
		if m.activeTab == logsTab && m.logDetailMode == logDetailReadonly {
			if key.Matches(msg, m.keys.Back) {
				m.CoreModel.selectedLog = nil
				m.logDetailMode = logDetailNone
				m.logViewFocus = focusList
				return m, nil
			}
		}

		// Skip general keys if task edit form is active
		if m.activeTab == tasksTab && m.taskDetailMode == taskDetailEdit && m.taskEditForm != nil {
			// Let the form handle the keys - don't process general keys
		} else {
			// Handle general project view keys
			switch {
			case key.Matches(msg, m.keys.UpdateProject):
				m.CoreModel.GoToUpdateView()
				if project := m.CoreModel.GetSelectedProject(); project != nil {
					m.form = updateProjectForm(*project)
					return m, m.form.Init()
				}
			case key.Matches(msg, m.keys.CreateTask):
				m.activeTab = tasksTab
				m.quickInputActive = true
				m.quickTaskInput.Focus()
				return m, textinput.Blink
			case key.Matches(msg, m.keys.CreateLog):
				m.CoreModel.state = fullscreenLogEditView
				m.logEditForm = newLogEditForm(m.width, m.height)
				return m, m.logEditForm.Init()
			case key.Matches(msg, m.keys.CreateObject) && m.activeTab == tasksTab:
				m.quickInputActive = true
				m.quickTaskInput.Focus()
				return m, textinput.Blink
			case key.Matches(msg, m.keys.ToggleCompleted) && m.activeTab == tasksTab:
				m.showCompleted = !m.showCompleted
				if !m.showCompleted && m.selectedTaskIndex > m.getMaxNavigableTaskIndex() {
					m.selectedTaskIndex = m.getMaxNavigableTaskIndex()
				}
				return m, nil
			case key.Matches(msg, m.keys.CreateLog):
				m.CoreModel.state = fullscreenLogEditView
				m.logEditForm = newLogEditForm(m.width, m.height)
				return m, m.logEditForm.Init()
			case key.Matches(msg, m.keys.CreateObject) && m.activeTab == logsTab:
				m.CoreModel.state = fullscreenLogEditView
				m.logEditForm = newLogEditForm(m.width, m.height)
				return m, m.logEditForm.Init()
			case key.Matches(msg, m.keys.GotoDetails):
				m.activeTab = projectDetailTab
				m.taskDetailMode = taskDetailNone
				m.logDetailMode = logDetailNone
				m.CoreModel.selectedTask = nil
				m.CoreModel.selectedLog = nil
			case key.Matches(msg, m.keys.GotoTasks):
				m.activeTab = tasksTab
				m.taskDetailMode = taskDetailNone
				m.logDetailMode = logDetailNone
				m.CoreModel.selectedTask = nil
				m.CoreModel.selectedLog = nil
			case key.Matches(msg, m.keys.GotoLogs):
				m.activeTab = logsTab
				m.taskDetailMode = taskDetailNone
				m.logDetailMode = logDetailNone
				m.CoreModel.selectedTask = nil
				m.CoreModel.selectedLog = nil
				m.logViewFocus = focusList
			case key.Matches(msg, m.keys.TabRight):
				m.activeTab = (m.activeTab + 1) % 3
				m.taskDetailMode = taskDetailNone
				m.logDetailMode = logDetailNone
				m.CoreModel.selectedTask = nil
				m.CoreModel.selectedLog = nil
				m.logViewFocus = focusList
			case key.Matches(msg, m.keys.TabLeft):
				m.activeTab = (m.activeTab - 1 + 3) % 3
				m.taskDetailMode = taskDetailNone
				m.logDetailMode = logDetailNone
				m.CoreModel.selectedTask = nil
				m.CoreModel.selectedLog = nil
				m.logViewFocus = focusList
			case key.Matches(msg, m.keys.Back):
				m.CoreModel.GoToListView()
				return m, nil
			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll
				return m, nil
			}
		}

		// Handle task-specific keybindings AFTER general navigation
		if m.activeTab == tasksTab {
			switch m.taskDetailMode {
			case taskDetailNone:
				model, cmd := m.updateTasksList(msg)
				if m.CoreModel.GetSelectedTask() != nil && m.taskDetailMode != taskDetailEdit {
					m.taskDetailMode = taskDetailReadonly
				}
				return model, cmd
			case taskDetailReadonly:
				switch {
				case key.Matches(msg, m.keys.Edit):
					m.taskDetailMode = taskDetailEdit
					if task := m.CoreModel.GetSelectedTask(); task != nil {
						m.taskEditForm = newTaskEditForm(*task)
						return m, m.taskEditForm.Init()
					}
				case key.Matches(msg, m.keys.CursorUp):
					if m.selectedTaskIndex > 0 {
						m.selectedTaskIndex--
					}
					if task := m.getVisualTask(m.selectedTaskIndex); task != nil {
						m.CoreModel.selectedTask = task
					}
				case key.Matches(msg, m.keys.CursorDown):
					if m.selectedTaskIndex < m.getMaxNavigableTaskIndex() {
						m.selectedTaskIndex++
					}
					if task := m.getVisualTask(m.selectedTaskIndex); task != nil {
						m.CoreModel.selectedTask = task
					}
				case key.Matches(msg, m.keys.DeleteObject):
					if task := m.getVisualTask(m.selectedTaskIndex); task != nil {
						m.deleteDialogType = taskDeleteDialog
						m.deleteConfirmCursor = 0
						m.deleteAction = func() CoreCommand {
							return m.CoreModel.DeleteTask(task.ID)
						}
						return m, nil
					}
				case key.Matches(msg, m.keys.ToggleDone):
					tasks := m.CoreModel.GetTasks()
					if m.selectedTaskIndex >= 0 && m.selectedTaskIndex < len(tasks) {
						task := m.getVisualTask(m.selectedTaskIndex)
						var completedAt *time.Time
						if task.CompletedAt == nil {
							now := time.Now()
							completedAt = &now
						}
						cmd := m.CoreModel.ToggleTaskCompletion(task.ID, completedAt)
						if cmd == CoreShowError {
							return m, nil
						}
						m.CoreModel.SelectProject(m.list.Index())
						if task.CompletedAt == nil && completedAt != nil {
							if !m.showCompleted && m.selectedTaskIndex > 0 {
								m.selectedTaskIndex--
							}
							maxIndex := m.getMaxNavigableTaskIndex()
							if m.selectedTaskIndex > maxIndex {
								m.selectedTaskIndex = maxIndex
							}
						}
					}
				}
			case taskDetailEdit:
				if m.taskEditForm != nil {
					var cmd tea.Cmd
					m.taskEditForm, cmd = m.taskEditForm.Update(msg)

					if m.taskEditForm.IsAborted() {
						m.taskEditForm = nil
						m.taskDetailMode = taskDetailReadonly
						return m, nil
					}

					if m.taskEditForm.IsCompleted() {
						task := m.CoreModel.GetSelectedTask()
						if task != nil {
							title := m.taskEditForm.GetTitle()
							desc := m.taskEditForm.GetDesc()

							if err := m.CoreModel.service.UpdateTask(task.ID, title, desc, task.CompletedAt); err != nil {
								m.CoreModel.err = err
								return m, nil
							}

							task.Title = title
							task.Desc = desc

							for i, t := range m.CoreModel.tasks {
								if t.ID == task.ID {
									m.CoreModel.tasks[i] = *task
									break
								}
							}
						}
						m.taskEditForm = nil
						m.taskDetailMode = taskDetailReadonly
						return m, nil
					}

					return m, cmd
				}
			}
		}

		// Handle log-specific keybindings AFTER general navigation
		if m.activeTab == logsTab {
			switch m.logDetailMode {
			case logDetailNone:
				model, cmd := m.updateLogsList(msg)
				if m.CoreModel.GetSelectedLog() != nil {
					m.logDetailMode = logDetailReadonly
					// Update viewport with rendered markdown
					if log := m.CoreModel.GetSelectedLog(); log != nil {
						rendered, err := m.glamourRenderer.Render(log.Desc)
						if err != nil {
							rendered = log.Desc // fallback to plain text
						}
						m.logViewport.SetContent(rendered)
					}
				}
				return model, cmd
			case logDetailReadonly:
				// Handle focus switching
				if m.logViewFocus == focusList {
					switch {
					case key.Matches(msg, m.keys.CursorUp):
						if m.selectedLogIndex > 0 {
							m.selectedLogIndex--
							if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
								m.CoreModel.selectedLog = log
								// Update viewport content
								rendered, err := m.glamourRenderer.Render(log.Desc)
								if err != nil {
									rendered = log.Desc
								}
								m.logViewport.SetContent(rendered)
								m.logViewport.GotoTop()
							}
						}
					case key.Matches(msg, m.keys.CursorDown):
						logs := m.CoreModel.GetLogs()
						if m.selectedLogIndex < len(logs)-1 {
							m.selectedLogIndex++
							if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
								m.CoreModel.selectedLog = log
								// Update viewport content
								rendered, err := m.glamourRenderer.Render(log.Desc)
								if err != nil {
									rendered = log.Desc
								}
								m.logViewport.SetContent(rendered)
								m.logViewport.GotoTop()
							}
						}
					case key.Matches(msg, m.keys.SwitchFocus):
						m.logViewFocus = focusForm
						return m, nil

					case key.Matches(msg, m.keys.CreateObject):
						m.CoreModel.state = fullscreenLogEditView
						m.logEditForm = newLogEditForm(m.width, m.height)
						return m, m.logEditForm.Init()

					case key.Matches(msg, m.keys.Edit):
						if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
							m.CoreModel.selectedLog = log
							m.CoreModel.state = updateLogView
							m.logEditForm = newLogEditFormWithData(m.width, m.height, log.Title, log.Desc)
							return m, m.logEditForm.Init()
						}
					case key.Matches(msg, m.keys.DeleteObject):
						if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
							m.deleteDialogType = logDeleteDialog
							m.deleteConfirmCursor = 0 // Default to Cancel
							m.deleteAction = func() CoreCommand {
								return m.CoreModel.DeleteLog(log.ID)
							}
							return m, nil
						}
					}
				}
			}
		}
	}
	return m, cmd
}

// updateTasksList handles updates for the tasks list.
func (m *Model) updateTasksList(msg tea.Msg) (tea.Model, tea.Cmd) {
	tasks := m.CoreModel.GetTasks()
	if len(tasks) == 0 {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.CursorUp):
			if m.selectedTaskIndex > 0 {
				m.selectedTaskIndex--
			}
		case key.Matches(msg, m.keys.CursorDown):
			if m.selectedTaskIndex < m.getMaxNavigableTaskIndex() {
				m.selectedTaskIndex++
			}
		case key.Matches(msg, m.keys.ToggleDone):
			task := m.getVisualTask(m.selectedTaskIndex)
			var completedAt *time.Time
			if task.CompletedAt == nil {
				now := time.Now()
				completedAt = &now
			}
			cmd := m.CoreModel.ToggleTaskCompletion(task.ID, completedAt)
			if cmd == CoreShowError {
				return m, nil
			}
			m.CoreModel.SelectProject(m.list.Index())
			if task.CompletedAt == nil && completedAt != nil {
				if !m.showCompleted && m.selectedTaskIndex > 0 {
					m.selectedTaskIndex--
				}
			}
			maxIndex := m.getMaxNavigableTaskIndex()
			if m.selectedTaskIndex > maxIndex {
				m.selectedTaskIndex = maxIndex
			}
		case key.Matches(msg, m.keys.DeleteObject):
			if task := m.getVisualTask(m.selectedTaskIndex); task != nil {
				m.CoreModel.selectedTask = task
				m.deleteDialogType = taskDeleteDialog
				m.deleteConfirmCursor = 0 // Default to Cancel
				m.deleteAction = func() CoreCommand {
					return m.CoreModel.DeleteTask(task.ID)
				}
			}
		case key.Matches(msg, m.keys.SelectObject):
			if task := m.getVisualTask(m.selectedTaskIndex); task != nil {
				m.CoreModel.selectedTask = task
				m.taskDetailMode = taskDetailReadonly
			}
			return m, nil

		case key.Matches(msg, m.keys.Edit):
			if task := m.getVisualTask(m.selectedTaskIndex); task != nil {
				m.CoreModel.selectedTask = task
			}
			m.taskDetailMode = taskDetailEdit
			if task := m.CoreModel.GetSelectedTask(); task != nil {
				m.taskEditForm = newTaskEditForm(*task)
				return m, m.taskEditForm.Init()
			}
		case key.Matches(msg, m.keys.Back):
			m.CoreModel.GoToListView()
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	}
	return m, nil
}

// updateLogsList handles updates for the logs list.
func (m *Model) updateLogsList(msg tea.Msg) (tea.Model, tea.Cmd) {
	logs := m.CoreModel.GetLogs()
	if len(logs) == 0 {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.CursorUp):
			if m.selectedLogIndex > 0 {
				m.selectedLogIndex--
			}
		case key.Matches(msg, m.keys.CursorDown):
			if m.selectedLogIndex < len(logs)-1 {
				m.selectedLogIndex++
			}
		case key.Matches(msg, m.keys.SelectObject):
			if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
				m.CoreModel.selectedLog = log
				m.logDetailMode = logDetailReadonly
				m.logViewFocus = focusList
			}
			return m, nil
		case key.Matches(msg, m.keys.Edit):
			if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
				m.CoreModel.selectedLog = log
				m.CoreModel.state = updateLogView
				m.logEditForm = newLogEditFormWithData(m.width, m.height, log.Title, log.Desc)
				return m, m.logEditForm.Init()
			}
		case key.Matches(msg, m.keys.DeleteObject):
			if log := m.getLogAtIndex(m.selectedLogIndex); log != nil {
				m.CoreModel.selectedLog = log
				m.deleteDialogType = logDeleteDialog
				m.deleteConfirmCursor = 0
				m.deleteAction = func() CoreCommand {
					return m.CoreModel.DeleteLog(log.ID)
				}
				return m, nil
			}
		case key.Matches(msg, m.keys.CreateObject):
			m.CoreModel.state = fullscreenLogEditView
			m.logEditForm = newLogEditForm(m.width, m.height)
			return m, m.logEditForm.Init()
		case key.Matches(msg, m.keys.Back):
			m.CoreModel.GoToListView()
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	}
	return m, nil
}

// updateFormView handles updates for the form view.
func (m *Model) updateFormView(msg tea.Msg, formType string) (tea.Model, tea.Cmd) {
	var formCmd tea.Cmd
	var updatedForm tea.Model
	updatedForm, formCmd = m.form.Update(msg)
	m.form = updatedForm.(*huh.Form)

	if m.form.State == huh.StateAborted {
		m.handleFormAbort(formType)
		m.form = nil
		return m, nil
	} else if m.form.State == huh.StateCompleted {
		coreCmd := m.handleFormCompletion(formType)
		m.form = nil

		switch coreCmd {
		case CoreRefreshProjects:
			if err := m.CoreModel.RefreshProjects(); err == nil {
				m.refreshListItems()
			}
		case CoreRefreshProjectView:
			m.loadProjectDetails(m.list.Index())
		case CoreQuit:
			return m, tea.Quit
		}
		return m, nil
	}

	return m, formCmd
}

// handleFormAbort handles the aborting of a form.
func (m *Model) handleFormAbort(formType string) {
	switch formType {
	case "create":
		m.CoreModel.GoToListView()
	case "update":
		m.CoreModel.GoToProjectView()
	case "delete":
		m.CoreModel.GoToListView()
	case "createTask", "createLog", "deleteTask":
		m.CoreModel.GoToProjectView()
		m.activeTab = tasksTab
	case "deleteLog":
		m.CoreModel.GoToProjectView()
		m.activeTab = logsTab
	}
}

// handleFormCompletion handles the completion of a form.
func (m *Model) handleFormCompletion(formType string) CoreCommand {
	switch formType {
	case "create":
		data := ProjectFormData{
			Name:    m.form.GetString("name"),
			Summary: m.form.GetString("summary"),
			Desc:    m.form.GetString("desc"),
			Status:  m.form.GetString("status"),
		}
		return m.CoreModel.CreateProject(data)
	case "update":
		data := ProjectFormData{
			Name:    m.form.GetString("name"),
			Summary: m.form.GetString("summary"),
			Desc:    m.form.GetString("desc"),
			Status:  m.form.GetString("status"),
		}
		return m.CoreModel.UpdateProject(data)
	case "delete":
		confirmed := m.form.GetBool("confirm")
		if confirmed {
			if selected := m.list.Index(); selected >= 0 {
				return m.CoreModel.DeleteProject(selected)
			}
		}
		m.CoreModel.GoToListView()
		return NoCoreCmd
	case "createTask":
		data := TaskFormData{
			Title: m.form.GetString("title"),
			Desc:  m.form.GetString("desc"),
		}
		return m.CoreModel.CreateTask(data)
	case "deleteTask":
		confirmed := m.form.GetBool("confirm")
		if confirmed {
			if task := m.getVisualTask(m.selectedTaskIndex); task != nil {
				cmd := m.CoreModel.DeleteTask(task.ID)
				if cmd == CoreShowError {
					return CoreShowError
				}
			}
		}
		m.CoreModel.GoToProjectView()
		m.activeTab = tasksTab
		return CoreRefreshProjectView
	case "deleteLog":
		confirmed := m.form.GetBool("confirm")
		if confirmed {
			if log := m.CoreModel.GetSelectedLog(); log != nil {
				cmd := m.CoreModel.ConfirmDeleteSelectedLog(true)
				if cmd == CoreShowError {
					return CoreShowError
				}
				m.activeTab = logsTab
				return CoreRefreshProjectView
			}
		}
		m.CoreModel.GoToProjectView()
		m.activeTab = logsTab
		return NoCoreCmd
	}

	return NoCoreCmd
}

// updateConfirmDeleteDialog handles updates for the delete confirmation dialog.
func (m *Model) updateConfirmDeleteDialog(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.deleteDialogType == noDialog {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.deleteDialogType = noDialog
			return m, nil

		case "left", "h", "tab":
			m.deleteConfirmCursor = (m.deleteConfirmCursor + 1) % 2

		case "right", "l", "shift+tab":
			m.deleteConfirmCursor = (m.deleteConfirmCursor + 1) % 2

		case "enter":
			m.deleteDialogType = noDialog
			if m.deleteConfirmCursor == 1 { // 1 is Delete
				coreCmd := m.deleteAction()
				switch coreCmd {
				case CoreRefreshProjects:
					if err := m.CoreModel.RefreshProjects(); err == nil {
						m.refreshListItems()
					}
				case CoreRefreshProjectView:
					m.loadProjectDetails(m.list.Index())
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// refreshListItems refreshes the list of projects.
func (m *Model) refreshListItems() {
	projects := m.CoreModel.GetProjects()
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = p
	}
	m.list.SetItems(items)

	if len(projects) == 0 {
		m.CoreModel.selectedProject = nil
		m.CoreModel.tasks = nil
		m.CoreModel.logs = nil
		return
	}
	idx := m.list.Index()
	if idx >= len(projects) {
		idx = len(projects) - 1
		m.list.Select(idx)
	}
	m.loadProjectDetails(idx)
}

// loadProjectDetails loads the details of a project.
func (m *Model) loadProjectDetails(index int) {
	if index < 0 || index >= len(m.CoreModel.GetProjects()) {
		return
	}

	projects := m.CoreModel.GetProjects()
	project := projects[index]
	m.CoreModel.selectedProject = &project

	if tasks, err := m.CoreModel.service.ListProjectTasks(project.ID); err == nil {
		m.CoreModel.tasks = tasks
	}

	if logs, err := m.CoreModel.service.ListProjectLogs(project.ID); err == nil {
		m.CoreModel.logs = logs
	}

	// Reset cursor positions for tasks and logs to prevent out-of-bounds errors
	// when I leave a project and move to another the selectedindex still persist
	// This can lead to cursor being out of position for some tasks and logs
	m.selectedTaskIndex = 0
	m.selectedLogIndex = 0
}

// getMaxNavigableTaskIndex returns the maximum navigable task index.
func (m *Model) getMaxNavigableTaskIndex() int {
	tasks := m.CoreModel.GetTasks()
	var pending []service.Task
	for _, t := range tasks {
		if t.CompletedAt == nil {
			pending = append(pending, t)
		}
	}

	maxIndex := len(pending) - 1
	if m.showCompleted {
		maxIndex = len(tasks) - 1
	}
	return maxIndex
}

// getVisualTask returns the visual task at the given index.
func (m *Model) getVisualTask(index int) *service.Task {
	tasks := m.CoreModel.GetTasks()

	var pending, completed []service.Task
	for _, t := range tasks {
		if t.CompletedAt == nil {
			pending = append(pending, t)
		} else {
			completed = append(completed, t)
		}
	}

	if index < len(pending) {
		return &pending[index]
	}

	if m.showCompleted {
		completedIndex := index - len(pending)
		if completedIndex >= 0 && completedIndex < len(completed) {
			return &completed[completedIndex]
		}
	}

	return nil
}

// getLogAtIndex returns the log at the given index.
func (m *Model) getLogAtIndex(index int) *service.Log {
	logs := m.CoreModel.GetLogs()
	if index >= 0 && index < len(logs) {
		return &logs[index]
	}
	return nil
}

