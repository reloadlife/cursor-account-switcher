package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/platform"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

type progressModel struct {
	spinner spinner.Model
	current string
	done    bool
	err     error
	account paths.AccountID
	stepCh  <-chan string
	doneCh  <-chan error
}

type progressStepMsg struct {
	label string
}

type switchFinishedMsg struct {
	err error
}

func RunSwitch(id paths.AccountID) error {
	stepCh := make(chan string, 16)
	doneCh := make(chan error, 1)

	go func() {
		doneCh <- app.SwitchTo(id, func(label string) {
			stepCh <- label
		})
		close(stepCh)
	}()

	m := progressModel{
		account: id,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		stepCh:  stepCh,
		doneCh:  doneCh,
		current: "Preparing...",
	}
	m.spinner.Style = lipgloss.NewStyle().Foreground(colorAccent)

	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}

	model := final.(progressModel)
	return model.err
}

func (m progressModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		nextStep(m.stepCh, m.doneCh),
	)
}

func nextStep(stepCh <-chan string, doneCh <-chan error) tea.Cmd {
	return func() tea.Msg {
		label, ok := <-stepCh
		if ok {
			return progressStepMsg{label: label}
		}
		return switchFinishedMsg{err: <-doneCh}
	}
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || (m.done && msg.String() != "") {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progressStepMsg:
		m.current = msg.label
		return m, nextStep(m.stepCh, m.doneCh)

	case switchFinishedMsg:
		m.done = true
		m.err = msg.err
		if msg.err != nil {
			m.current = msg.err.Error()
		}
		return m, nil
	}

	return m, nil
}

func (m progressModel) View() string {
	var b strings.Builder
	label := profiles.AccountLabel(m.account)

	b.WriteString(boxTitle(fmt.Sprintf("Switching to %s", label)))
	p, _ := platform.CurrentPlatform()
	if p != nil {
		b.WriteString("\n")
		b.WriteString(appSubtitle.Render(p.Name))
	}
	b.WriteString("\n\n")

	if m.done {
		if m.err != nil {
			b.WriteString(missingBadge.Render("✗ " + m.current))
		} else {
			b.WriteString(activeBadge.Render("✓ " + m.current))
		}
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("press any key"))
		return appStyle.Render(noticeBox.Render(b.String()))
	}

	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(valueStyle.Render(m.current))
	return appStyle.Render(b.String())
}
