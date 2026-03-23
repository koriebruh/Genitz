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
	CatConfig        = "config"
	CatMigration     = "migration"
	CatWorker        = "worker"
	CatObservability = "observability"
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
		ID: "gorm-postgres", Name: "GORM Postgres", Category: CatORM,
		ImportPath:  "gorm.io/driver/postgres",
		Description: "Official PostgreSQL driver for GORM",
		TemplateDir: fmt.Sprintf(depsPath, "gorm-postgres"),
	},
	{
		ID: "gorm-mysql", Name: "GORM MySQL", Category: CatORM,
		ImportPath:  "gorm.io/driver/mysql",
		Description: "Official MySQL driver for GORM",
		TemplateDir: fmt.Sprintf(depsPath, "gorm-mysql"),
	},
	{
		ID: "gorm-sqlite", Name: "GORM SQLite", Category: CatORM,
		ImportPath:  "gorm.io/driver/sqlite",
		Description: "Official SQLite driver for GORM",
		TemplateDir: fmt.Sprintf(depsPath, "gorm-sqlite"),
	},
	{
		ID: "gorm-sqlserver", Name: "GORM SQLServer", Category: CatORM,
		ImportPath:  "gorm.io/driver/sqlserver",
		Description: "Official SQL Server driver for GORM",
		TemplateDir: fmt.Sprintf(depsPath, "gorm-sqlserver"),
	},
	{
		ID: "zap", Name: "Uber Zap", Category: CatLogger,
		ImportPath:  "go.uber.org/zap",
		Description: "Blazing fast, structured, leveled logging",
		TemplateDir: fmt.Sprintf(depsPath, "zap"),
	},
	// --- NATIVE DATABASE DRIVERS ---
	{
		ID: "pgx", Name: "PGX (PostgreSQL)", Category: CatDriver,
		ImportPath:  "github.com/jackc/pgx/v5",
		Description: "PostgreSQL driver and toolkit for Go",
		TemplateDir: fmt.Sprintf(depsPath, "pgx"),
	},
	{
		ID: "mysql", Name: "Go SQL Driver (MySQL)", Category: CatDriver,
		ImportPath:  "github.com/go-sql-driver/mysql",
		Description: "A MySQL-Driver for Go's database/sql package",
		TemplateDir: fmt.Sprintf(depsPath, "mysql"),
	},
	{
		ID: "mssqldb", Name: "Microsoft SQL Server", Category: CatDriver,
		ImportPath:  "github.com/microsoft/go-mssqldb",
		Description: "Microsoft SQL server driver for Go",
		TemplateDir: fmt.Sprintf(depsPath, "mssqldb"),
	},
	{
		ID: "clickhouse", Name: "ClickHouse Go", Category: CatDriver,
		ImportPath:  "github.com/ClickHouse/clickhouse-go/v2",
		Description: "Golang SQL database driver for ClickHouse",
		TemplateDir: fmt.Sprintf(depsPath, "clickhouse"),
	},

	// ─── Web Frameworks (additional) ────────────────────────────────────────────
	{
		ID: "echo", Name: "Echo", Category: CatFramework,
		ImportPath:  "github.com/labstack/echo/v4",
		Description: "High performance, minimalist Go web framework",
		TemplateDir: fmt.Sprintf(depsPath, "echo"),
	},
	{
		ID: "chi", Name: "Chi Router", Category: CatFramework,
		ImportPath:  "github.com/go-chi/chi/v5",
		Description: "Lightweight, idiomatic and composable router for Go HTTP services",
		TemplateDir: fmt.Sprintf(depsPath, "chi"),
	},

	// ─── Auth & Security ─────────────────────────────────────────────────────
	{
		ID: "jwt", Name: "Golang JWT", Category: CatAuth,
		ImportPath:  "github.com/golang-jwt/jwt/v5",
		Description: "Community maintained Go implementation of JSON Web Tokens",
		TemplateDir: fmt.Sprintf(depsPath, "jwt"),
	},
	{
		ID: "casbin", Name: "Casbin", Category: CatAuth,
		ImportPath:  "github.com/casbin/casbin/v2",
		Description: "Authorization library supporting ACL, RBAC, ABAC",
		TemplateDir: fmt.Sprintf(depsPath, "casbin"),
	},

	// ─── Configuration ───────────────────────────────────────────────────────
	{
		ID: "viper", Name: "Viper", Category: CatConfig,
		ImportPath:  "github.com/spf13/viper",
		Description: "Complete configuration solution for Go apps (env, file, remote)",
		TemplateDir: fmt.Sprintf(depsPath, "viper"),
	},

	// ─── DB Migration ────────────────────────────────────────────────────────
	{
		ID: "goose", Name: "Goose (Migration)", Category: CatMigration,
		ImportPath:  "github.com/pressly/goose/v3",
		Description: "Database migration tool supporting Go, SQL, and OS migrations",
		TemplateDir: fmt.Sprintf(depsPath, "goose"),
	},
	{
		ID: "migrate", Name: "Golang-Migrate", Category: CatMigration,
		ImportPath:  "github.com/golang-migrate/migrate/v4",
		Description: "Database migrations written in Go (CLI & library)",
		TemplateDir: fmt.Sprintf(depsPath, "migrate"),
	},

	// ─── Logging ─────────────────────────────────────────────────────────────
	{
		ID: "logrus", Name: "Logrus", Category: CatLogger,
		ImportPath:  "github.com/sirupsen/logrus",
		Description: "Structured, pluggable logging for Go",
		TemplateDir: fmt.Sprintf(depsPath, "logrus"),
	},
	{
		ID: "zerolog", Name: "Zerolog", Category: CatLogger,
		ImportPath:  "github.com/rs/zerolog",
		Description: "Zero allocation JSON logger for Go",
		TemplateDir: fmt.Sprintf(depsPath, "zerolog"),
	},

	// ─── Observability ───────────────────────────────────────────────────────
	{
		ID: "prometheus", Name: "Prometheus Client", Category: CatObservability,
		ImportPath:  "github.com/prometheus/client_golang/prometheus",
		Description: "Prometheus instrumentation library for Go",
		TemplateDir: fmt.Sprintf(depsPath, "prometheus"),
	},
	{
		ID: "otel", Name: "OpenTelemetry", Category: CatObservability,
		ImportPath:  "go.opentelemetry.io/otel",
		Description: "OpenTelemetry SDK for distributed tracing & metrics",
		TemplateDir: fmt.Sprintf(depsPath, "otel"),
	},
	{
		ID: "sentry", Name: "Sentry Go", Category: CatObservability,
		ImportPath:  "github.com/getsentry/sentry-go",
		Description: "Official Sentry SDK for Go (error monitoring)",
		TemplateDir: fmt.Sprintf(depsPath, "sentry"),
	},

	// ─── Task Queues & Background Jobs ───────────────────────────────────────
	{
		ID: "asynq", Name: "Asynq (Task Queue)", Category: CatWorker,
		ImportPath:  "github.com/hibiken/asynq",
		Description: "Simple, reliable, efficient distributed task queue in Go (backed by Redis)",
		TemplateDir: fmt.Sprintf(depsPath, "asynq"),
	},
	{
		ID: "cron", Name: "Cron Job", Category: CatWorker,
		ImportPath:  "github.com/robfig/cron/v3",
		Description: "A cron library for Go",
		TemplateDir: fmt.Sprintf(depsPath, "cron"),
	},

	// ─── Message Brokers ─────────────────────────────────────────────────────
	{
		ID: "kafka", Name: "Kafka (Confluent)", Category: CatMessageBroker,
		ImportPath:  "github.com/confluentinc/confluent-kafka-go/kafka",
		Description: "Confluent's Apache Kafka client for Go",
		TemplateDir: fmt.Sprintf(depsPath, "kafka"),
	},
	{
		ID: "rabbitmq", Name: "RabbitMQ (AMQP)", Category: CatMessageBroker,
		ImportPath:  "github.com/rabbitmq/amqp091-go",
		Description: "An AMQP 0-9-1 client for Go",
		TemplateDir: fmt.Sprintf(depsPath, "rabbitmq"),
	},
}

// DepGroups defines the display order and category membership for each group.
var DepGroups = []struct {
	Label      string
	Categories []string
}{
	{"Web / Routing", []string{CatFramework, CatRPC}},
	{"Database / ORM", []string{CatORM}},
	{"Database / Drivers", []string{CatDriver}},
	{"DB Migration", []string{CatMigration}},
	{"Cache", []string{CatCache}},
	{"Messaging", []string{CatMessageBroker}},
	{"Auth & Security", []string{CatAuth}},
	{"Configuration", []string{CatConfig}},
	{"Logging", []string{CatLogger}},
	{"Observability", []string{CatObservability, CatTracing, CatMetrics}},
	{"Background Jobs", []string{CatWorker}},
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
