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
	{
		label:       "README.md",
		description: "project name, module path, quickstart, dependency list",
		get:         func(m *Model) bool { return m.IncludeReadme },
		set:         func(m *Model, v bool) { m.IncludeReadme = v },
	},
	{
		label:       "Community files",
		description: "CONTRIBUTING.md, SECURITY.md, issue templates, dependabot.yml",
		get:         func(m *Model) bool { return m.IncludeCommunityFiles },
		set:         func(m *Model, v bool) { m.IncludeCommunityFiles = v },
	},
}

// licenseChoices cycles through on the trailing StepExtras row — "" (none)
// first so the default (nothing selected) is the first press away from
// wherever the cycle currently sits.
var licenseChoices = []string{"", "mit", "apache-2.0"}

// nextLicenseChoice returns the next value in licenseChoices after current,
// wrapping back to the start — used by the License row's space/enter cycle.
func nextLicenseChoice(current string) string {
	for i, c := range licenseChoices {
		if c == current {
			return licenseChoices[(i+1)%len(licenseChoices)]
		}
	}
	return licenseChoices[0]
}

func licenseLabel(choice string) string {
	if choice == "" {
		return "none"
	}
	return choice
}

// handleExtrasKeys handles StepExtras (Init mode only): a short checkbox list.
func (m *Model) handleExtrasKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	licenseRow := len(extraItems)
	switch msg.String() {
	case "up", "k":
		if m.ExtrasCursor > 0 {
			m.ExtrasCursor--
		}
		return nil, true
	case "down", "j":
		if m.ExtrasCursor < licenseRow {
			m.ExtrasCursor++
		}
		return nil, true
	case " ":
		if m.ExtrasCursor == licenseRow {
			m.LicenseChoice = nextLicenseChoice(m.LicenseChoice)
			return nil, true
		}
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

	licenseRow := len(extraItems)
	isActive := m.ExtrasCursor == licenseRow
	cursor := "   "
	if isActive {
		cursor = styles.Cursor.Render(" ▶ ")
	}
	licenseText := "License: " + licenseLabel(m.LicenseChoice)
	label := styles.Name.Render(licenseText)
	if isActive {
		label = styles.Selected.Render(licenseText)
	}
	b.WriteString(cursor + label + "\n")
	b.WriteString("      " + styles.Description.Render("space to cycle: none / MIT / Apache-2.0") + "\n")

	b.WriteString("\n")
	b.WriteString(renderKeyHints([]keyHint{
		{"↑↓ / jk", "navigate"},
		{"space", "toggle / cycle"},
		{"enter", "continue"},
		{"b", "back"},
		{"q", "quit"},
	}))
	return b.String()
}
