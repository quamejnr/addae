package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/quamejnr/addae/internal/service"
)

type Service interface {
	ListProjects() ([]service.Project, error)
	DeleteProject(id int) error
	CreateProject(*service.Project) error
	UpdateProject(*service.Project) error
	ListProjectTasks(projectID int) ([]service.Task, error)
	ListProjectLogs(projectID int) ([]service.Log, error)
	CreateTask(projectID int, title, desc string) error
	UpdateTask(id int, title, desc string, completedAt *time.Time) error
	CreateLog(projectID int, title, desc string) error
	DeleteTask(id int) error
}

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
)

type detailTab int

const (
	projectDetailTab detailTab = iota
	tasksTab
	logsTab
)

type focusState int

const (
	focusList focusState = iota
	focusForm
)

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
	taskViewFocus     focusState
	taskDetailMode    taskDetailMode
	taskEditForm      *TaskEditForm
	quickTaskInput    textinput.Model
	quickInputActive  bool
	showCompleted     bool
}

type taskDetailMode int

const (
	taskDetailNone taskDetailMode = iota
	taskDetailReadonly
	taskDetailEdit
)

type ProjectKeyMap struct {
	UpdateProject   key.Binding
	CreateTask      key.Binding
	CreateLog       key.Binding
	GotoDetails     key.Binding
	GotoTasks       key.Binding
	GotoLogs        key.Binding
	TabLeft         key.Binding
	TabRight        key.Binding
	Back            key.Binding
	Help            key.Binding
	ToggleDone      key.Binding
	DeleteTask      key.Binding
	SelectTask      key.Binding
	CursorUp        key.Binding
	CursorDown      key.Binding
	ExitEdit        key.Binding
	SwitchFocus     key.Binding
	ToggleCompleted key.Binding
}

func (k ProjectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.UpdateProject, k.CreateTask, k.CreateLog, k.Back, k.Help, k.SwitchFocus}
}

func (k ProjectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.UpdateProject, k.CreateTask, k.CreateLog},
		{k.GotoDetails, k.GotoTasks, k.GotoLogs, k.TabLeft, k.TabRight},
		{k.Back, k.Help, k.SwitchFocus},
		{k.ToggleDone, k.DeleteTask, k.SelectTask, k.CursorUp, k.CursorDown, k.ExitEdit},
	}
}

var projectKeys = ProjectKeyMap{
	UpdateProject: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "update project"),
	),
	CreateTask: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "create task"),
	),
	CreateLog: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "create log"),
	),
	GotoDetails: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "show details"),
	),
	GotoTasks: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "show tasks"),
	),
	GotoLogs: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "show logs"),
	),
	TabLeft: key.NewBinding(
		key.WithKeys("left", "ctrl+h"),
		key.WithHelp("←/ctrl+h", "previous tab"),
	),
	TabRight: key.NewBinding(
		key.WithKeys("right", "ctrl+l"),
		key.WithHelp("→/ctrl+l", "next tab"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "b", "ctrl+c"),
		key.WithHelp("esc/b/ctrl+c", "back to list"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	ToggleDone: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle done"),
	),
	DeleteTask: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete task"),
	),
	SelectTask: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select task"),
	),
	CursorUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	CursorDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	ExitEdit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "exit edit"),
	),
	SwitchFocus: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	),
	ToggleCompleted: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "toggle completed"),
	),
}

