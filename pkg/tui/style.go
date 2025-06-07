package tui

import "github.com/charmbracelet/lipgloss"

var (
	headingStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Center).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Left).
			Padding(0, 0, 1)

	tableStyle = lipgloss.NewStyle()

	tableSelectedStyle = lipgloss.NewStyle()

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1)

	tableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	userMessageStyle = lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Left).
				Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			AlignVertical(lipgloss.Bottom).
			Reverse(true)
)
