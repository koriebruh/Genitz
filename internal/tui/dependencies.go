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

var DependencyRegistry = []Dependency{
	{
		ID: "redis", Name: "redis", Category: CatCache,
		ImportPath:  "github.com/redis/go-redis/v9",
		Description: "redis client for Go",
		TemplateDir: fmt.Sprintf(depsPath, "redis"),
	},
	{
		ID: "validator", Name: "go playground validator", Category: CatValidation,
		ImportPath:  "github.com/go-playground/validator/v10",
		Description: "Package validator implements value validations for structs and individual fields based on tags.",
		TemplateDir: fmt.Sprintf(depsPath, "validator"),
	},
	{
		ID: "fiber", Name: "Fiber", Category: CatFramework,
		ImportPath:  "github.com/gofiber/fiber/v3",
		Description: "Express inspired web framework written in Go",
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

func (m *Model) renderDependencyView() string {
	var s strings.Builder

	// Judul Section
	s.WriteString(styles.Title.Render("🚀 GENITZ: Select Dependencies") + "\n")

	for i, dep := range m.Registry {
		// 1. Render Kursor
		cursor := "  "
		if m.Cursor == i {
			cursor = styles.Cursor.Render("> ")
		}

		// 2. Render Checkbox
		checked := " [ ] "
		if _, ok := m.Chosen[i]; ok {
			checked = styles.Checkbox.Render(" [x] ")
		}

		// 3. Render Nama (Bold/Selected)
		name := styles.Name.Render(dep.Name)
		if m.Cursor == i {
			name = styles.Selected.Render(dep.Name)
		}

		// 4. Render Badge Kategori
		badge := getBadgeStyle(dep.Category).Render(strings.ToUpper(dep.Category))

		// Baris Utama
		s.WriteString(fmt.Sprintf("%s%s%s%s\n", cursor, checked, name, badge))

		// 5. Render Deskripsi
		s.WriteString(fmt.Sprintf("      %s\n", styles.Description.Render(dep.Description)))
	}

	s.WriteString("\n(Space: Toggle, Enter: Finish, Q: Quit)\n")
	return s.String()
}
