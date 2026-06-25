package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

type ActionKind int

const (
	ActionNone ActionKind = iota
	ActionQuit
	ActionStatus
	ActionSwitch
	ActionSave
	ActionAddAccount
)

type MenuResult struct {
	Kind      ActionKind
	AccountID paths.AccountID
}

type menuItem struct {
	title, desc string
	kind        ActionKind
	accountID   paths.AccountID
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

type menuModel struct {
	list     list.Model
	selected MenuResult
	quitting bool
	width    int
	height   int
}

func accountDesc(id paths.AccountID, currentEmail string, active *paths.AccountID) string {
	profile, _ := profiles.Load(id)
	if profile == nil {
		return missingBadge.Render("not saved")
	}

	isActive := (active != nil && *active == id) ||
		(profile.Email != nil && *profile.Email == currentEmail)

	if isActive {
		email := ""
		if profile.Email != nil {
			email = *profile.Email
		}
		return activeBadge.Render("active") + " · " + email
	}

	if profile.Email != nil {
		return savedBadge.Render(*profile.Email)
	}

	return savedBadge.Render("saved")
}

func buildMenuItems() ([]list.Item, error) {
	currentEmail, _ := profiles.CurrentEmail()
	active := profiles.ActiveAccount()
	accounts, err := profiles.ListAccounts()
	if err != nil {
		return nil, err
	}

	var items []list.Item

	for _, account := range accounts {
		items = append(items, menuItem{
			title:     "Switch to " + account.Label,
			desc:      accountDesc(account.ID, currentEmail, active),
			kind:      ActionSwitch,
			accountID: account.ID,
		})
	}

	if len(accounts) > 0 {
		items = append(items, menuItem{
			title: "───",
			desc:  "",
			kind:  ActionNone,
		})
	}

	for _, account := range accounts {
		items = append(items, menuItem{
			title:     "Save as " + account.Label,
			desc:      "capture current Cursor session",
			kind:      ActionSave,
			accountID: account.ID,
		})
	}

	items = append(items,
		menuItem{title: "Add account", desc: "register a new account with a custom label", kind: ActionAddAccount},
		menuItem{title: "Show status", desc: "view saved profiles and current session", kind: ActionStatus},
		menuItem{title: "Quit", desc: "exit without changes", kind: ActionQuit},
	)

	return items, nil
}

func RunMenu() (MenuResult, error) {
	items, err := buildMenuItems()
	if err != nil {
		return MenuResult{}, err
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("252"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "What would you like to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	m := menuModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())

	final, err := p.Run()
	if err != nil {
		return MenuResult{}, err
	}

	model := final.(menuModel)
	return model.selected, nil
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			m.selected = MenuResult{Kind: ActionQuit}
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(menuItem); ok {
				if item.kind == ActionNone {
					return m, nil
				}
				m.selected = MenuResult{Kind: item.kind, AccountID: item.accountID}
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(appTitle.Render("Cursor Account Switcher"))
	b.WriteString("\n")
	if email, _ := profiles.CurrentEmail(); email != "" {
		b.WriteString(appSubtitle.Render("Signed in as " + email))
		b.WriteString("\n")
	}
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ navigate · enter select · q quit"))
	return b.String()
}

var appStyle = lipgloss.NewStyle().Padding(1, 2)

func FormatSaveSuccess(id paths.AccountID, email string) string {
	return fmt.Sprintf("Saved %s profile (%s)", profiles.AccountLabel(id), email)
}
