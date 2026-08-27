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
	CatCLI           = "cli"
	CatQueue         = "queue"
	CatBlockchain    = "blockchain"
	CatCloud         = "cloud"
	CatTesting       = "testing"
	CatGUI           = "gui"
	CatPayment       = "payment"
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
	case CatCLI:
		return base.Background(lipgloss.Color("#455A64"))
	case CatQueue:
		return base.Background(lipgloss.Color("#8D6E63"))
	case CatBlockchain:
		return base.Background(lipgloss.Color("#627EEA"))
	case CatCloud:
		return base.Background(lipgloss.Color("#FF9900"))
	case CatTesting:
		return base.Background(lipgloss.Color("#EF6C00"))
	case CatGUI:
		return base.Background(lipgloss.Color("#5C6BC0"))
	case CatPayment:
		return base.Background(lipgloss.Color("#00C853"))
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

// docURL points at the package's pkg.go.dev page — it renders the module's
// README plus its API docs, so it's a correct reference for any Go module
// without having to hand-curate a doc link per registry entry.
func docURL(importPath string) string {
	return "https://pkg.go.dev/" + importPath
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
	{
		ID: "gqlgen", Name: "gqlgen", Category: CatRPC,
		ImportPath:  "github.com/99designs/gqlgen",
		Description: "Schema-first GraphQL server generator",
	},
	{
		ID: "gorilla-mux", Name: "Gorilla Mux", Category: CatFramework,
		ImportPath:  "github.com/gorilla/mux",
		Description: "Ubiquitous HTTP request router/dispatcher",
	},
	{
		ID: "websocket", Name: "Gorilla WebSocket", Category: CatFramework,
		ImportPath:  "github.com/gorilla/websocket",
		Description: "WebSocket implementation for real-time connections",
	},
	{
		ID: "alice", Name: "Alice", Category: CatFramework,
		ImportPath:  "github.com/justinas/alice",
		Description: "Painless net/http middleware chaining",
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
		ID: "bun", Name: "Bun", Category: CatORM,
		ImportPath:  "github.com/uptrace/bun",
		Description: "SQL-first ORM/query builder for Postgres, MySQL, SQLite",
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
		ID: "gocql", Name: "gocql (Cassandra)", Category: CatDriver,
		ImportPath:  "github.com/gocql/gocql",
		Description: "Cassandra / ScyllaDB client — wide-column NoSQL",
	},
	{
		ID: "elasticsearch", Name: "Elasticsearch", Category: CatDriver,
		ImportPath:  "github.com/elastic/go-elasticsearch/v8",
		Description: "Elasticsearch client — search & analytics store",
	},
	{
		ID: "dynamodb", Name: "AWS DynamoDB", Category: CatDriver,
		ImportPath:  "github.com/aws/aws-sdk-go-v2/service/dynamodb",
		Description: "DynamoDB client — managed NoSQL key-value/document store",
	},
	{
		ID: "etcd", Name: "etcd", Category: CatDriver,
		ImportPath:  "go.etcd.io/etcd/client/v3",
		Description: "Distributed key-value store — config, coordination, locks",
	},
	{
		ID: "badger", Name: "Badger", Category: CatDriver,
		ImportPath:  "github.com/dgraph-io/badger/v4",
		Description: "Embedded key-value NoSQL store, pure Go, no server needed",
	},
	{
		ID: "neo4j", Name: "Neo4j", Category: CatDriver,
		ImportPath:  "github.com/neo4j/neo4j-go-driver/v5",
		Description: "Neo4j graph database driver",
	},
	{
		ID: "bleve", Name: "Bleve", Category: CatDriver,
		ImportPath:  "github.com/blevesearch/bleve/v2",
		Description: "Embedded full-text search engine, pure Go",
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
	{
		ID: "godotenv", Name: "godotenv", Category: CatConfig,
		ImportPath:  "github.com/joho/godotenv",
		Description: "Loads environment variables from a .env file",
	},

	// ── Blockchain ───────────────────────────────────────────────
	{
		ID: "go-ethereum", Name: "go-ethereum", Category: CatBlockchain,
		ImportPath:  "github.com/ethereum/go-ethereum",
		Description: "Ethereum protocol client — signing, ABI, JSON-RPC",
	},
	{
		ID: "solana-go", Name: "solana-go", Category: CatBlockchain,
		ImportPath:  "github.com/gagliardetto/solana-go",
		Description: "Solana blockchain SDK for Go",
	},
	{
		ID: "btcd", Name: "btcd", Category: CatBlockchain,
		ImportPath:  "github.com/btcsuite/btcd",
		Description: "Full Bitcoin protocol implementation in Go",
	},
	{
		ID: "cosmos-sdk", Name: "Cosmos SDK", Category: CatBlockchain,
		ImportPath:  "github.com/cosmos/cosmos-sdk",
		Description: "Framework for building application-specific blockchains",
	},
	{
		ID: "fabric-sdk-go", Name: "Hyperledger Fabric SDK", Category: CatBlockchain,
		ImportPath:  "github.com/hyperledger/fabric-sdk-go",
		Description: "Permissioned enterprise blockchain SDK",
	},
	{
		ID: "cometbft", Name: "CometBFT", Category: CatBlockchain,
		ImportPath:  "github.com/cometbft/cometbft",
		Description: "BFT state machine replication / consensus engine",
	},

	// ── Cloud SDKs ───────────────────────────────────────────────
	{
		ID: "aws-sdk", Name: "AWS SDK v2", Category: CatCloud,
		ImportPath:  "github.com/aws/aws-sdk-go-v2",
		Description: "AWS SDK core — pair with per-service submodules",
	},
	{
		ID: "gcp-sdk", Name: "Google Cloud SDK", Category: CatCloud,
		ImportPath:  "cloud.google.com/go",
		Description: "Google Cloud client libraries root module",
	},
	{
		ID: "azure-sdk", Name: "Azure SDK", Category: CatCloud,
		ImportPath:  "github.com/Azure/azure-sdk-for-go",
		Description: "Azure SDK for Go",
	},
	{
		ID: "minio", Name: "MinIO client", Category: CatCloud,
		ImportPath:  "github.com/minio/minio-go/v7",
		Description: "S3-compatible object storage client",
	},
	{
		ID: "k8s-client", Name: "client-go", Category: CatCloud,
		ImportPath:  "k8s.io/client-go",
		Description: "Official Kubernetes API client",
	},
	{
		ID: "consul", Name: "Consul API", Category: CatCloud,
		ImportPath:  "github.com/hashicorp/consul/api",
		Description: "Service discovery and distributed config client",
	},
	{
		ID: "vault", Name: "Vault API", Category: CatCloud,
		ImportPath:  "github.com/hashicorp/vault/api",
		Description: "Secrets management client",
	},
	{
		ID: "docker-client", Name: "Docker client", Category: CatCloud,
		ImportPath:  "github.com/docker/docker/client",
		Description: "Docker Engine API client",
	},

	// ── CLI Tools ────────────────────────────────────────────────
	{
		ID: "cobra", Name: "Cobra", Category: CatCLI,
		ImportPath:  "github.com/spf13/cobra",
		Description: "Modern CLI framework — commands, flags, subcommands",
	},
	{
		ID: "urfave-cli", Name: "urfave/cli", Category: CatCLI,
		ImportPath:  "github.com/urfave/cli/v3",
		Description: "Simple, fast package for building CLI apps",
	},

	// ── Background Jobs ──────────────────────────────────────────
	{
		ID: "asynq", Name: "Asynq", Category: CatQueue,
		ImportPath:  "github.com/hibiken/asynq",
		Description: "Redis-backed distributed task queue",
	},
	{
		ID: "river", Name: "River", Category: CatQueue,
		ImportPath:  "github.com/riverqueue/river",
		Description: "Postgres-backed background job queue",
	},

	// ── Payments ─────────────────────────────────────────────────
	{
		ID: "stripe-go", Name: "Stripe", Category: CatPayment,
		ImportPath:  "github.com/stripe/stripe-go/v81",
		Description: "Stripe payments API client",
	},

	// ── Testing ──────────────────────────────────────────────────
	{
		ID: "testify", Name: "testify", Category: CatTesting,
		ImportPath:  "github.com/stretchr/testify",
		Description: "Testing toolkit — assertions, mocks, suites",
	},
	{
		ID: "gomock", Name: "go.uber.org/mock", Category: CatTesting,
		ImportPath:  "go.uber.org/mock",
		Description: "Mocking framework for Go interfaces",
	},
	{
		ID: "ginkgo", Name: "Ginkgo", Category: CatTesting,
		ImportPath:  "github.com/onsi/ginkgo/v2",
		Description: "BDD-style testing framework",
	},
	{
		ID: "gomega", Name: "Gomega", Category: CatTesting,
		ImportPath:  "github.com/onsi/gomega",
		Description: "Matcher/assertion library, pairs with Ginkgo",
	},
	{
		ID: "gofakeit", Name: "gofakeit", Category: CatTesting,
		ImportPath:  "github.com/brianvoe/gofakeit/v7",
		Description: "Fake/random test data generator",
	},

	// ── Desktop & Game Dev ───────────────────────────────────────
	{
		ID: "fyne", Name: "Fyne", Category: CatGUI,
		ImportPath:  "fyne.io/fyne/v2",
		Description: "Cross-platform desktop/mobile GUI toolkit",
	},
	{
		ID: "ebiten", Name: "Ebiten", Category: CatGUI,
		ImportPath:  "github.com/hajimehoshi/ebiten/v2",
		Description: "Simple, fast 2D game engine",
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
		ID: "resty", Name: "Resty", Category: CatUtility,
		ImportPath:  "github.com/go-resty/resty/v2",
		Description: "Simple, chainable HTTP client",
	},
	{
		ID: "go-money", Name: "go-money", Category: CatUtility,
		ImportPath:  "github.com/Rhymond/go-money",
		Description: "Money value type — safe currency arithmetic for finance",
	},
	{
		ID: "excelize", Name: "Excelize", Category: CatUtility,
		ImportPath:  "github.com/xuri/excelize/v2",
		Description: "Read/write Excel .xlsx files — reports, exports",
	},
	{
		ID: "copier", Name: "jinzhu/copier", Category: CatUtility,
		ImportPath:  "github.com/jinzhu/copier",
		Description: "Struct-to-struct field copying",
	},
	{
		ID: "msgpack", Name: "msgpack", Category: CatUtility,
		ImportPath:  "github.com/vmihailenco/msgpack/v5",
		Description: "Compact binary serialization — faster/smaller than JSON",
	},
	{
		ID: "x-sync", Name: "golang.org/x/sync", Category: CatUtility,
		ImportPath:  "golang.org/x/sync",
		Description: "errgroup, singleflight, semaphore — concurrency helpers",
	},
	{
		ID: "mapstructure", Name: "mapstructure", Category: CatUtility,
		ImportPath:  "github.com/mitchellh/mapstructure",
		Description: "Decode generic maps into structs",
	},
	{
		ID: "sprig", Name: "Sprig", Category: CatUtility,
		ImportPath:  "github.com/Masterminds/sprig/v3",
		Description: "100+ helper functions for text/template and html/template",
	},
	{
		ID: "gocsv", Name: "gocsv", Category: CatUtility,
		ImportPath:  "github.com/gocarina/gocsv",
		Description: "Struct-based CSV marshal/unmarshal",
	},
	{
		ID: "go-mail", Name: "go-mail", Category: CatUtility,
		ImportPath:  "github.com/wneessen/go-mail",
		Description: "Modern SMTP email client",
	},
	{
		ID: "gofpdf", Name: "gofpdf", Category: CatUtility,
		ImportPath:  "github.com/jung-kurt/gofpdf",
		Description: "PDF generation — invoices, reports",
	},
}

// depGroup is one category section in the dependency picker.
type depGroup struct {
	label      string
	categories []string
}

// depGroups defines the display order and category membership for each group.
var depGroups = []depGroup{
	{"Web / Routing", []string{CatFramework, CatRPC}},
	{"Database", []string{CatORM, CatDriver, CatMigration}},
	{"Cache", []string{CatCache}},
	{"Messaging", []string{CatMessageBroker}},
	{"Observability", []string{CatLogger, CatTracing, CatMetrics}},
	{"Security", []string{CatAuth}},
	{"Dependency Injection", []string{CatDI}},
	{"Configuration", []string{CatConfig}},
	{"CLI Tools", []string{CatCLI}},
	{"Background Jobs", []string{CatQueue}},
	{"Blockchain", []string{CatBlockchain}},
	{"Cloud SDKs", []string{CatCloud}},
	{"Payments", []string{CatPayment}},
	{"Testing", []string{CatTesting}},
	{"Desktop & Game Dev", []string{CatGUI}},
	{"Utilities", []string{CatValidation, CatDoc, CatUtility}},
}

// groupHeaderStyle is the amber label rendered above each category section.
var groupHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#E3B341")).
	Bold(true)

// depRow is one navigable line in the dependency picker: either a group
// header or, when that header's group is expanded, one of its dependencies.
type depRow struct {
	header   bool
	groupIdx int
	regIdx   int // valid only when !header
}

// visibleRows returns the current navigable rows. Only one group is expanded
// at a time (accordion) so the list never floods the screen — collapsing a
// 30+ package registry down to ~9 header lines by default. Searching expands
// every group that has a match, since results need to stay comparable across
// categories.
func (m *Model) visibleRows() []depRow {
	query := strings.ToLower(strings.TrimSpace(m.SearchQuery))
	searching := query != ""

	var rows []depRow
	for gi, group := range depGroups {
		var matched []int
		for i, dep := range m.Registry {
			for _, cat := range group.categories {
				if dep.Category != cat {
					continue
				}
				if !searching ||
					strings.Contains(strings.ToLower(dep.Name), query) ||
					strings.Contains(strings.ToLower(dep.Category), query) ||
					strings.Contains(strings.ToLower(dep.Description), query) {
					matched = append(matched, i)
				}
				break
			}
		}
		if len(matched) == 0 {
			continue
		}
		rows = append(rows, depRow{header: true, groupIdx: gi})
		if searching || m.ExpandedGroup == gi {
			for _, i := range matched {
				rows = append(rows, depRow{groupIdx: gi, regIdx: i})
			}
		}
	}
	return rows
}

// groupSelectedCount returns how many of a group's dependencies are chosen,
// shown on its (possibly collapsed) header so a selection is never hidden.
func (m *Model) groupSelectedCount(group depGroup) int {
	n := 0
	for i, dep := range m.Registry {
		for _, cat := range group.categories {
			if dep.Category != cat {
				continue
			}
			if _, ok := m.Chosen[i]; ok {
				n++
			}
			break
		}
	}
	return n
}

// activateCursorRow acts on whatever row the cursor is on: expand/collapse a
// group header, or toggle a dependency's selection.
func (m *Model) activateCursorRow() {
	rows := m.visibleRows()
	if m.Cursor < 0 || m.Cursor >= len(rows) {
		return
	}
	row := rows[m.Cursor]
	if !row.header {
		m.toggleDependency(row.regIdx)
		return
	}

	if m.ExpandedGroup == row.groupIdx {
		m.ExpandedGroup = -1
	} else {
		m.ExpandedGroup = row.groupIdx
	}

	// Expanding/collapsing a group shifts row indices under it (and, via the
	// accordion closing whatever was open, possibly rows before it too) —
	// re-locate the cursor onto the header the user just acted on.
	for i, r := range m.visibleRows() {
		if r.header && r.groupIdx == row.groupIdx {
			m.Cursor = i
			break
		}
	}
}

// rowLines returns how many terminal lines a row renders as.
func (m *Model) rowLines(row depRow) int {
	if row.header {
		return 1
	}
	return 2
}

// rowBudget estimates how many content lines are available for the row list
// after header, step nav, divider, panel chrome, and footer hints. Unknown
// height (before the first WindowSizeMsg) means "don't clip yet".
func (m *Model) rowBudget() int {
	if m.Height <= 0 {
		return 1 << 30
	}
	// Reserve room for header + mode line + step nav + divider + the panel's
	// own label/hint/search/footer chrome, plus the "N more above/below"
	// lines a clipped list adds — better to under-fill by a line than let
	// the terminal itself scroll the top of the frame off screen.
	b := m.Height - 22
	if b < 4 {
		b = 4
	}
	return b
}

// rowWindow returns the [start,end) slice of rows that fit the current
// height budget with the cursor guaranteed visible, plus how many rows are
// hidden above/below — this is what stops the cursor from running off the
// bottom of a short terminal with nothing telling the user more exists.
func (m *Model) rowWindow(rows []depRow) (start, end, hiddenAbove, hiddenBelow int) {
	budget := m.rowBudget()

	total := 0
	for _, r := range rows {
		total += m.rowLines(r)
	}
	if total <= budget {
		return 0, len(rows), 0, 0
	}

	start, end = m.Cursor, m.Cursor+1
	used := m.rowLines(rows[m.Cursor])
	for used < budget && (start > 0 || end < len(rows)) {
		if end < len(rows) {
			used += m.rowLines(rows[end])
			end++
			if used >= budget {
				break
			}
		}
		if start > 0 {
			start--
			used += m.rowLines(rows[start])
		}
	}
	return start, end, start, len(rows) - end
}

// renderDependencyView renders the StepDeps panel: collapsed group headers
// with a selection count, one group expandable at a time, scrolled to keep
// the cursor on screen.
func (m *Model) renderDependencyView() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("DEPENDENCIES") + "\n")
	b.WriteString(styles.PanelHint.Render("space opens a category, space again toggles a package, enter to review") + "\n\n")

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

	rows := m.visibleRows()

	if len(rows) == 0 {
		b.WriteString(styles.Description.Render("  no results for \""+m.SearchQuery+"\"") + "\n\n")
	} else {
		start, end, hiddenAbove, hiddenBelow := m.rowWindow(rows)

		if hiddenAbove > 0 {
			b.WriteString(styles.Description.Render(fmt.Sprintf("  ⋯ %d more above", hiddenAbove)) + "\n")
		}

		for i := start; i < end; i++ {
			row := rows[i]
			isActive := m.Cursor == i

			if row.header {
				group := depGroups[row.groupIdx]
				arrow := "▸"
				if m.ExpandedGroup == row.groupIdx || m.SearchQuery != "" {
					arrow = "▾"
				}
				label := arrow + " " + group.label
				if n := m.groupSelectedCount(group); n > 0 {
					label += fmt.Sprintf("  (%d selected)", n)
				}

				cursor := "  "
				style := groupHeaderStyle
				if isActive {
					cursor = styles.Cursor.Render("▶ ")
					style = style.Underline(true)
				}
				b.WriteString(cursor + style.Render(label) + "\n")
				continue
			}

			dep := m.Registry[row.regIdx]
			cursor := "     "
			if isActive {
				cursor = styles.Cursor.Render("   ▶ ")
			}

			_, chosen := m.Chosen[row.regIdx]
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
			b.WriteString(fmt.Sprintf("        %s\n", styles.Description.Render(dep.Description)))
		}

		if hiddenBelow > 0 {
			b.WriteString(styles.Description.Render(fmt.Sprintf("  ⋯ %d more below", hiddenBelow)) + "\n")
		}
		b.WriteRune('\n')
	}

	hints := []keyHint{
		{"↑↓ / jk", "navigate"},
		{"space", "open / toggle"},
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
