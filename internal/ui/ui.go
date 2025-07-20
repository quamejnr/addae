package ui

import (
	"strings"

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
	CreateLog(projectID int, title, desc string) error
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
)

type detailTab int

const (
	projectDetailTab detailTab = iota
	tasksTab
	logsTab
)

type Model struct {
	*CoreModel
	list      list.Model
	form      *huh.Form
	width     int
	height    int
	activeTab detailTab
	help      help.Model
	keys      ProjectKeyMap
}

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
}

func (k ProjectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.UpdateProject, k.CreateTask, k.CreateLog, k.Back, k.Help}
}

func (k ProjectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.UpdateProject, k.CreateTask, k.CreateLog},
		{k.GotoDetails, k.GotoTasks, k.GotoLogs, k.TabLeft, k.TabRight},
		{k.Back, k.Help},
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
}

var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	// Column styles
	leftColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("240")).
			PaddingRight(1)

	rightColumnStyle = lipgloss.NewStyle().
				PaddingLeft(1)

	// Detail panel styles
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
		CoreModel: coreModel,
		list:      projectList,
		width:     80,
		height:    24,
		activeTab: projectDetailTab,
		help:      help.New(),
		keys:      projectKeys,
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

		// Update list dimensions for left column
		listWidth := m.width/2 - 4
		listHeight := m.height - 4
		m.list.SetSize(listWidth, listHeight)
	}

	switch m.GetState() {
	case listView:
		return m.updateListView(msg)
	case projectView:
		return m.updateProjectView(msg)
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
	}

	return m, cmd
}

func (m *Model) updateListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// If selection changed, load project details without changing state
	if m.list.Index() >= 0 {
		m.loadProjectDetails(m.list.Index())
	}

	// don't allow these key presses during filtering
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
					// Switch to delete view and show confirmation
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

func (m *Model) updateProjectView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		case key.Matches(msg, m.keys.GotoTasks):
			m.activeTab = tasksTab
		case key.Matches(msg, m.keys.GotoLogs):
			m.activeTab = logsTab
		case key.Matches(msg, m.keys.TabRight):
			m.activeTab = (m.activeTab + 1) % 3
		case key.Matches(msg, m.keys.TabLeft):
			m.activeTab = (m.activeTab - 1 + 3) % 3
		case key.Matches(msg, m.keys.Back):
			m.CoreModel.GoToListView()
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}
	return m, nil
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

		// Handle the core command and refresh if needed
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
	case "createTask", "createLog":
		m.CoreModel.GoToProjectView()
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
			// Get the selected item from the list for deletion
			if selected := m.list.Index(); selected >= 0 {
				return m.CoreModel.DeleteProject(selected)
			}
		}
		// If not confirmed or no selection, just go back to list
		m.CoreModel.GoToListView()
		return NoCoreCmd
	case "createTask":
		data := TaskFormData{
			Title:  m.form.GetString("title"),
			Desc:   m.form.GetString("desc"),
		}
		return m.CoreModel.CreateTask(data)
	case "createLog":
		data := LogFormData{
			Title: m.form.GetString("title"),
			Desc:  m.form.GetString("desc"),
		}
		return m.CoreModel.CreateLog(data)
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

// loadProjectDetails loads project details without changing the view state
func (m *Model) loadProjectDetails(index int) {
	if index < 0 || index >= len(m.CoreModel.GetProjects()) {
		return
	}

	projects := m.CoreModel.GetProjects()
	project := projects[index]

	// Set the selected project directly without changing state
	m.CoreModel.selectedProject = &project

	// Load tasks and logs
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

	switch m.GetState() {
	case updateView, createView, deleteView, createTaskView, createLogView:
		if m.form != nil {
			return m.form.View()
		}
		return "Loading form..."
	case projectView:
		// In project view, show list on left and project details on right
		return m.renderDetailPanel()
	default: // listView
		return m.renderTabularView()
	}
}

func (m *Model) renderTabularView() string {
	leftWidth := m.width/2 - 4
	rightWidth := m.width/2 - 4

	// Left column - project list
	leftColumn := leftColumnStyle.
		Width(leftWidth).
		Height(m.height - 4).
		Render(m.list.View())

	// Right column - project details
	rightColumn := rightColumnStyle.
		Width(rightWidth).
		Height(m.height - 4).
		Render(m.renderDetailPanel())

	// Join columns horizontally
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

	// Render tabs
	s.WriteString(m.renderTabs())
	s.WriteString("\n")

	// Render content based on active tab
	switch m.activeTab {
	case projectDetailTab:
		s.WriteString(m.renderProjectDetails())
	case tasksTab:
		s.WriteString(m.renderTasksList())
	case logsTab:
		s.WriteString(m.renderLogsList())
	}

	// Help text at bottom
	s.WriteString("\n")
	s.WriteString(strings.Repeat("\n", 50))
	s.WriteString(m.help.View(m.keys))
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

func (m *Model) renderTasksList() string {
	tasks := m.GetTasks()
	var s strings.Builder
	s.WriteString("\n")
	if len(tasks) == 0 {
		s.WriteString(detailItemStyle.Render("No tasks for this project."))
		s.WriteString("\n")
	} else {
		for _, t := range tasks {
			status := ""
			if t.CompletedAt != nil {
				status = " (Completed)"
			}
			s.WriteString(detailItemStyle.Render("• " + t.Title + status))
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

// Form helper functions (unchanged)
func confirmDeleteForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to delete this project?").
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
				),
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
