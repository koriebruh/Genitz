package registry

import "fmt"

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

// DependencyRegistry is the list of selectable dependencies available in Genitz.
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

// DepGroups defines the display order and category membership for each group.
var DepGroups = []struct {
	Label      string
	Categories []string
}{
	{"Web / Routing", []string{CatFramework, CatRPC}},
	{"Database", []string{CatORM, CatDriver}},
	{"Cache", []string{CatCache}},
	{"Messaging", []string{CatMessageBroker}},
	{"Observability", []string{CatLogger, CatTracing, CatMetrics}},
	{"Security", []string{CatAuth}},
	{"Utilities", []string{CatValidation, CatDoc}},
}

// GetDependencyByID helps find a dependency for headless addition
func GetDependencyByID(id string) (*Dependency, error) {
	for _, dep := range DependencyRegistry {
		if dep.ID == id {
			return &dep, nil
		}
	}
	return nil, fmt.Errorf("dependency %q not found", id)
}