var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	leftColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("240")).
			PaddingRight(1)

	rightColumnStyle = lipgloss.NewStyle().
				PaddingLeft(1)

	detailTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
				MarginBottom(1)

	detailSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39")).
				MarginTop(1).
				MarginBottom(1)

	detailItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	emptyDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true).
				MarginTop(5)

	taskEditFormStyle = lipgloss.NewStyle().
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 2)

	subStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#6A6A6A",
	})
)

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

	return &Model{
		CoreModel:         coreModel,
		list:              projectList,
		width:             80,
		height:            24,
		activeTab:         projectDetailTab,
		help:              help.New(),
		keys:              projectKeys,
		selectedTaskIndex: 0,
		quickTaskInput:    quickInput,
		quickInputActive:  false,
		showCompleted:     false,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listWidth := m.width/2 - 4
		listHeight := m.height - 4
		m.list.SetSize(listWidth, listHeight)
	}

	switch m.GetState() {
	case listView:
		return m.updateListView(msg)
	case projectView:
		model, cmd := m.updateProjectViewCommon(msg)
		if cmd != nil {
			return model, cmd
		}

		switch m.activeTab {
		case tasksTab:
			// Handled by updateProjectViewCommon
		case projectDetailTab:
		case logsTab:
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

	}

	return m, cmd
}

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
				if selected := m.list.Index(); selected >= 0 {
					m.CoreModel.GoToDeleteView()
					m.form = confirmDeleteForm()
					return m, m.form.Init()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	return m, cmd
}

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
		// Handle ONLY the back key for taskDetailReadonly mode FIRST
		if m.activeTab == tasksTab && m.taskDetailMode == taskDetailReadonly {
			if key.Matches(msg, m.keys.Back) {
				m.CoreModel.selectedTask = nil
				m.taskDetailMode = taskDetailNone
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
				m.CoreModel.GoToCreateTaskView()
				m.form = createTaskForm()
				return m, m.form.Init()
			case key.Matches(msg, m.keys.CreateLog):
				m.CoreModel.GoToCreateLogView()
				m.form = createLogForm()
				return m, m.form.Init()
			case msg.String() == "n" && m.activeTab == tasksTab:
				m.quickInputActive = true
				m.quickTaskInput.Focus()
				return m, textinput.Blink
			case key.Matches(msg, m.keys.ToggleCompleted) && m.activeTab == tasksTab:
				m.showCompleted = !m.showCompleted
				if !m.showCompleted && m.selectedTaskIndex > m.getMaxNavigableIndex() {
					m.selectedTaskIndex = m.getMaxNavigableIndex()
				}
				return m, nil
			case key.Matches(msg, m.keys.GotoDetails):
				m.activeTab = projectDetailTab
				m.taskDetailMode = taskDetailNone // Reset task detail mode when switching tabs
				m.CoreModel.selectedTask = nil
			case key.Matches(msg, m.keys.GotoTasks):
				m.activeTab = tasksTab
				m.taskDetailMode = taskDetailNone // Reset task detail mode when switching to tasks tab
				m.CoreModel.selectedTask = nil
			case key.Matches(msg, m.keys.GotoLogs):
				m.activeTab = logsTab
				m.taskDetailMode = taskDetailNone // Reset task detail mode when switching tabs
				m.CoreModel.selectedTask = nil
			case key.Matches(msg, m.keys.TabRight):
				m.activeTab = (m.activeTab + 1) % 3
				m.taskDetailMode = taskDetailNone // Reset task detail mode when switching tabs
				m.CoreModel.selectedTask = nil
			case key.Matches(msg, m.keys.TabLeft):
				m.activeTab = (m.activeTab - 1 + 3) % 3
				m.taskDetailMode = taskDetailNone // Reset task detail mode when switching tabs
				m.CoreModel.selectedTask = nil
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
				// If no task is selected in detail view, delegate all keys to updateTasksList
				// This will handle CursorUp/Down and Enter to select a task
				model, cmd := m.updateTasksList(msg)
				// If a task was selected by updateTasksList, transition to readonly mode (unless already in edit mode)
				if m.CoreModel.GetSelectedTask() != nil && m.taskDetailMode != taskDetailEdit {
					m.taskDetailMode = taskDetailReadonly
				}
				return model, cmd
			case taskDetailReadonly:
				switch {
				case key.Matches(msg, m.keys.ExitEdit):
					m.taskDetailMode = taskDetailEdit
					if task := m.CoreModel.GetSelectedTask(); task != nil {
						m.taskEditForm = newTaskEditForm(*task)
						return m, m.taskEditForm.Init()
					}
				case key.Matches(msg, m.keys.CursorUp):
					if m.selectedTaskIndex > 0 {
						m.selectedTaskIndex--
					}
					coreCmd := m.CoreModel.SelectTask(m.selectedTaskIndex)
					if coreCmd == CoreShowError {
						return m, nil
					}
				case key.Matches(msg, m.keys.CursorDown):
					if m.selectedTaskIndex < m.getMaxNavigableIndex() {
						m.selectedTaskIndex++
					}
					coreCmd := m.CoreModel.SelectTask(m.selectedTaskIndex)
					if coreCmd == CoreShowError {
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
							maxIndex := m.getMaxNavigableIndex()
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
	}
	return m, cmd
}

func (m *Model) updateTasksList(msg tea.Msg) (tea.Model, tea.Cmd) {
	tasks := m.CoreModel.GetTasks()
	if len(tasks) == 0 {
		return m, nil
	}

	// Calculate max navigable index
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.CursorUp):
			if m.selectedTaskIndex > 0 {
				m.selectedTaskIndex--
			}
		case key.Matches(msg, m.keys.CursorDown):
			if m.selectedTaskIndex < maxIndex {
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
			maxIndex := m.getMaxNavigableIndex()
			if m.selectedTaskIndex > maxIndex {
				m.selectedTaskIndex = maxIndex
			}
		case key.Matches(msg, m.keys.SelectTask):
			coreCmd := m.CoreModel.SelectTask(m.selectedTaskIndex)
			if coreCmd == CoreShowError {
				return m, nil
			}
			m.taskDetailMode = taskDetailReadonly
			return m, nil
		case key.Matches(msg, m.keys.ExitEdit): // 'e' key for direct edit
			coreCmd := m.CoreModel.SelectTask(m.selectedTaskIndex)
			if coreCmd == CoreShowError {
				return m, nil
			}
			m.taskDetailMode = taskDetailEdit // Go directly to edit mode
			if task := m.CoreModel.GetSelectedTask(); task != nil {
				m.taskEditForm = newTaskEditForm(*task)
				return m, m.taskEditForm.Init()
			}
		case key.Matches(msg, m.keys.DeleteTask):
			m.CoreModel.GoToDeleteTaskView()
			m.form = confirmDeleteTaskForm()
			return m, m.form.Init()
		case key.Matches(msg, m.keys.Back):
			m.CoreModel.GoToListView()
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) updateTaskForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedForm, cmd := m.form.Update(msg)
	m.form = updatedForm.(*huh.Form)

	if m.form.State == huh.StateAborted {
		m.form = nil
		m.taskDetailMode = taskDetailReadonly
		return m, nil
	} else if m.form.State == huh.StateCompleted {
		data := TaskFormData{
			Title: m.form.GetString("title"),
			Desc:  m.form.GetString("desc"),
		}
		task := m.CoreModel.GetSelectedTask()
		if task != nil {
			if err := m.CoreModel.service.UpdateTask(task.ID, data.Title, data.Desc, task.CompletedAt); err != nil {
				m.CoreModel.err = err
				return m, nil
			}
			task.Title = data.Title
			task.Desc = data.Desc

			// Update the task in the tasks slice directly
			for i, t := range m.CoreModel.tasks {
				if t.ID == task.ID {
					m.CoreModel.tasks[i] = *task
					break
				}
			}

		}
		m.form = nil
		m.taskDetailMode = taskDetailReadonly
		return m, nil
	}
	return m, cmd
}

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
	}
}

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
	case "createLog":
		data := LogFormData{
			Title: m.form.GetString("title"),
			Desc:  m.form.GetString("desc"),
		}
		return m.CoreModel.CreateLog(data)
	case "deleteTask":
		confirmed := m.form.GetBool("confirm")
		if confirmed {
			tasks := m.CoreModel.GetTasks()
			if m.selectedTaskIndex >= 0 && m.selectedTaskIndex < len(tasks) {
				task := tasks[m.selectedTaskIndex]
				cmd := m.CoreModel.DeleteTask(task.ID)
				if cmd == CoreShowError {
					return CoreShowError
				}
			}
		}
		m.CoreModel.GoToProjectView()
		m.activeTab = tasksTab
		return CoreRefreshProjectView
	}
	return NoCoreCmd
}

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
}

