package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quamejnr/addae/internal/service"
)

type Service interface {
	ListProjects() ([]service.Project, error)
	DeleteProject(id int) error
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

func New(svc Service) (*Model, error) {
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
	projectList.Title = "Projects"
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
		case "n":
			// Handle new project (implement as needed)
			return m, nil
		case "d":
			if _, ok := m.list.SelectedItem().(service.Project); ok {
				// err := m.service.DeleteProject(i.ID)
				// if err != nil {
				// 	m.err = err
				// 	return m, nil
				// }
				// Refresh projects
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
			}
		}
	}

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
