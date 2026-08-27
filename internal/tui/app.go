package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/platform"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

type screen int

const (
	screenDashboard screen = iota
	screenAddAccount
	screenPickDefault
	screenSwitching
	screenNotice
)

type switchDoneMsg struct {
	err error
}

type switchStepMsg struct {
	label string
}

type appModel struct {
	screen   screen
	overview []profiles.PlatformOverview
	platIdx  int
	acctIdx  int
	width    int
	height   int
	quitting bool

	input   textinput.Model
	spinner spinner.Model

	switchStep string
	switchErr  error
	switchAcct paths.AccountID
	switchPlat platform.ID

	noticeTitle string
	noticeBody  string
	noticeOK    bool

	stepCh <-chan string
	doneCh <-chan error
}

func RunApp() error {
	m, err := newAppModel()
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func RunMenu() (MenuResult, error) {
	if err := RunApp(); err != nil {
		return MenuResult{}, err
	}
	return MenuResult{Kind: ActionQuit}, nil
}

func newAppModel() (appModel, error) {
	overview, err := profiles.AllPlatformsOverview()
	if err != nil {
		return appModel{}, err
	}

	ti := textinput.New()
	ti.Placeholder = "Freelance Client"
	ti.CharLimit = 64
	ti.Prompt = "Label  "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(colorMuted)
	ti.TextStyle = lipgloss.NewStyle().Foreground(colorText)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)

	return appModel{
		screen:   screenDashboard,
		overview: overview,
		input:    ti,
		spinner:  sp,
	}, nil
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) refreshOverview() (appModel, error) {
	overview, err := profiles.AllPlatformsOverview()
	if err != nil {
		return m, err
	}
	m.overview = overview
	if m.platIdx >= len(m.overview) {
		m.platIdx = maxInt(0, len(m.overview)-1)
	}
	if m.acctIdx >= len(m.currentAccounts()) {
		m.acctIdx = maxInt(0, len(m.currentAccounts())-1)
	}
	return m, nil
}

func (m appModel) currentPlatform() *profiles.PlatformOverview {
	if m.platIdx < 0 || m.platIdx >= len(m.overview) {
		return nil
	}
	return &m.overview[m.platIdx]
}

