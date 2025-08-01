package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/quamejnr/addae/internal/service"
)

var theme *huh.Theme = huh.ThemeDracula()

func updateProjectForm(p service.Project) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Key("name").
				Value(&p.Name).
				Placeholder("My Awesome Project"),
			huh.NewText().
				Title("Project Summary").
				Key("summary").
				CharLimit(255).
				Value(&p.Summary).
				Placeholder("A short description of what this project is about..."),
			huh.NewText().
				Title("Detailed Description").
				Key("desc").
				Value(&p.Desc).
				CharLimit(0).
				Placeholder("Provide detailed information about your project..."),
			huh.NewSelect[string]().
				Title("Project Status").
				Key("status").
				Options(
					huh.NewOption("◯ Todo", "todo"),
					huh.NewOption("◐ In Progress", "in progress"),
					huh.NewOption("● Completed", "completed"),
					huh.NewOption("▣ Archived", "archived"),
				).
				Value(&p.Status),
		).Title("Update Project").
			Description("Modify your project details"),
	).WithTheme(theme)
}

func createProjectForm() *huh.Form {
	defaultValue := "todo"
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Key("name").
				Placeholder("My New Project").
				Validate(func(str string) error {
					if len(str) == 0 {
						return fmt.Errorf("project name is required")
					}
					return nil
				}),
			huh.NewText().
				Title("Project Summary").
				Key("summary").
				CharLimit(255).
				Placeholder("What is this project about in one sentence?"),
			huh.NewText().
				Title("Description (Optional)").
				Key("desc").
				CharLimit(0).
				Placeholder("Provide comprehensive details about your project goals, requirements, and scope..."),
			huh.NewSelect[string]().
				Title("Status").
				Key("status").
				Options(
					huh.NewOption("◯ Todo", "todo"),
					huh.NewOption("◐ In Progress", "in progress"),
					huh.NewOption("● Completed", "completed"),
					huh.NewOption("▣ Archived", "archived"),
				).
				Value(&defaultValue),
		).Title("Create New Project").
			Description("Set up your new project with essential details"),
	).WithTheme(theme)
}

// TaskEditForm represents the form for editing a task.
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
	descInput.SetHeight(5)
	descInput.FocusedStyle.CursorLine = lipgloss.NewStyle()
	descInput.ShowLineNumbers = false
    descInput.CharLimit = 0

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
	s.WriteString("\n")
	s.WriteString(f.titleInput.View())
	s.WriteString("\n")

	s.WriteString("\n")
	s.WriteString(f.descInput.View())
	s.WriteString(strings.Repeat("\n", 10))

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

// LogEditForm represents the form for editing a log entry.
type LogEditForm struct {
	titleInput textinput.Model
	textarea   textarea.Model

	width     int
	height    int
	focus     int
	completed bool
	aborted   bool
}

func newLogEditForm(width, height int) *LogEditForm {
	titleInput := textinput.New()
	titleInput.Placeholder = "Log Title"
	titleInput.Focus()
	titleInput.Width = width - 6

	ta := textarea.New()
	ta.Placeholder = "Describe your log entry here [markdown support]..."
	ta.SetWidth(width - 6)
	ta.SetHeight(height - 6)
	ta.ShowLineNumbers = false
	ta.CharLimit = 0

	return &LogEditForm{
		titleInput: titleInput,
		textarea:   ta,
		width:      width,
		height:     height,
		focus:      0,
	}
}

func newLogEditFormWithData(width, height int, title, desc string) *LogEditForm {
	form := newLogEditForm(width, height)
	form.titleInput.SetValue(title)
	form.textarea.SetValue(desc)
	return form
}

func (f *LogEditForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f *LogEditForm) Update(msg tea.Msg) (*LogEditForm, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			f.completed = true
			return f, nil
		case "esc":
			f.aborted = true
			return f, nil
		case "tab":
			f.focus = (f.focus + 1) % 2
			if f.focus == 0 {
				f.titleInput.Focus()
				f.textarea.Blur()
			} else {
				f.titleInput.Blur()
				f.textarea.Focus()
			}
			return f, nil
		}
	}

	if f.focus == 0 {
		f.titleInput, cmd = f.titleInput.Update(msg)
	} else {
		f.textarea, cmd = f.textarea.Update(msg)
	}
	cmds = append(cmds, cmd)

	return f, tea.Batch(cmds...)
}

func (f *LogEditForm) View() string {
	help := subStyle.Render("ctrl+s: save • esc: cancel • tab: next field ")
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		f.titleInput.View(),
		"",
		f.textarea.View(),
		"",
		help,
	)
}

func (f *LogEditForm) GetContent() (string, string) {
	return f.titleInput.Value(), f.textarea.Value()
}

func (f *LogEditForm) IsCompleted() bool {
	return f.completed
}

func (f *LogEditForm) IsAborted() bool {
	return f.aborted
}