func (m Model) View() string {
	if err := m.GetError(); err != nil {
		return "Error: " + err.Error()
	}

	var mainContent string

	switch m.GetState() {
	case listView:
		mainContent = m.renderTabularView()
	case projectView:
		mainContent = m.renderDetailPanel()

	case updateView, createView, deleteView, createTaskView, createLogView, deleteTaskView:
		mainContent = m.form.View()
	}

	return mainContent
}

func (m *Model) renderTabularView() string {
	leftWidth := m.width/2 - 4
	rightWidth := m.width/2 - 4
	panelHeight := m.height - 4 // Consistent with app margins

	leftColumn := leftColumnStyle.
		Width(leftWidth).
		Height(panelHeight).
		Render(m.list.View())

	rightColumn := rightColumnStyle.
		Width(rightWidth).
		Height(panelHeight).
		Render(m.renderDetailPanel())

	return appStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn),
	)
}

func (m *Model) renderDetailPanel() string {
	project := m.GetSelectedProject()
	if project == nil {
		return emptyDetailStyle.Render("← Select a project to view details")
	}

	// Calculate consistent content height (total height - app margins - tabs - help)
	contentHeight := m.height - 6 // 2 for app margins + 2 for tabs + 2 for help spacing

	// Handle tasks split view separately to avoid superimposition
	if m.activeTab == tasksTab && m.taskDetailMode != taskDetailNone {
		var s strings.Builder
		s.WriteString(m.renderTabs())
		s.WriteString("\n")
		s.WriteString(m.renderTasksSplitView(contentHeight))
		s.WriteString("\n")
		s.WriteString(m.help.View(m.keys))
		return s.String()
	}

	// Handle all other views normally
	var s strings.Builder
	s.WriteString(m.renderTabs())
	s.WriteString("\n")

	// Render content with consistent height
	var content string
	switch m.activeTab {
	case projectDetailTab:
		content = m.renderProjectDetails()
	case tasksTab:
		// This will only be reached when taskDetailMode == taskDetailNone
		content = m.renderTasksListOnly()
	case logsTab:
		content = m.renderLogsList()
	}

	// Ensure content takes up the available space consistently
	contentStyle := lipgloss.NewStyle().Height(contentHeight - 2) // -2 for tab spacing
	s.WriteString(contentStyle.Render(content))
	s.WriteString("\n")
	s.WriteString(m.help.View(m.keys))
	return s.String()
}

