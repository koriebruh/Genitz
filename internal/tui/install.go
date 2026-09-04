package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// InstallStep is one unit of install work (a go mod/go get/go fmt call)
// the TUI runs and animates progress for. Run is expected to block until
// the underlying command finishes.
type InstallStep struct {
	Label string
	Run   func() error
}

// stepDoneMsg reports the result of running the InstallStep at InstallIndex.
type stepDoneMsg struct{ err error }

// newSpinner builds the spinner used on the Installing screen.
func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = spinnerStyle
	return s
}

// beginInstall builds the step list via BuildSteps, resets install state,
// and kicks off the first step plus the spinner animation.
func (m *Model) beginInstall() tea.Cmd {
	if m.BuildSteps == nil {
		m.Done = true
		return tea.Quit
	}
	m.InstallSteps = m.BuildSteps(m)
	m.InstallIndex = 0
	m.InstallErr = nil
	m.Step = StepInstalling
	if len(m.InstallSteps) == 0 {
		m.Done = true
		return tea.Quit
	}
	return tea.Batch(m.Spinner.Tick, m.runStepCmd())
}

// runStepCmd runs the current InstallStep in the background and reports
// its result as a stepDoneMsg.
func (m *Model) runStepCmd() tea.Cmd {
	step := m.InstallSteps[m.InstallIndex]
	return func() tea.Msg {
		return stepDoneMsg{err: step.Run()}
	}
}

// handleStepDone advances past a successful step (running the next one, or
// finishing) or records a failure for display.
func (m *Model) handleStepDone(msg stepDoneMsg) (tea.Cmd, bool) {
	if msg.err != nil {
		m.InstallErr = msg.err
		return tea.Quit, true
	}
	m.InstallIndex++
	if m.InstallIndex >= len(m.InstallSteps) {
		m.Done = true
		return tea.Quit, true
	}
	return m.runStepCmd(), true
}

// viewInstalling renders the animated install progress screen: completed
// steps get a checkmark, the current one spins, a failure gets a cross plus
// the captured command output.
func (m *Model) viewInstalling() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("INSTALLING") + "\n")
	b.WriteString(styles.PanelHint.Render("Running go mod / go get — sit tight") + "\n\n")
	b.WriteString(m.renderMascot() + "\n")

	for i, step := range m.InstallSteps {
		switch {
		case i < m.InstallIndex:
			b.WriteString("  " + styles.Checkbox.Render("✔") + " " + step.Label + "\n")
		case i == m.InstallIndex && m.InstallErr == nil:
			b.WriteString("  " + m.Spinner.View() + " " + styles.Selected.Render(step.Label) + "\n")
		case i == m.InstallIndex && m.InstallErr != nil:
			b.WriteString("  " + installFailStyle.Render("✖") + " " + step.Label + "\n")
		default:
			b.WriteString("  " + styles.Description.Render(step.Label) + "\n")
		}
	}

	if m.InstallErr != nil {
		b.WriteString("\n" + errLine(m.InstallErr.Error()))
		b.WriteString("\n")
		b.WriteString(renderKeyHints([]keyHint{{"ctrl+c", "quit"}}))
	}

	return b.String()
}
