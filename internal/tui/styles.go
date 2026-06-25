package tui

import "github.com/charmbracelet/lipgloss"

var (
	appTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	appSubtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	activeBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	savedBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	missingBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))
)
