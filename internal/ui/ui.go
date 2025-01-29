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
}

// type Project struct {
// 	ID          int
// 	Name        string
// 	Description string
// 	Status      string
// 	DateCreated time.Time
// 	Tasks       []Task
// 	Logs        []Log
// }
//
// func (p Project) Title() string       { return p.Name }
// func (p Project) Description() string { return p.Status }
// func (p Project) FilterValue() string { return p.Name }

type Model struct {
	list     list.Model
	service  Service
	projects []service.Project
	err      error
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
		projects: projects,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

			// Create new project
		case "n":
			p := createProjectForm()
			if p != nil {
				err := m.service.CreateProject(p)
				if err != nil {
					m.err = err
					return m, nil
				}
			}

		// Update project
		case "u":
			if i, ok := m.list.SelectedItem().(service.Project); ok {
				p := updateProjectForm(i)
				if p != nil {
					err := m.service.UpdateProject(p)
					if err != nil {
						m.err = err
						return m, nil
					}
				}
			}
			// Delete project
		case "d":
			confirmDelete := confirmDelete()
			if confirmDelete {
				if i, ok := m.list.SelectedItem().(service.Project); ok {
					err := m.service.DeleteProject(i.ID)
					if err != nil {
						m.err = err
						return m, nil
					}
				}

			}
		}
	}

	// refresh projects
	projects, err := m.service.ListProjects()
	if err != nil {
		m.err = err
		return m, nil
	}
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = p
	}

	m.list.SetItems(items)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return appStyle.Render(m.list.View())
}

func confirmDelete() bool {
	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to delete?").
				Value(&confirm),
		),
	)
	form.Run()
	return confirm

}

func updateProjectForm(p service.Project) *service.Project {
	var save bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("Enter name of project").
				Value(&p.Name).
				Accessor(huh.NewPointerAccessor(&p.Name)),
			huh.NewText().
				CharLimit(255).
				Title("Summary").
				Placeholder("Enter overview of project").
				Value(&p.Summary).
				Accessor(huh.NewPointerAccessor(&p.Summary)),
			huh.NewText().
				Title("Description (Optional)").
				Placeholder("Enter detailed description of project").
				Value(&p.Desc).
				Accessor(huh.NewPointerAccessor(&p.Desc)),
			huh.NewSelect[string]().Title("Status").
				Options(
					huh.NewOption("Todo", "todo"),
					huh.NewOption("In Progress", "in progress"),
					huh.NewOption("Done", "completed"),
					huh.NewOption("Archived", "archived"),
				).
				Value(&p.Status).
				Accessor(huh.NewPointerAccessor(&p.Status)),
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
