package ui

import "github.com/charmbracelet/bubbles/key"

// ProjectKeyMap defines the keybindings for the application.
type ProjectKeyMap struct {
	UpdateProject   key.Binding
	CreateLog       key.Binding
	GotoDetails     key.Binding
	GotoTasks       key.Binding
	GotoLogs        key.Binding
	TabLeft         key.Binding
	TabRight        key.Binding
	Back            key.Binding
	Help            key.Binding
	ToggleDone      key.Binding
	DeleteObject    key.Binding
	SelectObject    key.Binding
	CursorUp        key.Binding
	CursorDown      key.Binding
	Edit            key.Binding
	SwitchFocus     key.Binding
	ToggleCompleted key.Binding
	CreateObject    key.Binding
	CreateTask      key.Binding
}

// ShortHelp returns a slice of keybindings for the short help view.
func (k ProjectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.UpdateProject, k.CreateObject, k.TabRight, k.ToggleDone, k.Back, k.Help,
	}
}

// FullHelp returns a slice of keybindings for the full help view.
func (k ProjectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// navigation
		{
			k.TabLeft, k.TabRight, k.GotoDetails, k.GotoTasks,
			k.GotoLogs, k.CursorUp, k.CursorDown, k.SwitchFocus, k.Back,
		},
		// actions
		{
			k.CreateObject, k.UpdateProject, k.CreateTask, k.CreateLog, k.Edit,
			k.ToggleDone, k.ToggleCompleted, k.DeleteObject,
		},
		// help
		{k.Help},
	}
}

// projectKeys holds the actual keybindings for the application.
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
		key.WithHelp("esc/b", "back"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	ToggleDone: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle done"),
	),
	DeleteObject: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete object"),
	),
	SelectObject: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select object"),
	),
	CursorUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	CursorDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	CreateObject: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "create object"),
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