func (m *Model) renderTasksSplitView(totalHeight int) string {
	leftWidth := m.width/2 - 4
	rightWidth := m.width/2 - 4
	splitHeight := totalHeight - 2

	var taskListContent strings.Builder

	if m.quickInputActive {
		taskListContent.WriteString(m.quickTaskInput.View() + "\n\n")
	}

	tasks := m.CoreModel.GetTasks()
	if len(tasks) == 0 {
		taskListContent.WriteString(emptyDetailStyle.Render("No tasks for this project."))
	} else {
		// Separate tasks
		var pending, completed []service.Task
		for _, t := range tasks {
			if t.CompletedAt != nil {
				completed = append(completed, t)
			} else {
				pending = append(pending, t)
			}
		}

		// Render pending tasks
		for i, t := range pending {
			taskLine := "[ ] " + t.Title
			if i == m.selectedTaskIndex {
				taskLine = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
					Render(taskLine)
			}
			taskListContent.WriteString(detailItemStyle.Render(taskLine))
			taskListContent.WriteString("\n")
		}

		// Completed section
		if len(completed) > 0 {
			taskListContent.WriteString("\n")

			toggle := "▶"
			if m.showCompleted {
				toggle = "▼"
			}

			separator := fmt.Sprintf("%s Completed (%d)", toggle, len(completed))
			taskListContent.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render(separator))
			taskListContent.WriteString("\n")

			// Show completed tasks if expanded
			if m.showCompleted {
				for i, t := range completed {
					taskLine := "[x] " + t.Title
					completedIndex := len(pending) + i
					if completedIndex == m.selectedTaskIndex {
						taskLine = lipgloss.NewStyle().
							Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
							Render(taskLine)
					}
					taskListContent.WriteString(lipgloss.NewStyle().
						Foreground(lipgloss.Color("240")).
						Render(taskLine))
					taskListContent.WriteString("\n")
				}
			}
		}
	}

	leftColumn := leftColumnStyle.
		Width(leftWidth).
		Height(splitHeight).
		Render(taskListContent.String())

	rightColumn := rightColumnStyle.
		Width(rightWidth).
		Height(splitHeight).
		Render(m.renderTasksListPanel())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}

func (m *Model) renderTasksListPanel() string {
	var s strings.Builder

	switch m.taskDetailMode {
	case taskDetailNone:
		return emptyDetailStyle.Render("← Select a task to view details")
	case taskDetailReadonly:
		s.WriteString(m.renderTaskReadonlyView())
	case taskDetailEdit:
		s.WriteString(m.renderTaskForm())
	}

	return s.String()
}

func (m *Model) renderTaskReadonlyView() string {
	var s strings.Builder

	task := m.getVisualTask(m.selectedTaskIndex)
	if task == nil {
		return emptyDetailStyle.Render("No task selected")
	}

	s.WriteString(detailTitleStyle.Render(task.Title))
	s.WriteString("\n")
	s.WriteString(task.Desc)
	s.WriteString("\n\n")

	// Show completion status
	if task.CompletedAt != nil {
		s.WriteString(subStyle.Render("Status: Completed"))
		s.WriteString("\n")
		s.WriteString(subStyle.Render("Completed at: " + task.CompletedAt.Format("2006-01-02 15:04")))
	} else {
		s.WriteString(subStyle.Render("Status: Pending"))
	}

	return s.String()
}

