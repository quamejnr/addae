package ui

import (
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
	list list.Model
	form *huh.Form
}

var appStyle = lipgloss.NewStyle().Margin(1, 2)

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
	projectList := list.New(items, delegate, 80, 20)
	projectList.Title = "Addae"
	projectList.SetShowHelp(true)

	return &Model{
		CoreModel: coreModel,
		list:      projectList,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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
	m.list, cmd = m.list.Update(msg)

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
		return m.projectView()
	default: // listView
		return appStyle.Render(m.list.View())
	}
}

func (m *Model) projectView() string {
	project := m.GetSelectedProject()
	if project == nil {
		return "No project selected."
	}

	var s string
	s += lipgloss.NewStyle().Bold(true).Render(project.Name) + "\n"
	s += "Status: " + project.Status + "\n"
	s += "Description: " + project.Desc + "\n\n"

	s += lipgloss.NewStyle().Bold(true).Render("Tasks") + "\n"
	tasks := m.GetTasks()
	if len(tasks) == 0 {
		s += "No tasks for this project.\n"
	} else {
		for _, t := range tasks {
			s += "- " + t.Title + " (" + t.Status + ")\n"
		}
	}
	s += "\n"

	s += lipgloss.NewStyle().Bold(true).Render("Logs") + "\n"
	logs := m.GetLogs()
	if len(logs) == 0 {
		s += "No logs for this project.\n"
	} else {
		for _, l := range logs {
			s += "- " + l.Title + "\n"
		}
	}

	return appStyle.Render(s)
}

// Form helper functions
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
