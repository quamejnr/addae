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
)

type Model struct {
	list            list.Model
	service         Service
	state           viewState
	selectedProject *service.Project
	projects        []service.Project
	tasks           []service.Task
	logs            []service.Log
	form            *huh.Form
	err             error
}

var appStyle = lipgloss.NewStyle().Margin(1, 2)

func NewModel(svc Service) (*Model, error) {
	projects, err := svc.ListProjects()
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = p
	}

	delegate := list.NewDefaultDelegate()
	projectList := list.New(items, delegate, 80, 20)
	projectList.Title = "Addae"
	projectList.SetShowHelp(true)

	return &Model{
		list:     projectList,
		service:  svc,
		state:    listView,
		projects: projects,
	}, nil
}

func (m Model) Init() tea.Cmd {
  return tea.Batch(tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case listView:
		// Handle list view updates
		m.list, cmd = m.list.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if i, ok := m.list.SelectedItem().(service.Project); ok {
					m.state = projectView
					m.selectedProject = &i
					// Load tasks and logs for the selected project
					tasks, err := m.service.ListProjectTasks(i.ID)
					if err != nil {
						m.err = err
					} else {
						m.tasks = tasks
					}
					logs, err := m.service.ListProjectLogs(i.ID)
					if err != nil {
						m.err = err
					} else {
						m.logs = logs
					}
				}
			case "n":
				p := createProjectForm()
				if p != nil {
					if err := m.service.CreateProject(p); err != nil {
						m.err = err
					}
				}
			case "d":
				if confirmDelete() {
					if i, ok := m.list.SelectedItem().(service.Project); ok {
						if err := m.service.DeleteProject(i.ID); err != nil {
							m.err = err
						}
					}
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
		// Refresh projects list after potential changes
		projects, err := m.service.ListProjects()
		if err != nil {
			m.err = err
		} else {
			items := make([]list.Item, len(projects))
			for i, p := range projects {
				items[i] = p
			}
			m.list.SetItems(items)
		}

	case projectView:
		// Handle project view updates
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "u":
				m.state = updateView
				m.form = updateProjectForm(*m.selectedProject)
				return m, m.form.Init()
			case "esc", "b", "q", "ctrl+c":
				m.state = listView
				m.selectedProject = nil
			}
		}

	case updateView:
		// Handle update form updates
		var formCmd tea.Cmd
		var updatedForm tea.Model
		updatedForm, formCmd = m.form.Update(msg)
		m.form = updatedForm.(*huh.Form)
		cmd = formCmd

		if m.form.State == huh.StateAborted {
			m.state = projectView
			m.form = nil
		} else if m.form.State == huh.StateCompleted {
			p := m.selectedProject
			p.Name = m.form.GetString("name")
			p.Summary = m.form.GetString("summary")
			p.Desc = m.form.GetString("desc")
			p.Status = m.form.GetString("status")
			if err := m.service.UpdateProject(p); err != nil {
				m.err = err
			}
			m.state = projectView
			m.form = nil
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}

	switch m.state {
	case updateView:
		return m.form.View()
	case projectView:
		return m.projectView()
	default: // listView
		return appStyle.Render(m.list.View())
	}
}

func (m *Model) projectView() string {
	if m.selectedProject == nil {
		return "No project selected."
	}

	var s string
	s += lipgloss.NewStyle().Bold(true).Render(m.selectedProject.Name) + "\n"
	s += "Status: " + m.selectedProject.Status + "\n"
	s += "Description: " + m.selectedProject.Desc + "\n\n"

	s += lipgloss.NewStyle().Bold(true).Render("Tasks") + "\n"
	if len(m.tasks) == 0 {
		s += "No tasks for this project.\n"
	} else {
		for _, t := range m.tasks {
			s += "- " + t.Title + " (" + t.Status + ")\n"
		}
	}
	s += "\n"

	s += lipgloss.NewStyle().Bold(true).Render("Logs") + "\n"
	if len(m.logs) == 0 {
		s += "No logs for this project.\n"
	} else {
		for _, l := range m.logs {
			s += "- " + l.Title + "\n"
		}
	}

	return appStyle.Render(s)
}

func confirmDelete() bool {
	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to delete this project?").
				Value(&confirm),
		),
	)
	form.Run()
	return confirm
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

func createProjectForm() *service.Project {
	var p service.Project
	var save bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("Enter name of project").
				Value(&p.Name),
			huh.NewText().
				CharLimit(255).
				Title("Summary").
				Placeholder("Enter overview of project").
				Value(&p.Summary),
			huh.NewText().
				Title("Description (Optional)").
				Placeholder("Enter detailed description of project").
				Value(&p.Desc),
			huh.NewSelect[string]().Title("Status").
				Options(
					huh.NewOption("Todo", "todo"),
					huh.NewOption("In Progress", "in progress"),
					huh.NewOption("Done", "completed"),
					huh.NewOption("Archived", "archived"),
				).
				Value(&p.Status),
			huh.NewConfirm().
				Affirmative("Save").
				Negative("Cancel").
				Value(&save),
		),
	)
	form.Run()
	if !save {
		return nil
	}
	return &p
}