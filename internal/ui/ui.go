package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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
}

type taskDetailMode int

const (
	taskDetailNone taskDetailMode = iota
	taskDetailReadonly
	taskDetailEdit
)

type ProjectKeyMap struct {
	UpdateProject key.Binding
	CreateTask    key.Binding
	CreateLog     key.Binding
	GotoDetails   key.Binding
	GotoTasks     key.Binding
	GotoLogs      key.Binding
	TabLeft       key.Binding
	TabRight      key.Binding
	Back          key.Binding
	Help          key.Binding
	ToggleDone    key.Binding
	DeleteTask    key.Binding
	SelectTask    key.Binding
	CursorUp      key.Binding
	CursorDown    key.Binding
	ExitEdit      key.Binding
	SwitchFocus   key.Binding
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
)

func NewModel(svc Service) (*Model, error) {
	coreModel, err := NewCoreModel(svc)
	if err != nil {
		return nil, err
	}

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
		},
		nil
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
			model, cmd = m.updateTasksList(msg)
			if cmd != nil {
				return model, cmd
			}
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
		if m.activeTab == tasksTab {
			// Handle task-specific keybindings
			switch m.taskDetailMode {
			case taskDetailNone:
				// If no task is selected in detail view, delegate all keys to updateTasksList
				// This will handle CursorUp/Down and Enter to select a task
				model, cmd := m.updateTasksList(msg)
				// If a task was selected by updateTasksList, transition to readonly mode
				if m.CoreModel.GetSelectedTask() != nil {
					m.taskDetailMode = taskDetailReadonly
				}
				return model, cmd
			case taskDetailReadonly:
				switch {
				case key.Matches(msg, m.keys.ExitEdit): // 'e' key to edit
					m.taskDetailMode = taskDetailEdit
					if task := m.CoreModel.GetSelectedTask(); task != nil {
						m.form = updateTaskForm(*task)
						return m, m.form.Init()
					}
				case key.Matches(msg, m.keys.Back):
					m.CoreModel.selectedTask = nil
					m.taskDetailMode = taskDetailNone
					return m, nil
				case key.Matches(msg, m.keys.CursorUp), key.Matches(msg, m.keys.CursorDown):
					// Allow navigation of the list even in readonly mode
					model, cmd := m.updateTasksList(msg)
					return model, cmd
				}
			case taskDetailEdit:
				// Delegate messages to the form
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
		}

		// General project view keys (applies to all tabs if not handled above)
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
	return m, cmd
}

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
			if m.selectedTaskIndex < len(tasks)-1 {
				m.selectedTaskIndex++
			}
		case key.Matches(msg, m.keys.ToggleDone):
			task := tasks[m.selectedTaskIndex]
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
		case key.Matches(msg, m.keys.SelectTask):
			coreCmd := m.CoreModel.SelectTask(m.selectedTaskIndex)
			if coreCmd == CoreShowError {
				return m, nil
			}
			m.taskDetailMode = taskDetailReadonly
			return m, nil
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

	leftColumn := leftColumnStyle.
		Width(leftWidth).
		Height(m.height - 4).
		Render(m.list.View())

	rightColumn := rightColumnStyle.
		Width(rightWidth).
		Height(m.height - 4).
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

	var s strings.Builder

	s.WriteString(m.renderTabs())
	s.WriteString("\n")

	switch m.activeTab {
	case projectDetailTab:
		s.WriteString(m.renderProjectDetails())
	case tasksTab:
		if m.CoreModel.GetSelectedTask() == nil || m.taskDetailMode == taskDetailNone {
			s.WriteString(m.renderTasksListOnly())
		} else {
			return m.renderTasksSplitView()
		}
	case logsTab:
		s.WriteString(m.renderLogsList())
	}

	s.WriteString("\n")
	s.WriteString(strings.Repeat("\n", 50))
	s.WriteString(m.help.View(m.keys))
	return s.String()
}

