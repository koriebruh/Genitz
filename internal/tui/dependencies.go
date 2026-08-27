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
	CatDI            = "di"
	CatConfig        = "config"
	CatMigration     = "migration"
	CatUtility       = "utility"
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
	case CatDI:
		return base.Background(lipgloss.Color("#9C27B0"))
	case CatConfig:
		return base.Background(lipgloss.Color("#795548"))
	case CatMigration:
		return base.Background(lipgloss.Color("#37474F"))
	case CatUtility:
		return base.Background(lipgloss.Color("#009688"))
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
	Description string
}

// DependencyRegistry is the list of selectable dependencies shown in StepDeps.
var DependencyRegistry = []Dependency{
	// ── Web / Routing ────────────────────────────────────────────
	{
		ID: "fiber", Name: "Fiber", Category: CatFramework,
		ImportPath:  "github.com/gofiber/fiber/v3",
		Description: "Express-inspired web framework written in Go",
	},
	{
		ID: "gin", Name: "Gin Gonic", Category: CatFramework,
		ImportPath:  "github.com/gin-gonic/gin",
		Description: "High-performance HTTP web framework",
	},
	{
		ID: "echo", Name: "Echo", Category: CatFramework,
		ImportPath:  "github.com/labstack/echo/v4",
		Description: "Minimalist, high-performance web framework",
	},
	{
		ID: "chi", Name: "Chi", Category: CatFramework,
		ImportPath:  "github.com/go-chi/chi/v5",
		Description: "Lightweight, idiomatic, composable HTTP router",
	},
	{
		ID: "grpc", Name: "gRPC", Category: CatRPC,
		ImportPath:  "google.golang.org/grpc",
		Description: "High-performance RPC framework for service-to-service calls",
	},
	{
		ID: "protobuf", Name: "Protocol Buffers", Category: CatRPC,
		ImportPath:  "google.golang.org/protobuf",
		Description: "Protocol Buffers codegen runtime, pairs with gRPC",
	},

	// ── Database ─────────────────────────────────────────────────
	{
		ID: "gorm", Name: "GORM", Category: CatORM,
		ImportPath:  "gorm.io/gorm",
		Description: "The fantastic ORM library for Golang",
	},
	{
		ID: "sqlx", Name: "sqlx", Category: CatORM,
		ImportPath:  "github.com/jmoiron/sqlx",
		Description: "database/sql extensions — struct scanning for raw queries",
	},
	{
		ID: "mysql-driver", Name: "MySQL Driver", Category: CatDriver,
		ImportPath:  "github.com/go-sql-driver/mysql",
		Description: "MySQL driver for database/sql — raw query access",
	},
	{
		ID: "postgres-driver", Name: "pgx (PostgreSQL)", Category: CatDriver,
		ImportPath:  "github.com/jackc/pgx/v5",
		Description: "PostgreSQL driver/toolkit — raw query access",
	},
	{
		ID: "mongo-driver", Name: "MongoDB Driver", Category: CatDriver,
		ImportPath:  "go.mongodb.org/mongo-driver",
		Description: "Official MongoDB driver — NoSQL document store",
	},
	{
		ID: "migrate", Name: "golang-migrate", Category: CatMigration,
		ImportPath:  "github.com/golang-migrate/migrate/v4",
		Description: "Database schema migrations (MySQL, Postgres, Mongo, ...)",
	},

	// ── Cache ────────────────────────────────────────────────────
	{
		ID: "redis", Name: "redis", Category: CatCache,
		ImportPath:  "github.com/redis/go-redis/v9",
		Description: "Redis client — cache, sessions, or NoSQL key-value store",
	},

	// ── Messaging ────────────────────────────────────────────────
	{
		ID: "kafka-go", Name: "kafka-go", Category: CatMessageBroker,
		ImportPath:  "github.com/segmentio/kafka-go",
		Description: "Kafka client library for event streaming",
	},
	{
		ID: "nats", Name: "NATS", Category: CatMessageBroker,
		ImportPath:  "github.com/nats-io/nats.go",
		Description: "Lightweight, high-throughput pub/sub messaging",
	},
	{
		ID: "rabbitmq", Name: "RabbitMQ (amqp091)", Category: CatMessageBroker,
		ImportPath:  "github.com/rabbitmq/amqp091-go",
		Description: "AMQP 0-9-1 client for RabbitMQ — reliable queues",
	},

	// ── Observability ────────────────────────────────────────────
	{
		ID: "zap", Name: "Uber Zap", Category: CatLogger,
		ImportPath:  "go.uber.org/zap",
		Description: "Blazing fast, structured, leveled logging",
	},
	{
		ID: "logrus", Name: "Logrus", Category: CatLogger,
		ImportPath:  "github.com/sirupsen/logrus",
		Description: "Structured, pluggable logging",
	},
	{
		ID: "zerolog", Name: "Zerolog", Category: CatLogger,
		ImportPath:  "github.com/rs/zerolog",
		Description: "Zero-allocation JSON logger",
	},
	{
		ID: "otel", Name: "OpenTelemetry", Category: CatTracing,
		ImportPath:  "go.opentelemetry.io/otel",
		Description: "Distributed tracing and metrics instrumentation API",
	},
	{
		ID: "prometheus", Name: "Prometheus client", Category: CatMetrics,
		ImportPath:  "github.com/prometheus/client_golang",
		Description: "Prometheus metrics client for exposing /metrics",
	},

	// ── Security ─────────────────────────────────────────────────
	{
		ID: "jwt", Name: "golang-jwt", Category: CatAuth,
		ImportPath:  "github.com/golang-jwt/jwt/v5",
		Description: "JSON Web Token creation and verification",
	},
	{
		ID: "crypto", Name: "golang.org/x/crypto", Category: CatAuth,
		ImportPath:  "golang.org/x/crypto",
		Description: "Password hashing (bcrypt) and extended cryptography",
	},
	{
		ID: "casbin", Name: "Casbin", Category: CatAuth,
		ImportPath:  "github.com/casbin/casbin/v2",
		Description: "Authorization library — RBAC/ABAC access control",
	},

	// ── Dependency Injection ─────────────────────────────────────
	{
		ID: "dig", Name: "Uber Dig", Category: CatDI,
		ImportPath:  "go.uber.org/dig",
		Description: "Reflection-based dependency injection container",
	},
	{
		ID: "wire", Name: "Google Wire", Category: CatDI,
		ImportPath:  "github.com/google/wire",
		Description: "Compile-time dependency injection code generator",
	},

	// ── Configuration ────────────────────────────────────────────
	{
		ID: "viper", Name: "Viper", Category: CatConfig,
		ImportPath:  "github.com/spf13/viper",
		Description: "Config management across files, env vars, and flags",
	},

	// ── Utilities ────────────────────────────────────────────────
	{
		ID: "validator", Name: "go playground validator", Category: CatValidation,
		ImportPath:  "github.com/go-playground/validator/v10",
		Description: "Struct and field validation via struct tags",
	},
	{
		ID: "swag", Name: "Swaggo", Category: CatDoc,
		ImportPath:  "github.com/swaggo/swag",
		Description: "Generates Swagger 2.0 docs from Go annotations",
	},
	{
		ID: "uuid", Name: "Google UUID", Category: CatUtility,
		ImportPath:  "github.com/google/uuid",
		Description: "UUID generation",
	},
	{
		ID: "decimal", Name: "shopspring decimal", Category: CatUtility,
		ImportPath:  "github.com/shopspring/decimal",
		Description: "Arbitrary-precision decimal — safe money/finance math",
	},
	{
		ID: "run", Name: "oklog/run", Category: CatUtility,
		ImportPath:  "github.com/oklog/run",
		Description: "Actor-group lifecycle — coordinated graceful shutdown",
	},
	{
		ID: "cron", Name: "robfig/cron", Category: CatUtility,
		ImportPath:  "github.com/robfig/cron/v3",
		Description: "Cron-style job scheduler",
	},
	{
		ID: "ratelimit", Name: "golang.org/x/time", Category: CatUtility,
		ImportPath:  "golang.org/x/time",
		Description: "Token-bucket rate limiting",
	},
	{
		ID: "gobreaker", Name: "sony/gobreaker", Category: CatUtility,
		ImportPath:  "github.com/sony/gobreaker",
		Description: "Circuit breaker pattern for fault-tolerant calls",
	},
	{
		ID: "testify", Name: "testify", Category: CatUtility,
		ImportPath:  "github.com/stretchr/testify",
		Description: "Testing toolkit — assertions, mocks, suites",
	},
}

// depGroups defines the display order and category membership for each group.
var depGroups = []struct {
	label      string
	categories []string
}{
	{"Web / Routing", []string{CatFramework, CatRPC}},
	{"Database", []string{CatORM, CatDriver, CatMigration}},
	{"Cache", []string{CatCache}},
	{"Messaging", []string{CatMessageBroker}},
	{"Observability", []string{CatLogger, CatTracing, CatMetrics}},
	{"Security", []string{CatAuth}},
	{"Dependency Injection", []string{CatDI}},
	{"Configuration", []string{CatConfig}},
	{"Utilities", []string{CatValidation, CatDoc, CatUtility}},
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
