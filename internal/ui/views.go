package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/quamejnr/addae/internal/service"
)

func (m *Model) renderTabularView() string {
	leftWidth := m.width/2 - 4
	rightWidth := m.width/2 - 4
	splitHeight := m.height - 5

	m.list.SetHeight(splitHeight)

	leftColumn := leftColumnStyle.
		Width(leftWidth).
		Margin(2).
		Render(m.list.View())

	rightColumn := rightColumnStyle.
		Width(rightWidth).
		MaxHeight(splitHeight).
		Render(m.renderDetailPanel())

	return lipgloss.NewStyle().Render(
		lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn),
	)
}

func (m *Model) renderDetailPanel() string {
	project := m.GetSelectedProject()
	if project == nil {
		return emptyDetailStyle.Render("← Select a project to view details")
	}

	// Handle tasks split view
	if m.activeTab == tasksTab && m.taskDetailMode != taskDetailNone {
		return lipgloss.JoinVertical(
			lipgloss.Left, m.renderTabs(), m.renderTasksSplitView(), m.help.View(m.keys),
		)
	}

	// Handle logs split view
	if m.activeTab == logsTab && m.logDetailMode != logDetailNone {
		return lipgloss.JoinVertical(
			lipgloss.Left, m.renderTabs(), m.renderLogsSplitView(), m.help.View(m.keys),
		)
	}

	// Handle all other views normally
	var content string
	switch m.activeTab {
	case projectDetailTab:
		content = m.renderProjectDetails()
	case tasksTab:
		content = m.renderTasksListOnly()
	case logsTab:
		content = m.renderLogsListOnly()
	}

	availHeight := m.height - 4
	tabs := lipgloss.NewStyle().Render(m.renderTabs())
	tabsHeight := lipgloss.Height(tabs)
	availHeight -= tabsHeight
	helpView := m.help.View(m.keys)
	helpHeight := lipgloss.Height(helpView)
	availHeight -= helpHeight

	content = lipgloss.NewStyle().Height(availHeight).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, tabs, content, helpView)
}

func (m *Model) renderLogsSplitView() string {
	leftWidth := m.width/2 - 4
	rightWidth := m.width/2 - 4
	splitHeight := m.height - 8

	var logListContent strings.Builder

	logs := m.CoreModel.GetLogs()
	if len(logs) == 0 {
		logListContent.WriteString(emptyDetailStyle.Render("No logs for this project."))
	} else {
		for i, l := range logs {
			logLine := "• " + l.Title
			if i == m.selectedLogIndex {
				logLine = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
					Render(logLine)
			}
			logListContent.WriteString(detailItemStyle.Render(logLine))
			logListContent.WriteString("\n")
		}
	}

	focusIndicator := ""
	if m.logViewFocus == focusList {
		focusIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("69")).
			Render("[List Focus]")
	} else {
		focusIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("[Pager Focus]")
	}
	logListContent.WriteString("\n" + focusIndicator)

	leftColumn := leftColumnStyle.
		Width(leftWidth).
		Height(splitHeight).
		Render(logListContent.String())

	rightColumn := rightColumnStyle.
		Width(rightWidth).
		Height(splitHeight).
		Render(m.renderLogDetailPanel())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}

func (m *Model) renderLogDetailPanel() string {
	if m.logDetailMode == logDetailNone {
		return emptyDetailStyle.Render("← Select a log to view details")
	}

	log := m.CoreModel.GetSelectedLog()
	if log == nil {
		return emptyDetailStyle.Render("No log selected")
	}

	var s strings.Builder
	s.WriteString(detailTitleStyle.PaddingLeft(2).Render(log.Title))

	s.WriteString(m.logViewport.View())

	return s.String()
}

