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

	helpStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

var (
	ANSIGreen  = lipgloss.ANSIColor(28) // 6)
	ANSIYellow = lipgloss.ANSIColor(3)
	ANSIRed    = lipgloss.ANSIColor(196)
)

func ganToANSIColor(gan int) lipgloss.ANSIColor {
	if gan > 1 {
		return ANSIGreen
	}
	if gan > -1 {
		return ANSIYellow
	}
	return ANSIRed
}

func sldToANSIColor(sld int) lipgloss.ANSIColor {
	if sld > 6 {
		return ANSIGreen
	}
	if sld == 6 {
		return ANSIYellow
	}
	return ANSIRed
}

func serversToANSIColor(servers int) lipgloss.ANSIColor {
	if servers > 2 {
		return ANSIGreen
	}
	if servers == 2 {
		return ANSIYellow
	}
	return ANSIRed
}
