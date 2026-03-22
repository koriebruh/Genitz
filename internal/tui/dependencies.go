package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	CatFramework     = "framework"
	CatORM           = "orm"
	CatDriver        = "driver"
	CatCache         = "cache"
	CatMessageBroker = "broker"
	CatRPC           = "rpc"
	CatLogger        = "logger"
	CatTracing       = "tracing"
	CatMetrics       = "metrics"
	CatAuth          = "auth"
	CatValidation    = "validation"
	CatDoc           = "documentation"
)

// getBadgeStyle returns a coloured pill style for a given dependency category.
func getBadgeStyle(category string) lipgloss.Style {
	base := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginLeft(1).
		Foreground(lipgloss.Color("#FFFFFF"))

	switch category {
	case CatFramework:
		return base.Background(lipgloss.Color("#00ADD8"))
	case CatORM:
		return base.Background(lipgloss.Color("#F7931E"))
	case CatDriver:
		return base.Background(lipgloss.Color("#4DB33D"))
	case CatCache:
		return base.Background(lipgloss.Color("#D82C20"))
	case CatMessageBroker:
		return base.Background(lipgloss.Color("#004E7A"))
	case CatRPC:
		return base.Background(lipgloss.Color("#00B5AD"))
	case CatLogger:
		return base.Background(lipgloss.Color("#555555"))
	case CatTracing:
		return base.Background(lipgloss.Color("#6B4E90"))
	case CatMetrics:
		return base.Background(lipgloss.Color("#FF4500"))
	case CatAuth:
		return base.Background(lipgloss.Color("#E91E63"))
	case CatValidation:
		return base.Background(lipgloss.Color("#8BC34A"))
	case CatDoc:
		return base.Background(lipgloss.Color("#3F51B5"))
	default:
		return base.Background(lipgloss.Color("#222222"))
	}
}

// Dependency describes a Go dependency the user can opt into.
type Dependency struct {
	ID          string
	Name        string
	Category    string
	ImportPath  string
	IsDefault   bool
	Requires    []string
	Description string
	TemplateDir string
}

// depsPath is the path prefix inside the embedded templates FS (see generator/embed.go).
var depsPath = "templates/feature/%s"

// DependencyRegistry is the list of selectable dependencies shown in StepDeps.
var DependencyRegistry = []Dependency{
	{
		ID: "redis", Name: "Redis", Category: CatCache,
		ImportPath:  "github.com/redis/go-redis/v9",
		Description: "Redis client for Go",
		TemplateDir: fmt.Sprintf(depsPath, "redis"),
	},
	{
		ID: "validator", Name: "Go Playground Validator", Category: CatValidation,
		ImportPath:  "github.com/go-playground/validator/v10",
		Description: "Struct and field validation via struct tags",
		TemplateDir: fmt.Sprintf(depsPath, "validator"),
	},
	{
		ID: "fiber", Name: "Fiber", Category: CatFramework,
		ImportPath:  "github.com/gofiber/fiber/v3",
		Description: "Express-inspired web framework written in Go",
		TemplateDir: fmt.Sprintf(depsPath, "fiber"),
	},
	{
		ID: "gin", Name: "Gin Gonic", Category: CatFramework,
		ImportPath:  "github.com/gin-gonic/gin",
		Description: "High-performance HTTP web framework",
		TemplateDir: fmt.Sprintf(depsPath, "gin"),
	},
	{
		ID: "gorm", Name: "GORM", Category: CatORM,
		ImportPath:  "gorm.io/gorm",
		Description: "The fantastic ORM library for Golang",
		TemplateDir: fmt.Sprintf(depsPath, "gorm"),
	},
	{
		ID: "zap", Name: "Uber Zap", Category: CatLogger,
		ImportPath:  "go.uber.org/zap",
		Description: "Blazing fast, structured, leveled logging",
		TemplateDir: fmt.Sprintf(depsPath, "zap"),
	},
}


// depGroups defines the display order and category membership for each group.
var depGroups = []struct {
	label      string
	categories []string
}{
	{"Web / Routing", []string{CatFramework, CatRPC}},
	{"Database", []string{CatORM, CatDriver}},
	{"Cache", []string{CatCache}},
	{"Messaging", []string{CatMessageBroker}},
	{"Observability", []string{CatLogger, CatTracing, CatMetrics}},
	{"Security", []string{CatAuth}},
	{"Utilities", []string{CatValidation, CatDoc}},
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
	for _, group := range depGroups {
		for i, dep := range m.Registry {
			if seen[i] {
				continue
			}
			for _, cat := range group.categories {
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

// findDepGroup returns the display label of the group a dependency belongs to.
func findDepGroup(dep Dependency) string {
	for _, group := range depGroups {
		for _, cat := range group.categories {
			if dep.Category == cat {
				return group.label
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

