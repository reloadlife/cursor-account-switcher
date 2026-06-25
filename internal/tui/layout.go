package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func truncate(s string, max int) string {
	if max <= 3 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-len(r))
}

func joinHelp(keys ...string) string {
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = helpStyle.Render(k)
	}
	return strings.Join(parts, helpStyle.Render("  ·  "))
}

func renderKeyHelp(bindings [][2]string) string {
	parts := make([]string, 0, len(bindings))
	for _, b := range bindings {
		parts = append(parts, helpKey.Render(b[0])+helpStyle.Render(" "+b[1]))
	}
	return strings.Join(parts, helpStyle.Render("  "))
}

func boxTitle(title string) string {
	return appTitle.Render(title)
}

func statusDot(live, saved bool) string {
	switch {
	case live:
		return activeBadge.Render("●")
	case saved:
		return savedBadge.Render("○")
	default:
		return dimStyle.Render("·")
	}
}

func fitHeight(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	return strings.Join(lines[:maxLines], "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sidebarWidth(total int) int {
	w := total / 4
	if w < 18 {
		w = 18
	}
	if w > 28 {
		w = 28
	}
	return w
}

func contentWidth(total, side int) int {
	w := total - side - 4
	if w < 30 {
		w = 30
	}
	return w
}

func lipglossWidth(s string) int {
	return lipgloss.Width(s)
}
