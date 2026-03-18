package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	genitzStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7F56D9")).Bold(true)
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true).MarginBottom(1)
	selStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true)
	descStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Italic(true)
	checkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	nameStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
)

type Model struct {
	Step        Step
	FolderInput textinput.Model
	PkgInput    textinput.Model

	ArchOptions  []string
	ArchCursor   int
	SelectedArch string

	Registry []Dependency
	Chosen   map[int]struct{}
	Cursor   int

	Done bool
}

func InitialModel() Model {
	// Setup Folder Input
	f := textinput.New()
	f.Placeholder = "my-awesome-app"
	f.Focus()

	// Setup Package Input
	p := textinput.New()
	p.Placeholder = "github.com/username/repo"

	return Model{
		Step:        StepSplash,
		FolderInput: f,
		PkgInput:    p,
		ArchOptions: []string{ArchStandard, ArchMicro, ArchClean, ArchDDD, ArchCLI},
		Registry:    DependencyRegistry,
		Chosen:      make(map[int]struct{}),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			switch m.Step {
			case StepSplash:
				m.Step = StepFolder
				return m, nil
			case StepFolder:
				m.Step = StepPackage
				m.PkgInput.Focus()
				return m, nil
			case StepPackage:
				m.Step = StepArch
				return m, nil
			case StepArch:
				m.SelectedArch = m.ArchOptions[m.ArchCursor]
				m.Step = StepDeps
				return m, nil
			case StepDeps:
				m.Done = true
				return m, tea.Quit
			}
		}
	}

	// Logic navigasi per step
	switch m.Step {
	case StepFolder:
		m.FolderInput, cmd = m.FolderInput.Update(msg)
	case StepPackage:
		m.PkgInput, cmd = m.PkgInput.Update(msg)
	case StepArch:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "up", "k":
				if m.ArchCursor > 0 {
					m.ArchCursor--
				}
			case "down", "j":
				if m.ArchCursor < len(m.ArchOptions)-1 {
					m.ArchCursor++
				}
			}
		}
	case StepDeps:
		// Logic navigasi dependency (sama kayak kode sebelumnya)
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				}
			case "down", "j":
				if m.Cursor < len(m.Registry)-1 {
					m.Cursor++
				}
			case " ":
				if _, ok := m.Chosen[m.Cursor]; ok {
					delete(m.Chosen, m.Cursor)
				} else {
					m.Chosen[m.Cursor] = struct{}{}
				}
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if m.Done {
		return "\n🚀 Generating " + m.SelectedArch + " project...\n"
	}

	switch m.Step {
	case StepSplash:
		return RenderSplashView("1.20.3") // Bisa diganti dengan versi Go yang sebenarnya

	case StepFolder:
		return fmt.Sprintf(
			"📁 %s\n\n%s\n\n%s",
			genitzStyle.Render("Folder Name:"),
			m.FolderInput.View(),
			"(Enter to continue)",
		)

	case StepPackage:
		return fmt.Sprintf(
			"📦 %s\n\n%s\n\n%s",
			genitzStyle.Render("Package Name:"),
			m.PkgInput.View(),
			"(Enter to continue)",
		)

	case StepArch:
		var s strings.Builder
		s.WriteString("🏗️  " + genitzStyle.Render("Choose Architecture:") + "\n\n")
		for i, opt := range m.ArchOptions {
			cursor := "  "
			if m.ArchCursor == i {
				cursor = cursorStyle.Render("> ")
				s.WriteString(fmt.Sprintf("%s%s\n", cursor, selStyle.Render(opt)))
			} else {
				s.WriteString(fmt.Sprintf("%s%s\n", cursor, opt))
			}
		}
		return s.String()

	case StepDeps:
		// View dependency yang sudah kita buat sebelumnya
		return m.renderDependencyView()
	}

	return ""
}
