package tui

import "github.com/charmbracelet/lipgloss"

var (
	headingStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Align(lipgloss.Left)

	tableStyle = lipgloss.NewStyle()

	userMessageStyle = lipgloss.NewStyle().
				Align(lipgloss.Left).
				Bold(true)
)
