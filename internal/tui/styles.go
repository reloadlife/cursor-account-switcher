package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent  = lipgloss.Color("205")
	colorSuccess = lipgloss.Color("42")
	colorWarn    = lipgloss.Color("214")
	colorError   = lipgloss.Color("203")
	colorMuted   = lipgloss.Color("241")
	colorText    = lipgloss.Color("252")
	colorBorder  = lipgloss.Color("238")
	colorDim     = lipgloss.Color("245")

	appTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	appSubtitle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpKey = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	activeBadge = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	savedBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	missingBadge = lipgloss.NewStyle().
			Foreground(colorError)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorText)

	headerBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorBorder).
			Padding(0, 1).
			MarginBottom(1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	sidebarStyle = lipgloss.NewStyle().
			Padding(0, 1)

	sidebarActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Padding(0, 1)

	sidebarItem = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1)

	tableHeader = lipgloss.NewStyle().
			Foreground(colorMuted).
			Bold(true)

	rowSelected = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	rowNormal = lipgloss.NewStyle().
			Foreground(colorText)

	footerBar = lipgloss.NewStyle().
			Foreground(colorMuted).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(colorBorder).
			Padding(1, 0, 0, 0).
			MarginTop(1)

	noticeBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2).
			Width(52)

	inputBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)
)