func (m appModel) currentAccounts() []profiles.AccountOverview {
	if p := m.currentPlatform(); p != nil {
		return p.Accounts
	}
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = minInt(40, maxInt(20, msg.Width-20))
		return m, nil

	case switchDoneMsg:
		m.screen = screenNotice
		m.switchErr = msg.err
		if msg.err != nil {
			m.noticeTitle = "Switch failed"
			m.noticeBody = msg.err.Error()
			m.noticeOK = false
		} else {
			m.noticeTitle = "Switched"
			m.noticeBody = m.switchStep
			m.noticeOK = true
		}
		refreshed, err := m.refreshOverview()
		if err != nil {
			m.noticeTitle = "Error"
			m.noticeBody = err.Error()
			m.noticeOK = false
		} else {
			m = refreshed
		}
		return m, nil

	case switchStepMsg:
		m.switchStep = msg.label
		return m, m.watchSwitch()

	case tea.KeyMsg:
		if m.screen == screenNotice {
			m.screen = screenDashboard
			return m, nil
		}

		switch m.screen {
		case screenDashboard:
			return m.updateDashboard(msg)
		case screenAddAccount:
			return m.updateAddAccount(msg)
		case screenPickDefault:
			return m.updatePickDefault(msg)
		case screenSwitching:
			if msg.String() == "ctrl+c" {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	if m.screen == screenSwitching {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.screen == screenAddAccount {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m appModel) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.acctIdx > 0 {
			m.acctIdx--
		}
	case "down", "j":
		if m.acctIdx < len(m.currentAccounts())-1 {
			m.acctIdx++
		}
	case "left", "h", "tab":
		if m.platIdx > 0 {
			m.platIdx--
			m.acctIdx = 0
		}
	case "right", "l", "shift+tab":
		if m.platIdx < len(m.overview)-1 {
			m.platIdx++
			m.acctIdx = 0
		}
	case "enter":
		return m.startSwitch()
	case "s":
		return m.doSave()
	case "a":
		m.screen = screenAddAccount
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink
	case "d":
		m.screen = screenPickDefault
		return m, nil
	case "r":
		refreshed, err := m.refreshOverview()
		if err != nil {
			m.screen = screenNotice
			m.noticeTitle = "Refresh failed"
			m.noticeBody = err.Error()
			m.noticeOK = false
			return refreshed, nil
		}
		return refreshed, nil
	}
	return m, nil
}

func (m appModel) updateAddAccount(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenDashboard
		return m, nil
	case "enter":
		label := strings.TrimSpace(m.input.Value())
		if label == "" {
			m.screen = screenNotice
			m.noticeTitle = "Add account"
			m.noticeBody = "Label cannot be empty"
			m.noticeOK = false
			return m, nil
		}
		plat := m.currentPlatform()
		if plat == nil {
			m.screen = screenDashboard
			return m, nil
		}
		var id paths.AccountID
		err := profiles.WithPlatform(plat.ID, func() error {
			var addErr error
			id, addErr = app.RegisterAccountFromLabel(label)
			return addErr
		})
		if err != nil {
			m.screen = screenNotice
			m.noticeTitle = "Add account"
			m.noticeBody = err.Error()
			m.noticeOK = false
			return m, nil
		}
		refreshed, err := m.refreshOverview()
		if err != nil {
			refreshed.screen = screenNotice
			refreshed.noticeTitle = "Refresh failed"
			refreshed.noticeBody = err.Error()
			refreshed.noticeOK = false
			return refreshed, nil
		}
		m = refreshed
		m.screen = screenNotice
		m.noticeTitle = "Account added"
		m.noticeBody = fmt.Sprintf("%s (%s) on %s", profiles.AccountLabelFor(plat.ID, id), id, plat.Name)
		m.noticeOK = true
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m appModel) updatePickDefault(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.screen = screenDashboard
		return m, nil
	case "up", "k":
		if m.platIdx > 0 {
			m.platIdx--
		}
	case "down", "j":
		if m.platIdx < len(m.overview)-1 {
			m.platIdx++
		}
	case "enter":
		plat := m.currentPlatform()
		if plat == nil {
			m.screen = screenDashboard
			return m, nil
		}
		if err := profiles.SetActivePlatform(plat.ID); err != nil {
			m.screen = screenNotice
			m.noticeTitle = "Default platform"
			m.noticeBody = err.Error()
			m.noticeOK = false
			return m, nil
		}
		refreshed, err := m.refreshOverview()
		if err != nil {
			refreshed.screen = screenNotice
			refreshed.noticeTitle = "Refresh failed"
			refreshed.noticeBody = err.Error()
			refreshed.noticeOK = false
			return refreshed, nil
		}
		m = refreshed
		m.screen = screenNotice
		m.noticeTitle = "Default platform"
		m.noticeBody = FormatPlatformSwitch(plat.Name)
		m.noticeOK = true
		return m, nil
	}
	return m, nil
}

func (m appModel) startSwitch() (tea.Model, tea.Cmd) {
	accounts := m.currentAccounts()
	if m.acctIdx < 0 || m.acctIdx >= len(accounts) {
		return m, nil
	}
	plat := m.currentPlatform()
	if plat == nil {
		return m, nil
	}
	acct := accounts[m.acctIdx]
	if !acct.Saved {
		m.screen = screenNotice
		m.noticeTitle = "Not saved"
		m.noticeBody = fmt.Sprintf("Save %s on %s first", acct.Label, plat.Name)
		m.noticeOK = false
		return m, nil
	}

	m.screen = screenSwitching
	m.switchAcct = acct.ID
	m.switchPlat = plat.ID
	m.switchStep = "Preparing..."
	m.switchErr = nil

	stepCh := make(chan string, 16)
	doneCh := make(chan error, 1)
	m.stepCh = stepCh
	m.doneCh = doneCh

	go func() {
		err := profiles.WithPlatform(plat.ID, func() error {
			return app.SwitchTo(acct.ID, app.SwitchOptions{}, func(label string) {
				stepCh <- label
			})
		})
		close(stepCh)
		doneCh <- err
	}()

	return m, tea.Batch(m.spinner.Tick, m.watchSwitch())
}

func (m appModel) watchSwitch() tea.Cmd {
	return func() tea.Msg {
		label, ok := <-m.stepCh
		if ok {
			return switchStepMsg{label: label}
		}
		return switchDoneMsg{err: <-m.doneCh}
	}
}

func (m appModel) doSave() (tea.Model, tea.Cmd) {
	accounts := m.currentAccounts()
	if m.acctIdx < 0 || m.acctIdx >= len(accounts) {
		return m, nil
	}
	plat := m.currentPlatform()
	if plat == nil {
		return m, nil
	}
	acct := accounts[m.acctIdx]

	var profile *profiles.Profile
	err := profiles.WithPlatform(plat.ID, func() error {
		var saveErr error
		profile, saveErr = app.SaveAs(acct.ID, "")
		return saveErr
	})
	if err != nil {
		m.screen = screenNotice
		m.noticeTitle = "Save failed"
		m.noticeBody = err.Error()
		m.noticeOK = false
		return m, nil
	}

	refreshed, err := m.refreshOverview()
	if err != nil {
		refreshed.screen = screenNotice
		refreshed.noticeTitle = "Refresh failed"
		refreshed.noticeBody = err.Error()
		refreshed.noticeOK = false
		return refreshed, nil
	}
	m = refreshed

	email := ""
	if profile.Email != nil {
		email = *profile.Email
	}
	m.screen = screenNotice
	m.noticeTitle = "Saved"
	m.noticeBody = FormatSaveSuccess(plat.ID, acct.ID, email)
	m.noticeOK = true
	return m, nil
}

func (m appModel) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case screenAddAccount:
		return m.viewAddAccount()
	case screenPickDefault:
		return m.viewPickDefault()
	case screenSwitching:
		return m.viewSwitching()
	case screenNotice:
		return m.viewNotice()
	default:
		return m.viewDashboard()
	}
}