func (m *Model) renderTaskForm() string {
	if m.taskEditForm != nil {
		return m.taskEditForm.View()
	}
	return "No form"
}

func (m *Model) renderTabs() string {
	tabStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		MarginBottom(1)

	activeTabStyle := tabStyle.
		BorderForeground(lipgloss.Color("69")).
		Foreground(lipgloss.Color("69"))

	var tabs []string

	tabs = append(tabs, m.tabTitle("Details", projectDetailTab, tabStyle, activeTabStyle))
	tabs = append(tabs, m.tabTitle("Tasks", tasksTab, tabStyle, activeTabStyle))
	tabs = append(tabs, m.tabTitle("Logs", logsTab, tabStyle, activeTabStyle))

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m *Model) tabTitle(title string, tab detailTab, style, activeStyle lipgloss.Style) string {
	if m.activeTab == tab {
		return activeStyle.Render(title)
	}
	return style.Render(title)
}

func (m *Model) renderProjectDetails() string {
	project := m.GetSelectedProject()
	var s strings.Builder

	s.WriteString(detailTitleStyle.Render(project.Name))
	s.WriteString("\n")
	s.WriteString(detailItemStyle.Render("Status: " + project.Status))
	s.WriteString("\n")
	if project.Summary != "" {
		s.WriteString(detailItemStyle.Render("Summary: " + project.Summary))
		s.WriteString("\n")
	}
	if project.Desc != "" {
		s.WriteString(detailItemStyle.Render("Description: " + project.Desc))
		s.WriteString("\n")
	}
	return s.String()
}

func (m *Model) renderTasksListOnly() string {
	var s strings.Builder

	if m.quickInputActive {
		s.WriteString(m.quickTaskInput.View() + "\n\n")
	}

	tasks := m.GetTasks()
	if len(tasks) == 0 {
		s.WriteString(detailItemStyle.Render("No tasks for this project."))
		return s.String()
	}

	// Separate tasks
	var pending, completed []service.Task
	for _, t := range tasks {
		if t.CompletedAt != nil {
			completed = append(completed, t)
		} else {
			pending = append(pending, t)
		}
	}

	// Render pending tasks
	for i, t := range pending {
		taskLine := "[ ] " + t.Title
		if i == m.selectedTaskIndex {
			taskLine = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
				Render(taskLine)
		}
		s.WriteString(detailItemStyle.Render(taskLine))
		s.WriteString("\n")
	}

	// Completed section
	if len(completed) > 0 {
		s.WriteString("\n")

		toggle := "▶"
		if m.showCompleted {
			toggle = "▼"
		}

		separator := fmt.Sprintf("%s Completed (%d)", toggle, len(completed))
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(separator))
		s.WriteString("\n")

		// Show completed tasks if expanded
		if m.showCompleted {
			for i, t := range completed {
				taskLine := "[x] " + t.Title
				completedIndex := len(pending) + i
				if completedIndex == m.selectedTaskIndex {
					taskLine = lipgloss.NewStyle().
						Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
						Render(taskLine)
				}
				s.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Render(taskLine))
				s.WriteString("\n")
			}
		}
	}

	return s.String()
}

func (m *Model) renderLogsList() string {
	logs := m.GetLogs()
	var s strings.Builder

	if len(logs) == 0 {
		s.WriteString(detailItemStyle.Render("No logs for this project."))
		s.WriteString("\n")
	} else {
		for _, l := range logs {
			s.WriteString(detailItemStyle.Render("• " + l.Title))
			s.WriteString("\n")
		}
	}
	return s.String()
}

func (m *Model) getMaxNavigableIndex() int {
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

func (m *Model) getVisualTask(index int) *service.Task {
	tasks := m.CoreModel.GetTasks()

	// Separate into pending and completed
	var pending, completed []service.Task
	for _, t := range tasks {
		if t.CompletedAt == nil {
			pending = append(pending, t)
		} else {
			completed = append(completed, t)
		}
	}

	// Map visual index to actual task
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

func confirmDeleteForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to delete this project?").
				Key("confirm"),
		),
	)
}

func confirmDeleteTaskForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to delete this task?").
				Key("confirm"),
		),
	)
}

