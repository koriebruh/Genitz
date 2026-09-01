package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleDockerKeys handles StepDocker (Init mode only): a single toggle,
// nothing else to navigate.
func (m *Model) handleDockerKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case " ", "y", "n":
		if msg.String() == "y" {
			m.IncludeDocker = true
		} else if msg.String() == "n" {
			m.IncludeDocker = false
		} else {
			m.IncludeDocker = !m.IncludeDocker
		}
		return nil, true
	case "enter":
		m.Step = StepReview
		return nil, true
	case "b":
		m.Step = StepDeps
		return nil, true
	case "q":
		return tea.Quit, true
	}
	return nil, false
}

// dockerPreview returns the compose service names implied by the currently
// selected dependencies, via the closure main.go injected (LookupComposeServices)
// so this package never needs to import generator's compose-mapping table.
func (m *Model) dockerPreview() []string {
	if m.LookupComposeServices == nil {
		return nil
	}
	ids := make([]string, 0, len(m.Chosen))
	for _, dep := range m.Chosen {
		ids = append(ids, dep.ID)
	}
	return m.LookupComposeServices(ids)
}

// viewDocker renders the StepDocker toggle screen.
func (m *Model) viewDocker() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("DOCKER") + "\n")
	b.WriteString(styles.PanelHint.Render("Optional — a multistage Dockerfile, and a docker-compose.yml if it's warranted") + "\n\n")

	check := styles.Description.Render("[ ] ")
	label := styles.Name.Render("Include Docker setup")
	if m.IncludeDocker {
		check = styles.Checkbox.Render("[✓] ")
		label = styles.Selected.Render("Include Docker setup")
	}
	b.WriteString("  " + check + label + "\n\n")

	if m.IncludeDocker {
		b.WriteString(styles.Description.Render("  Will generate: Dockerfile, .dockerignore") + "\n")
		if services := m.dockerPreview(); len(services) > 0 {
			b.WriteString(styles.Description.Render("  docker-compose.yml will include: "+strings.Join(services, ", ")) + "\n")
		} else {
			b.WriteString(styles.Description.Render("  docker-compose.yml: skipped — none of your selected dependencies need a container") + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(renderKeyHints([]keyHint{
		{"space / y / n", "toggle"},
		{"enter", "continue"},
		{"b", "back"},
		{"q", "quit"},
	}))
	return b.String()
}
