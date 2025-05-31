package tui

import "github.com/charmbracelet/lipgloss"

var (
	headingStyle = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center)

	boxStyle = lipgloss.NewStyle().
			Align(lipgloss.Left)
)
