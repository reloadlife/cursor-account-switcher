package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	missStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

func StatusText() string {
	currentEmail, _ := profiles.CurrentEmail()
	active := profiles.ActiveAccount()
	accounts, _ := profiles.ListAccounts()

	var b strings.Builder
	b.WriteString(titleStyle.Render("Cursor Account Switcher"))
	b.WriteString("\n\n")

	writeRow(&b, "Current email", formatEmail(currentEmail))
	writeRow(&b, "Active profile", formatActive(active))

	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Accounts"))
	b.WriteString("\n")

	for _, account := range accounts {
		profile, _ := profiles.Load(account.ID)
		line := missStyle.Render("not saved")
		if profile != nil {
			email := "unknown"
			if profile.Email != nil {
				email = *profile.Email
			}
			savedAt := profile.SavedAt
			if t, err := time.Parse(time.RFC3339, profile.SavedAt); err == nil {
				savedAt = t.Local().Format("Jan 2 15:04")
			}
			line = okStyle.Render(fmt.Sprintf("%s · saved %s", email, savedAt))
		}
		b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
			valueStyle.Render(account.Label),
			dimStyle.Render("("+string(account.ID)+")"),
			line,
		))
	}

	return b.String()
}

func writeRow(b *strings.Builder, label, value string) {
	b.WriteString(fmt.Sprintf("  %s  %s\n",
		labelStyle.Render(label+":"),
		valueStyle.Render(value),
	))
}

func formatEmail(email string) string {
	if email == "" {
		return missStyle.Render("not signed in")
	}
	return email
}

func formatActive(active *paths.AccountID) string {
	if active == nil {
		return missStyle.Render("unknown")
	}
	return profiles.AccountLabel(*active)
}