func (m appModel) viewDashboard() string {
	sideW := sidebarWidth(m.width)
	mainW := contentWidth(m.width, sideW)

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	sidebar := m.renderSidebar(sideW, m.height-6)
	main := m.renderAccounts(mainW, m.height-6)
	row := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
	b.WriteString(row)
	b.WriteString("\n")
	b.WriteString(m.renderFooter())
	return appStyle.Render(b.String())
}

func (m appModel) renderHeader() string {
	var b strings.Builder
	b.WriteString(boxTitle("Account Switcher"))
	b.WriteString("\n")
	b.WriteString(appSubtitle.Render("Switch saved sessions across AI coding tools"))
	return headerBox.Render(b.String())
}

func (m appModel) renderSidebar(width, height int) string {
	var lines []string
	lines = append(lines, tableHeader.Render("PLATFORMS"))
	lines = append(lines, "")

	for i, plat := range m.overview {
		marker := "  "
		if i == m.platIdx {
			marker = activeBadge.Render("▸ ")
		}
		name := plat.Name
		if plat.IsDefault {
			name += " " + dimStyle.Render("*")
		}
		live := dimStyle.Render("out")
		if plat.LiveID != "" {
			live = activeBadge.Render("●")
		}
		line := fmt.Sprintf("%s%-14s %s", marker, truncate(name, 14), live)
		if i == m.platIdx {
			lines = append(lines, sidebarActive.Render(padRight(line, width-2)))
		} else {
			lines = append(lines, sidebarItem.Render(padRight(line, width-2)))
		}
		if plat.LiveID != "" {
			lines = append(lines, dimStyle.Render("  "+truncate(plat.LiveID, width-4)))
		}
	}

	body := strings.Join(lines, "\n")
	body = fitHeight(body, maxInt(4, height))
	return panelStyle.Width(width).Height(height).Render(body)
}

func (m appModel) renderAccounts(width, height int) string {
	plat := m.currentPlatform()
	if plat == nil {
		return panelStyle.Width(width).Height(height).Render(missingBadge.Render("No platform selected"))
	}

	var lines []string
	title := fmt.Sprintf("%s", plat.Name)
	if plat.IsDefault {
		title += dimStyle.Render("  default")
	}
	lines = append(lines, appTitle.Render(title))
	if plat.LiveID != "" {
		lines = append(lines, appSubtitle.Render("Live: "+plat.LiveID))
	} else {
		lines = append(lines, appSubtitle.Render("Not signed in"))
	}
	lines = append(lines, "")
	lines = append(lines, tableHeader.Render(
		padRight("ACCOUNT", 16)+padRight("STATUS", 12)+padRight("IDENTITY", width-32),
	))
	lines = append(lines, dimStyle.Render(strings.Repeat("─", width-4)))

	accounts := plat.Accounts
	if len(accounts) == 0 {
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("No accounts — press a to add one"))
	} else {
		for i, acct := range accounts {
			prefix := "  "
			style := rowNormal
			if i == m.acctIdx {
				prefix = "▸ "
				style = rowSelected
			}
			line := style.Render(
				padRight(prefix+acct.Label, 16) +
					padRight(accountStatusPlain(acct), 12) +
					truncate(accountIdentityPlain(acct), width-32),
			)
			lines = append(lines, line)
		}
	}

	body := strings.Join(lines, "\n")
	body = fitHeight(body, maxInt(6, height))
	return panelStyle.Width(width).Height(height).Render(body)
}

