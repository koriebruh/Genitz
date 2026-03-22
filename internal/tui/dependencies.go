package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/koriebruh/Genitz/internal/registry"
)

// getBadgeStyle returns a coloured pill style for a given dependency category.
func getBadgeStyle(category string) lipgloss.Style {
	base := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginLeft(1).
		Foreground(lipgloss.Color("#FFFFFF"))

	switch category {
	case registry.CatFramework:
		return base.Background(lipgloss.Color("#00ADD8"))
	case registry.CatORM:
		return base.Background(lipgloss.Color("#F7931E"))
	case registry.CatDriver:
		return base.Background(lipgloss.Color("#4DB33D"))
	case registry.CatCache:
		return base.Background(lipgloss.Color("#D82C20"))
	case registry.CatMessageBroker:
		return base.Background(lipgloss.Color("#004E7A"))
	case registry.CatRPC:
		return base.Background(lipgloss.Color("#00B5AD"))
	case registry.CatLogger:
		return base.Background(lipgloss.Color("#555555"))
	case registry.CatTracing:
		return base.Background(lipgloss.Color("#6B4E90"))
	case registry.CatMetrics:
		return base.Background(lipgloss.Color("#FF4500"))
	case registry.CatAuth:
		return base.Background(lipgloss.Color("#E91E63"))
	case registry.CatValidation:
		return base.Background(lipgloss.Color("#8BC34A"))
	case registry.CatDoc:
		return base.Background(lipgloss.Color("#3F51B5"))
	default:
		return base.Background(lipgloss.Color("#222222"))
	}
}

// groupHeaderStyle is the amber label rendered above each category section.
var groupHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#E3B341")).
	Bold(true)

// buildVisibleOrder returns Registry indices in visual group order, optionally
// filtered by m.SearchQuery (case-insensitive substring match on name/category).
// This is the canonical order for Cursor navigation — Cursor is a position
// inside this slice, not a raw Registry index.
func (m *Model) buildVisibleOrder() []int {
	query := strings.ToLower(strings.TrimSpace(m.SearchQuery))
	var order []int
	seen := make(map[int]bool)
	for _, group := range registry.DepGroups {
		for i, dep := range m.Registry {
			if seen[i] {
				continue
			}
			for _, cat := range group.Categories {
				if dep.Category != cat {
					continue
				}
				if query == "" ||
					strings.Contains(strings.ToLower(dep.Name), query) ||
					strings.Contains(strings.ToLower(dep.Category), query) ||
					strings.Contains(strings.ToLower(dep.Description), query) {
					order = append(order, i)
					seen[i] = true
				}
				break
			}
		}
	}
	return order
}

func findDepGroup(dep registry.Dependency) string {
	for _, group := range registry.DepGroups {
		for _, cat := range group.Categories {
			if dep.Category == cat {
				return group.Label
			}
		}
	}
	return "Other"
}

// renderDependencyView renders the StepDeps panel body with scroll support.
// Only (depsMaxVisible) items are rendered at once; scroll indicators show
// how many items are hidden above/below.
func (m *Model) renderDependencyView() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("DEPENDENCIES") + "\n")
	b.WriteString(styles.PanelHint.Render("Toggle with space · enter to review · / to search") + "\n\n")

	// ── Search bar ────────────────────────────────────────────────────────────
	if m.SearchActive || m.SearchQuery != "" {
		indicator := styles.Description.Render("/")
		query := styles.Selected.Render(m.SearchQuery)
		cursor := ""
		if m.SearchActive {
			cursor = styles.Cursor.Render("▌")
		}
		b.WriteString(indicator + " " + query + cursor + "\n\n")
	} else {
		b.WriteString(styles.Description.Render("  press / to search") + "\n\n")
	}

	visibleOrder := m.buildVisibleOrder()

	if len(visibleOrder) == 0 {
		b.WriteString(styles.Description.Render("  no results for \""+m.SearchQuery+"\"") + "\n\n")
	} else {
		maxVis := m.depsMaxVisible()
		start := m.DepsOffset
		if start >= len(visibleOrder) {
			start = 0
		}
		end := start + maxVis
		if end > len(visibleOrder) {
			end = len(visibleOrder)
		}

		// ── Scroll up indicator ───────────────────────────────────────────────
		if start > 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Italic(true).
				Render(fmt.Sprintf("  ↑ %d more above", start)) + "\n")
		}

		// ── Render visible window of items ────────────────────────────────────
		prevGroup := ""
		for pos := start; pos < end; pos++ {
			idx := visibleOrder[pos]
			dep := m.Registry[idx]

			// Insert group header when the group changes within the visible window
			depGroup := findDepGroup(dep)
			if depGroup != prevGroup {
				if pos > start {
					b.WriteRune('\n')
				}
				b.WriteString(groupHeaderStyle.Render("  ▸ "+depGroup) + "\n")
				prevGroup = depGroup
			}

			isActive := m.Cursor == pos

			cursor := "   "
			if isActive {
				cursor = styles.Cursor.Render(" ▶ ")
			}

			_, chosen := m.Chosen[idx]
			check := styles.Description.Render("[ ] ")
			if chosen {
				check = styles.Checkbox.Render("[✓] ")
			}

			name := styles.Name.Render(dep.Name)
			if isActive {
				name = styles.Selected.Render(dep.Name)
			}

			badge := getBadgeStyle(dep.Category).Render(strings.ToUpper(dep.Category))

			b.WriteString(fmt.Sprintf("%s%s%s%s\n", cursor, check, name, badge))
			b.WriteString(fmt.Sprintf("      %s\n",
				styles.Description.Render(dep.Description)))
		}

		// ── Scroll down indicator ─────────────────────────────────────────────
		if end < len(visibleOrder) {
			b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Italic(true).
				Render(fmt.Sprintf("  ↓ %d more below", len(visibleOrder)-end)) + "\n")
		}

		// ── Item counter ──────────────────────────────────────────────────────
		counter := fmt.Sprintf("  %d/%d  ·  %d selected",
			m.Cursor+1, len(visibleOrder), len(m.Chosen))
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(colorMuted).Render(counter) + "\n")
	}

	// ── Footer hints ──────────────────────────────────────────────────────────
	hints := []keyHint{
		{"↑↓ / jk", "navigate"},
		{"space", "toggle"},
		{"enter", "review"},
	}
	if m.SearchActive {
		hints = append(hints, keyHint{"esc", "close search"})
	} else {
		hints = append(hints, keyHint{"/", "search"})
		hints = append(hints, keyHint{"q", "quit"})
	}
	b.WriteString("\n")
	b.WriteString(renderKeyHints(hints))
	return b.String()
}