func updateProjectForm(p service.Project) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Key("name").
				Value(&p.Name),
			huh.NewText().
				Title("Summary").
				Key("summary").
				CharLimit(255).
				Value(&p.Summary),
			huh.NewText().
				Title("Description").
				Key("desc").
				Value(&p.Desc),
			huh.NewSelect[string]().
				Title("Status").
				Key("status").
				Options(
					huh.NewOption("Todo", "todo"),
					huh.NewOption("In Progress", "in progress"),
					huh.NewOption("Completed", "completed"),
					huh.NewOption("Archived", "archived"),
				).
				Value(&p.Status),
		),
	)
}

func createProjectForm() *huh.Form {
	defaultValue := "todo"
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Key("name").
				Placeholder("Enter name of project"),
			huh.NewText().
				CharLimit(255).
				Title("Summary").
				Key("summary").
				Placeholder("Enter overview of project"),
			huh.NewText().
				Title("Description (Optional)").
				Key("desc").
				Placeholder("Enter detailed description of project"),
			huh.NewSelect[string]().Title("Status").
				Key("status").
				Options(
					huh.NewOption("Todo", "todo"),
					huh.NewOption("In Progress", "in progress"),
					huh.NewOption("Done", "completed"),
					huh.NewOption("Archived", "archived"),
				).
				Value(&defaultValue),
		),
	)
}

func createTaskForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Key("title").
				Placeholder("Enter title of task"),
			huh.NewText().
				Title("Description (Optional)").
				Key("desc").
				Placeholder("Enter detailed description of task"),
		),
	)
}

func createLogForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Key("title").
				Placeholder("Enter title of log"),
			huh.NewText().
				Title("Description (Optional)").
				Key("desc").
				Placeholder("Enter detailed description of log"),
		),
	)
}

func updateTaskForm(t service.Task) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Key("title").
				Value(&t.Title),
			huh.NewText().
				Title("Description").
				Key("desc").
				Value(&t.Desc),
		),
	)
}

type TaskEditForm struct {
	titleInput textinput.Model
	descInput  textarea.Model
	focusIndex int
	completed  bool
	aborted    bool
}

func newTaskEditForm(task service.Task) *TaskEditForm {
	titleInput := textinput.New()
	titleInput.SetValue(task.Title)
	titleInput.Focus()
	titleInput.Width = 50

	descInput := textarea.New()
	descInput.SetValue(task.Desc)
	descInput.SetWidth(70)
	descInput.SetHeight(5)
	descInput.FocusedStyle.CursorLine = lipgloss.NewStyle()
	descInput.ShowLineNumbers = false

	return &TaskEditForm{
		titleInput: titleInput,
		descInput:  descInput,
		focusIndex: 0,
	}
}

func (f *TaskEditForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f *TaskEditForm) Update(msg tea.Msg) (*TaskEditForm, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			f.aborted = true
			return f, nil
		case "enter":
			if f.focusIndex == 0 {
				f.focusIndex = 1
				f.titleInput.Blur()
				f.descInput.Focus()
				return f, textinput.Blink
			} else {
				f.completed = true
				return f, nil
			}
		case "tab":
			if f.focusIndex == 0 {
				f.focusIndex = 1
				f.titleInput.Blur()
				f.descInput.Focus()
			} else {
				f.focusIndex = 0
				f.descInput.Blur()
				f.titleInput.Focus()
			}
			return f, textinput.Blink
		}
	}

	var cmd tea.Cmd
	if f.focusIndex == 0 {
		f.titleInput, cmd = f.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		f.descInput, cmd = f.descInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return f, tea.Batch(cmds...)
}

func (f *TaskEditForm) View() string {
	var s strings.Builder

	s.WriteString(detailTitleStyle.Render("Edit Task"))
	// s.WriteString("\n")

	// s.WriteString(detailSectionStyle.Render("Title:"))
	s.WriteString("\n")
	s.WriteString(f.titleInput.View())
	s.WriteString("\n")

	// s.WriteString(detailSectionStyle.Render("Description:"))
	s.WriteString("\n")
	s.WriteString(f.descInput.View())
	s.WriteString("\n\n\n")

	helpText := subStyle.Render("Tab: next field • Enter: save • Esc: cancel")
	s.WriteString(helpText)

	return taskEditFormStyle.Render(s.String())
}

func (f *TaskEditForm) GetTitle() string {
	return f.titleInput.Value()
}

func (f *TaskEditForm) GetDesc() string {
	return f.descInput.Value()
}

func (f *TaskEditForm) IsCompleted() bool {
	return f.completed
}

func (f *TaskEditForm) IsAborted() bool {
	return f.aborted
}