func accountStatusPlain(acct profiles.AccountOverview) string {
	switch {
	case acct.IsLive:
		return "active"
	case !acct.Saved:
		return "empty"
	case acct.IsActive:
		return "selected"
	default:
		return "saved"
	}
}

func accountIdentityPlain(acct profiles.AccountOverview) string {
	if !acct.Saved {
		return "—"
	}
	if acct.Profile != nil && acct.Profile.Email != nil {
		return *acct.Profile.Email
	}
	if acct.Profile != nil {
		return profiles.FormatSavedAt(acct.Profile.SavedAt)
	}
	return "saved"
}

func (m appModel) renderFooter() string {
	return footerBar.Render(renderKeyHelp([][2]string{
		{"←/→", "platform"},
		{"↑/↓", "account"},
		{"enter", "switch"},
		{"s", "save"},
		{"a", "add"},
		{"d", "default"},
		{"r", "refresh"},
		{"q", "quit"},
	}))
}

func (m appModel) viewAddAccount() string {
	plat := m.currentPlatform()
	title := "Add account"
	if plat != nil {
		title = fmt.Sprintf("Add account on %s", plat.Name)
	}
	var b strings.Builder
	b.WriteString(boxTitle(title))
	b.WriteString("\n\n")
	b.WriteString(appSubtitle.Render("Display label for the new profile slot"))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	b.WriteString(renderKeyHelp([][2]string{{"enter", "create"}, {"esc", "back"}}))
	return appStyle.Render(inputBox.Render(b.String()))
}

func (m appModel) viewPickDefault() string {
	var lines []string
	lines = append(lines, boxTitle("Default platform"))
	lines = append(lines, "")
	lines = append(lines, appSubtitle.Render("Used when --platform flag is omitted"))
	lines = append(lines, "")

	for i, plat := range m.overview {
		marker := "  "
		if plat.IsDefault {
			marker = activeBadge.Render("* ")
		}
		if i == m.platIdx {
			marker = activeBadge.Render("▸ ")
		}
		restart := "CLI only"
		if p, _ := platform.Get(plat.ID); p != nil && p.NeedsRestart() {
			restart = "restarts app"
		}
		lines = append(lines, fmt.Sprintf("%s%s  %s", marker, padRight(plat.Name, 14), dimStyle.Render(restart)))
	}

	var b strings.Builder
	b.WriteString(strings.Join(lines, "\n"))
	b.WriteString("\n\n")
	b.WriteString(renderKeyHelp([][2]string{{"enter", "set default"}, {"esc", "back"}}))
	return appStyle.Render(panelStyle.Render(b.String()))
}

func (m appModel) viewSwitching() string {
	plat := m.currentPlatform()
	label := string(m.switchAcct)
	if plat != nil {
		label = profiles.AccountLabelFor(m.switchPlat, m.switchAcct)
	}
	var b strings.Builder
	b.WriteString(boxTitle("Switching to " + label))
	if plat != nil {
		b.WriteString("\n")
		b.WriteString(appSubtitle.Render(plat.Name))
	}
	b.WriteString("\n\n")
	b.WriteString(m.spinner.View() + " ")
	b.WriteString(valueStyle.Render(m.switchStep))
	return appStyle.Render(b.String())
}

func (m appModel) viewNotice() string {
	icon := activeBadge.Render("✓")
	if !m.noticeOK {
		icon = missingBadge.Render("✗")
	}
	var b strings.Builder
	b.WriteString(icon + " " + appTitle.Render(m.noticeTitle))
	b.WriteString("\n\n")
	b.WriteString(m.noticeBody)
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("press any key"))
	return appStyle.Render(noticeBox.Render(b.String()))
}

var appStyle = lipgloss.NewStyle().Padding(1, 2)

type ActionKind int

const (
	ActionNone ActionKind = iota
	ActionQuit
	ActionSwitch
	ActionSave
	ActionAddAccount
	ActionSetDefaultPlatform
)

type MenuResult struct {
	Kind       ActionKind
	PlatformID platform.ID
	AccountID  paths.AccountID
}

func FormatSaveSuccess(platformID platform.ID, id paths.AccountID, email string) string {
	label := profiles.AccountLabelFor(platformID, id)
	p, _ := platform.Get(platformID)
	name := string(platformID)
	if p != nil {
		name = p.Name
	}
	if email == "" || email == "unknown" {
		return fmt.Sprintf("Saved %s on %s", label, name)
	}
	return fmt.Sprintf("Saved %s on %s (%s)", label, name, email)
}

func FormatPlatformSwitch(name string) string {
	return fmt.Sprintf("Default platform set to %s", name)
}
