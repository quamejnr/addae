package ui

import "github.com/charmbracelet/lipgloss"

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

	taskEditFormStyle = lipgloss.NewStyle().
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 2)

	projectDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#BD93F9"))

	subStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#6A6A6A",
	})
)
