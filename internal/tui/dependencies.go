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

var depsPath = "internal/generator/templetes/feature/%s"

// DependencyRegistry is the list of selectable dependencies shown in StepDeps.
var DependencyRegistry = []Dependency{
	{
		ID: "redis", Name: "redis", Category: CatCache,
		ImportPath:  "github.com/redis/go-redis/v9",
		Description: "Redis client for Go",
		TemplateDir: fmt.Sprintf(depsPath, "redis"),
	},
	{
		ID: "validator", Name: "go playground validator", Category: CatValidation,
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
		TemplateDir: "",
	},
	{
		ID: "gorm", Name: "GORM", Category: CatORM,
		ImportPath:  "gorm.io/gorm",
		Description: "The fantastic ORM library for Golang",
		TemplateDir: "",
	},
	{
		ID: "zap", Name: "Uber Zap", Category: CatLogger,
		ImportPath:  "go.uber.org/zap",
		Description: "Blazing fast, structured, leveled logging",
		TemplateDir: "",
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

// renderDependencyView renders the StepDeps panel body grouped by category.
// Navigation cursor tracks visual position via buildVisibleOrder, so it never
// jumps across group boundaries.
func (m *Model) renderDependencyView() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("DEPENDENCIES") + "\n")
	b.WriteString(styles.PanelHint.Render("Toggle packages, then press enter to review") + "\n\n")

	// ── Search bar ────────────────────────────────────────────
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
		// Track visual position across all items for cursor matching.
		visualPos := 0

		for _, group := range depGroups {
			// Collect visible indices that belong to this group, in order.
			var groupIndices []int
			for _, idx := range visibleOrder {
				dep := m.Registry[idx]
				for _, cat := range group.categories {
					if dep.Category == cat {
						groupIndices = append(groupIndices, idx)
						break
					}
				}
			}
			if len(groupIndices) == 0 {
				continue
			}

			b.WriteString(groupHeaderStyle.Render("▸ "+group.label) + "\n")

			for _, i := range groupIndices {
				dep := m.Registry[i]
				isActive := m.Cursor == visualPos

				cursor := "   "
				if isActive {
					cursor = styles.Cursor.Render(" ▶ ")
				}

				_, chosen := m.Chosen[i]
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
				b.WriteString(fmt.Sprintf("      %s\n", styles.Description.Render(dep.Description)))

				visualPos++
			}

			b.WriteRune('\n')
		}
	}

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
	b.WriteString(renderKeyHints(hints))
	return b.String()
}
