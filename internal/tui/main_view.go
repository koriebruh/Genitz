package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Step Step

	FolderInput textinput.Model
	PkgInput    textinput.Model

	ArchOptions  []string
	ArchCursor   int
	SelectedArch string

	Registry []Dependency
	Chosen   map[int]Dependency
	Cursor   int

	Done bool
}

func InitialModel() *Model {
	f := textinput.New()
	f.Placeholder = "my-awesome-app"

	p := textinput.New()
	p.Placeholder = "github.com/username/repo"

	return &Model{
		Step:        StepSplash,
		FolderInput: f,
		PkgInput:    p,
		ArchOptions: []string{ArchMicro, ArchClean},
		Registry:    DependencyRegistry,
		Chosen:      make(map[int]Dependency),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if cmd, handled := m.handleGlobalKeys(keyMsg); handled {
			return m, cmd
		}
		if cmd, handled := m.handleStepKeys(keyMsg); handled {
			return m, cmd
		}
	}

	return m, m.updateStep(msg)
}

func (m *Model) View() string {
	if m.Done {
		return "\n🚀 Generating " + m.SelectedArch + " project...\n"
	}

	switch m.Step {
	case StepSplash:
		return RenderSplashView()
	case StepFolder:
		return m.viewFolder()
	case StepPackage:
		return m.viewPackage()
	case StepArch:
		return m.viewArchitecture()
	case StepDeps:
		return m.renderDependencyView()
	}

	return ""
}

func (m *Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c", "q":
		return tea.Quit, true
	}
	return nil, false
}

func (m *Model) handleStepKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch m.Step {
	case StepSplash:
		return m.handleSplashKeys(msg)
	case StepFolder:
		return m.handleFolderKeys(msg)
	case StepPackage:
		return m.handlePackageKeys(msg)
	case StepArch:
		return m.handleArchKeys(msg)
	case StepDeps:
		return m.handleDepsKeys(msg)
	}
	return nil, false
}

func (m *Model) updateStep(msg tea.Msg) tea.Cmd {
	switch m.Step {
	case StepFolder:
		return m.updateFolderInput(msg)
	case StepPackage:
		return m.updatePackageInput(msg)
	}
	return nil
}

func (m *Model) handleSplashKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.String() == "enter" {
		m.Step = StepFolder
		m.FolderInput.Focus()
		return nil, true
	}
	return nil, false
}

func (m *Model) handleFolderKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.String() == "enter" {
		m.FolderInput.Blur()
		m.PkgInput.Focus()
		m.Step = StepPackage
		return nil, true
	}
	return nil, false
}

func (m *Model) handlePackageKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.String() == "enter" {
		m.PkgInput.Blur()
		m.Step = StepArch
		return nil, true
	}
	return nil, false
}

func (m *Model) handleArchKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "up", "k":
		if m.ArchCursor > 0 {
			m.ArchCursor--
		}
		return nil, true
	case "down", "j":
		if m.ArchCursor < len(m.ArchOptions)-1 {
			m.ArchCursor++
		}
		return nil, true
	case "enter":
		m.SelectedArch = m.ArchOptions[m.ArchCursor]
		m.Step = StepDeps
		return nil, true
	}
	return nil, false
}

func (m *Model) handleDepsKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
		return nil, true
	case "down", "j":
		if m.Cursor < len(m.Registry)-1 {
			m.Cursor++
		}
		return nil, true
	case " ":
		m.toggleDependency(m.Cursor)
		return nil, true
	case "enter":
		m.Done = true
		return tea.Quit, true
	}
	return nil, false
}

func (m *Model) updateFolderInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.FolderInput, cmd = m.FolderInput.Update(msg)
	return cmd
}

func (m *Model) updatePackageInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.PkgInput, cmd = m.PkgInput.Update(msg)
	return cmd
}

func (m *Model) toggleDependency(index int) {
	if _, ok := m.Chosen[index]; ok {
		delete(m.Chosen, index)
		return
	}
	// Keep the dependency metadata so downstream generators know what to install.
	m.Chosen[index] = m.Registry[index]
}

func (m *Model) viewFolder() string {
	return fmt.Sprintf(
		"📁 %s\n\n%s\n\n%s",
		styles.Brand.Render("Folder Name:"),
		m.FolderInput.View(),
		"(Enter to continue)",
	)
}

func (m *Model) viewPackage() string {
	return fmt.Sprintf(
		"📦 %s\n\n%s\n\n%s",
		styles.Brand.Render("Package Name:"),
		m.PkgInput.View(),
		"(Enter to continue)",
	)
}

func (m *Model) viewArchitecture() string {
	var b strings.Builder
	b.WriteString("🏗️  " + styles.Brand.Render("Choose Architecture:") + "\n\n")

	for i, opt := range m.ArchOptions {
		cursor := "  "
		if m.ArchCursor == i {
			cursor = styles.Cursor.Render("> ")
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, styles.Selected.Render(opt)))
			continue
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, opt))
	}

	return b.String()
}
