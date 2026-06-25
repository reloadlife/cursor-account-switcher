package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

type addAccountModel struct {
	input    textinput.Model
	err      string
	quitting bool
	done     bool
	result   paths.AccountID
}

func RunAddAccount() (paths.AccountID, error) {
	ti := textinput.New()
	ti.Placeholder = "e.g. Freelance Client"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40
	ti.Prompt = "Label: "

	m := addAccountModel{input: ti}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}

	model := final.(addAccountModel)
	if model.err != "" {
		return "", fmt.Errorf("%s", model.err)
	}
	return model.result, nil
}

func (m addAccountModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addAccountModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			label := strings.TrimSpace(m.input.Value())
			if label == "" {
				m.err = "label cannot be empty"
				return m, tea.Quit
			}
			id, err := app.RegisterAccountFromLabel(label)
			if err != nil {
				m.err = err.Error()
				return m, tea.Quit
			}
			m.done = true
			m.result = id
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m addAccountModel) View() string {
	var b strings.Builder
	b.WriteString(appTitle.Render("Add Account"))
	b.WriteString("\n\n")
	b.WriteString(appSubtitle.Render("Enter a display label for the new account"))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	if m.err != "" {
		b.WriteString(missingBadge.Render(m.err))
		b.WriteString("\n")
	} else if m.done {
		b.WriteString(activeBadge.Render(fmt.Sprintf("Created %s (%s)", profiles.AccountLabel(m.result), m.result)))
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render("enter confirm · esc cancel"))
	return b.String()
}