func (m *Model) renderTasksSplitView() string {
	leftWidth := m.width/2 - 4
	rightWidth := m.width/2 - 4

	tasks := m.CoreModel.GetTasks()
	var taskListContent string
	if len(tasks) == 0 {
		taskListContent = emptyDetailStyle.Render("No tasks for this project.")
	} else {
		var s strings.Builder
		for i, t := range tasks {
			checkbox := "[ ]"
			if t.CompletedAt != nil {
				checkbox = "[x]"
			}
			taskLine := checkbox + " " + t.Title
			if i == m.selectedTaskIndex {
				taskLine = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
					Render(taskLine)
			}
			s.WriteString(detailItemStyle.Render(taskLine))
			s.WriteString("\n")
		}
		taskListContent = s.String()
	}

	leftColumn := leftColumnStyle.
		Width(leftWidth).
		Height(m.height - 4).
		Render(taskListContent)

	rightColumn := rightColumnStyle.
		Width(rightWidth).
		Height(m.height - 4).
		Render(m.renderTasksListPanel())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}

func (m *Model) renderTasksListPanel() string {

	var s strings.Builder

	s.WriteString(detailSectionStyle.Render("Tasks"))
	s.WriteString("\n")

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

	task := m.CoreModel.GetSelectedTask()
	if task == nil {
		return emptyDetailStyle.Render("No task selected")
	}

	s.WriteString(detailTitleStyle.Render(task.Title))
	s.WriteString("\n\n")

	if task.Desc != "" {
		s.WriteString(detailItemStyle.Render("Description:"))
		s.WriteString("\n")
		s.WriteString(detailItemStyle.Render(task.Desc))
		s.WriteString("\n")
	} else {
		s.WriteString(detailItemStyle.Render("No description"))
		s.WriteString("\n")
	}

	// Show completion status
	if task.CompletedAt != nil {
		s.WriteString("\n")
		s.WriteString(detailItemStyle.Render("Status: Completed"))
		s.WriteString("\n")
		s.WriteString(detailItemStyle.Render("Completed at: " + task.CompletedAt.Format("2006-01-02 15:04")))
	} else {
		s.WriteString("\n")
		s.WriteString(detailItemStyle.Render("Status: Pending"))
	}

	return s.String()
}

func (m *Model) renderTaskForm() string {

	var s strings.Builder

	s.WriteString(detailSectionStyle.Render("Task Details"))
	s.WriteString("\n")

	if m.form != nil {
		s.WriteString(m.form.View())
	} else {
		task := m.CoreModel.GetSelectedTask()
		if task != nil {
			s.WriteString(detailTitleStyle.Render(task.Title))
			s.WriteString("\n\n")
			if task.Desc != "" {
				s.WriteString(detailItemStyle.Render("Description:"))
				s.WriteString("\n")
				s.WriteString(detailItemStyle.Render(task.Desc))
				s.WriteString("\n")
			} else {
				s.WriteString(detailItemStyle.Render("No description"))
				s.WriteString("\n")
			}

			// Show completion status
			if task.CompletedAt != nil {
				s.WriteString("\n")
				s.WriteString(detailItemStyle.Render("Status: Completed"))
				s.WriteString("\n")
				s.WriteString(detailItemStyle.Render("Completed at: " + task.CompletedAt.Format("2006-01-02 15:04")))
			} else {
				s.WriteString("\n")
				s.WriteString(detailItemStyle.Render("Status: Pending"))
			}
		} else {
			s.WriteString(detailItemStyle.Render("No task selected"))
		}
	}

	return s.String()
}

func (m *Model) renderTabs() string {
	tabStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

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

	s.WriteString("\n")
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
	tasks := m.GetTasks()
	var s strings.Builder
	s.WriteString("\n")
	if len(tasks) == 0 {
		s.WriteString(detailItemStyle.Render("No tasks for this project."))
		s.WriteString("\n")
	} else {
		for i, t := range tasks {
			checkbox := "[ ]"
			if t.CompletedAt != nil {
				checkbox = "[x]"
			}
			taskLine := checkbox + " " + t.Title
			if i == m.selectedTaskIndex {
				taskLine = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
					Render(taskLine)
			}
			s.WriteString(detailItemStyle.Render(taskLine))
			s.WriteString("\n")
		}
	}
	return s.String()
}

func (m *Model) renderLogsList() string {

	logs := m.GetLogs()
	var s strings.Builder
	s.WriteString("\n")
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

