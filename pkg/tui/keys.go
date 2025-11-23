package tui

import "github.com/charmbracelet/bubbles/key"

var DefaultKeyMap = KeyMap{
	ToggleTrace: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle tracing"),
	),
	Help: key.NewBinding(
		key.WithKeys("h", "?"),
		key.WithHelp("h", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl-c"),
		key.WithHelp("q", "exit tetra-mess"),
	),
}

type KeyMap struct {
	ToggleTrace key.Binding
	Help        key.Binding
	Quit        key.Binding
}

func (m KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{m.ToggleTrace, m.Quit}
}

func (m KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.ToggleTrace},
		{m.Help, m.Quit},
	}
}
