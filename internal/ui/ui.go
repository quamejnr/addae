package ui

import (
	"strings"

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
}

type viewState int

const (
	listView viewState = iota
	projectView
	updateView
	createView
	deleteView
)

type Model struct {
	*CoreModel
	list   list.Model
	form   *huh.Form
	width  int
	height int
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
				Bold(true).
				Foreground(lipgloss.Color("205")).
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
	}

	return m, cmd
}

func (m *Model) updateListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	oldIndex := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	m.loadProjectDetails(oldIndex)

	// If selection changed, load project details without changing state
	if m.list.Index() != oldIndex && m.list.Index() >= 0 {
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
		switch msg.String() {
		case "u":
			m.CoreModel.GoToUpdateView()
			if project := m.CoreModel.GetSelectedProject(); project != nil {
				m.form = updateProjectForm(*project)
				return m, m.form.Init()
			}
		case "esc", "b", "q", "ctrl+c":
			m.CoreModel.GoToListView()
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
	if tasks, err := m.CoreModel.service.ListProjectTasks(project.ID); err != nil {
		m.CoreModel.tasks = tasks
	}

	if logs, err := m.CoreModel.service.ListProjectLogs(project.ID); err != nil {
		m.CoreModel.logs = logs
	}
}

func (m Model) View() string {
	if err := m.GetError(); err != nil {
		return "Error: " + err.Error()
	}

	switch m.GetState() {
	case updateView, createView, deleteView:
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

	// Project header
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

	// Tasks section
	s.WriteString(detailSectionStyle.Render("Tasks"))
	s.WriteString("\n")
	tasks := m.GetTasks()
	if len(tasks) == 0 {
		s.WriteString(detailItemStyle.Render("No tasks for this project."))
		s.WriteString("\n")
	} else {
		for _, t := range tasks {
			s.WriteString(detailItemStyle.Render("• " + t.Title + " (" + t.Status + ")"))
			s.WriteString("\n")
		}
	}

	// Logs section
	s.WriteString(detailSectionStyle.Render("Logs"))
	s.WriteString("\n")
	logs := m.GetLogs()
	if len(logs) == 0 {
		s.WriteString(detailItemStyle.Render("No logs for this project."))
		s.WriteString("\n")
	} else {
		for _, l := range logs {
			s.WriteString(detailItemStyle.Render("• " + l.Title))
			s.WriteString("\n")
		}
	}

	// Help text at bottom
	s.WriteString("\n")
	s.WriteString(emptyDetailStyle.Render("Press 'enter' to focus • 'u' to update • 'n' to create • 'd' to delete"))

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
