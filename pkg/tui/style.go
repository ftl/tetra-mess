package tui

import "github.com/charmbracelet/lipgloss"

var (
	headingStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Align(lipgloss.Left)

	userMessageStyle = lipgloss.NewStyle().
				Align(lipgloss.Left).
				Bold(true)
)