func (m *Model) renderTasksSplitView() string {
	leftWidth := m.width/2 - 4
	rightWidth := m.width/2 - 4
	splitHeight := m.height - 8

	var taskListContent strings.Builder

	if m.quickInputActive {
		taskListContent.WriteString(m.quickTaskInput.View() + "\n\n")
	}

	tasks := m.CoreModel.GetTasks()
	if len(tasks) == 0 {
		taskListContent.WriteString(emptyDetailStyle.Render("No tasks for this project."))
	} else {
		var pending, completed []service.Task
		for _, t := range tasks {
			if t.CompletedAt != nil {
				completed = append(completed, t)
			} else {
				pending = append(pending, t)
			}
		}

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
	s.WriteString(detailItemStyle.Render(task.Desc))
	s.WriteString("\n\n")

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
		Margin(1, 0, 1)

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
	s.WriteString(getStatusIndicator(project.Status))
	if project.Summary != "" {
		s.WriteString("\n\n")
		s.WriteString(projectDetailStyle.Render("Summary: ") + detailItemStyle.Render(project.Summary))
	}
	if project.Desc != "" {
		s.WriteString("\n\n")
		s.WriteString(projectDetailStyle.Render("Description: ") + detailItemStyle.Render(project.Desc))
	}
	s.WriteString("\n")
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

	var pending, completed []service.Task
	for _, t := range tasks {
		if t.CompletedAt != nil {
			completed = append(completed, t)
		} else {
			pending = append(pending, t)
		}
	}

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

func (m *Model) renderLogsListOnly() string {
	logs := m.GetLogs()
	var s strings.Builder

	if len(logs) == 0 {
		s.WriteString(detailItemStyle.Render("No logs for this project."))
		s.WriteString("\n")
	} else {
		for i, l := range logs {
			logLine := "• " + l.Title
			if i == m.selectedLogIndex {
				logLine = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
					Render(logLine)
			}
			s.WriteString(detailItemStyle.Render(logLine))
			s.WriteString("\n")
		}
	}
	return s.String()
}

func getStatusIndicator(status string) string {
	switch strings.ToLower(status) {
	case "todo":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("167")).
			Render("░░░ TODO")
	case "in progress":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("▓░░ IN PROGRESS")
	case "completed":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("71")).
			Render("▓▓▓ COMPLETED")
	case "archived":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("▓▓▓ ARCHIVED")
	default:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("▓▓▓ ARCHIVED")
	}
}

func (m *Model) renderCenteredForm() string {
	if m.form == nil {
		return ""
	}

	formContent := m.form.View()

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		formContent,
	)
}

func (m *Model) View() string {
	if err := m.GetError(); err != nil {
		return "Error: " + err.Error()
	}

	var mainContent string

	switch m.GetState() {
	case listView:
		mainContent = m.renderTabularView()
	case projectView:
		mainContent = m.renderDetailPanel()
	case fullscreenLogEditView, updateLogView:
		if m.logEditForm != nil {
			mainContent = m.logEditForm.View()
		}
	case updateView, createView, deleteView, createTaskView, createLogView, deleteTaskView, deleteLogView:
		mainContent = m.renderCenteredForm()
	}

	var finalView string
	switch m.deleteDialogType {
	case noDialog:
		finalView = mainContent
	case projectDeleteDialog:
		finalView = m.renderProjectDeleteDialog()
	case taskDeleteDialog:
		finalView = m.renderTaskDeleteDialog()
	case logDeleteDialog:
		finalView = m.renderLogDeleteDialog()
	default:
		finalView = mainContent
	}

	return appStyle.Render(finalView)
}

func (m *Model) renderProjectDeleteDialog() string {
	project := m.GetSelectedProject()
	styledName := detailTitleStyle.Render(fmt.Sprintf("'%s'", project.Name))
	question := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("Delete "),
		styledName,
		lipgloss.NewStyle().Bold(true).Render("?"),
	)
	subText := "This will delete all tasks and logs. This action cannot be undone."
	return m.renderConfirmationDialog(question, subText)
}

func (m *Model) renderTaskDeleteDialog() string {
	task := m.GetSelectedTask()
	styledName := detailTitleStyle.Render(fmt.Sprintf("'%s'", task.Title))
	question := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("Delete "),
		styledName,
		lipgloss.NewStyle().Bold(true).Render("?"),
	)
	return m.renderConfirmationDialog(question, "")
}

func (m *Model) renderLogDeleteDialog() string {
	log := m.GetSelectedLog()
	styledName := detailTitleStyle.Render(fmt.Sprintf("'%s'", log.Title))
	question := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("Delete "),
		styledName,
		lipgloss.NewStyle().Bold(true).Render("?"),
	)
	return m.renderConfirmationDialog(question, "")
}
func (m *Model) renderConfirmationDialog(question string, subtext string) string {
	// Dialog styling
	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	// Button styling
	cancelButton := "[ Cancel ]"
	deleteButton := "[ Delete ]"

	if m.deleteConfirmCursor == 0 {
		cancelButton = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(cancelButton)
	} else {
		deleteButton = lipgloss.NewStyle().Foreground(lipgloss.Color("197")).Bold(true).Render(deleteButton)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, cancelButton, " ", deleteButton)

	// Assemble dialog content
	ui := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render(question),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(subtext),
		"",
		buttons,
	)

	dialog := dialogBox.Render(ui)

	// Place the dialog in the center of the screen
	centeredDialog := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)

	return lipgloss.JoinVertical(lipgloss.Left, centeredDialog)
}
