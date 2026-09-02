package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// extraItem is one checkbox on the StepExtras screen.
type extraItem struct {
	label       string
	description string
	get         func(*Model) bool
	set         func(*Model, bool)
}

var extraItems = []extraItem{
	{
		label:       "GitHub Actions CI",
		description: "build, vet, test, gofmt check on every push",
		get:         func(m *Model) bool { return m.IncludeCI },
		set:         func(m *Model, v bool) { m.IncludeCI = v },
	},
	{
		label:       "Makefile",
		description: "build/test/run/fmt/vet targets",
		get:         func(m *Model) bool { return m.IncludeMakefile },
		set:         func(m *Model, v bool) { m.IncludeMakefile = v },
	},
	{
		label:       "git init + first commit",
		description: "local only — publishing to GitHub is suggested, not automatic",
		get:         func(m *Model) bool { return m.IncludeGitInit },
		set:         func(m *Model, v bool) { m.IncludeGitInit = v },
	},
}

// handleExtrasKeys handles StepExtras (Init mode only): a short checkbox list.
func (m *Model) handleExtrasKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "up", "k":
		if m.ExtrasCursor > 0 {
			m.ExtrasCursor--
		}
		return nil, true
	case "down", "j":
		if m.ExtrasCursor < len(extraItems)-1 {
			m.ExtrasCursor++
		}
		return nil, true
	case " ":
		item := extraItems[m.ExtrasCursor]
		item.set(m, !item.get(m))
		return nil, true
	case "enter":
		m.Step = StepReview
		return nil, true
	case "b":
		m.Step = StepDocker
		return nil, true
	case "q":
		return tea.Quit, true
	}
	return nil, false
}

// viewExtras renders the StepExtras checkbox list.
func (m *Model) viewExtras() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("EXTRAS") + "\n")
	b.WriteString(styles.PanelHint.Render("Optional project scaffolding") + "\n\n")

	for i, item := range extraItems {
		isActive := m.ExtrasCursor == i

		cursor := "   "
		if isActive {
			cursor = styles.Cursor.Render(" ▶ ")
		}

		check := styles.Description.Render("[ ] ")
		if item.get(m) {
			check = styles.Checkbox.Render("[✓] ")
		}

		label := styles.Name.Render(item.label)
		if isActive {
			label = styles.Selected.Render(item.label)
		}

		b.WriteString(cursor + check + label + "\n")
		b.WriteString("      " + styles.Description.Render(item.description) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderKeyHints([]keyHint{
		{"↑↓ / jk", "navigate"},
		{"space", "toggle"},
		{"enter", "continue"},
		{"b", "back"},
		{"q", "quit"},
	}))
	return b.String()
}
