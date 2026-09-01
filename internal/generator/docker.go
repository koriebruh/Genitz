package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// composeService describes a docker-compose service for a registry
// dependency that has a real backing container image. name is the compose
// service key — kafka-go and sarama both point at the same "kafka" service
// so picking either (or both) never produces duplicate blocks.
type composeService struct {
	name        string
	image       string
	ports       []string
	environment []string
}

// dockerComposeServices maps a Dependency.ID (see internal/tui/dependencies.go)
// to the compose service it needs. Only infra-backed dependencies belong
// here — embedded/in-process libraries (badger, bleve, ristretto, bigcache,
// go-cache, groupcache), cloud-hosted-by-nature ones (dynamodb), and
// anything too opinionated to auto-spin-up (vault, consul, k8s-client,
// docker-client) are deliberately left out. Add an entry here when a new
// registry dependency has an obvious, safe-to-default local container image.
var dockerComposeServices = map[string]composeService{
	"redis": {
		name:  "redis",
		image: "redis:7-alpine",
		ports: []string{"6379:6379"},
	},
	"mysql-driver": {
		name:        "mysql",
		image:       "mysql:8",
		ports:       []string{"3306:3306"},
		environment: []string{"MYSQL_ROOT_PASSWORD=root", "MYSQL_DATABASE=app"},
	},
	"postgres-driver": {
		name:        "postgres",
		image:       "postgres:16-alpine",
		ports:       []string{"5432:5432"},
		environment: []string{"POSTGRES_PASSWORD=postgres", "POSTGRES_DB=app"},
	},
	"mongo-driver": {
		name:  "mongo",
		image: "mongo:7",
		ports: []string{"27017:27017"},
	},
	"rabbitmq": {
		name:  "rabbitmq",
		image: "rabbitmq:3-management-alpine",
		ports: []string{"5672:5672", "15672:15672"},
	},
	"nats": {
		name:  "nats",
		image: "nats:2-alpine",
		ports: []string{"4222:4222"},
	},
	"kafka-go": {
		name:  "kafka",
		image: "bitnami/kafka:latest",
		ports: []string{"9092:9092"},
		environment: []string{
			"KAFKA_CFG_NODE_ID=0",
			"KAFKA_CFG_PROCESS_ROLES=controller,broker",
			"KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka:9093",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER",
		},
	},
	"sarama": {
		name:  "kafka",
		image: "bitnami/kafka:latest",
		ports: []string{"9092:9092"},
		environment: []string{
			"KAFKA_CFG_NODE_ID=0",
			"KAFKA_CFG_PROCESS_ROLES=controller,broker",
			"KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka:9093",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER",
		},
	},
	"elasticsearch": {
		name:        "elasticsearch",
		image:       "docker.elastic.co/elasticsearch/elasticsearch:8.15.0",
		ports:       []string{"9200:9200"},
		environment: []string{"discovery.type=single-node", "xpack.security.enabled=false"},
	},
	"gocql": {
		name:  "cassandra",
		image: "cassandra:5",
		ports: []string{"9042:9042"},
	},
	"etcd": {
		name:        "etcd",
		image:       "bitnami/etcd:latest",
		ports:       []string{"2379:2379"},
		environment: []string{"ALLOW_NONE_AUTHENTICATION=yes"},
	},
	"neo4j": {
		name:        "neo4j",
		image:       "neo4j:5",
		ports:       []string{"7474:7474", "7687:7687"},
		environment: []string{"NEO4J_AUTH=none"},
	},
}

// ComposeServiceNames returns the compose service names (deduped, sorted)
// that the given dependency IDs would produce — main.go injects this into
// the TUI Model for the Docker step's live "docker-compose will include:
// ..." preview, so the tui package doesn't need its own copy of the
// dockerComposeServices table (and can't import generator to reuse this
// one — the dependency runs the other way).
func ComposeServiceNames(depIDs []string) []string {
	seen := make(map[string]bool)
	var names []string
	for _, id := range depIDs {
		svc, ok := dockerComposeServices[id]
		if !ok || seen[svc.name] {
			continue
		}
		seen[svc.name] = true
		names = append(names, svc.name)
	}
	sort.Strings(names)
	return names
}

// composeContent builds docker-compose.yml for req: an "app" service plus
// one block per distinct backing service its selected dependencies need.
// ok is false when nothing selected maps to a service — a compose file
// with only "app" adds nothing over the Dockerfile alone.
func composeContent(req Requirement) (content string, ok bool) {
	ids := make([]string, 0, len(req.Deps))
	for _, dep := range req.Deps {
		ids = append(ids, dep.ID)
	}
	names := ComposeServiceNames(ids)
	if len(names) == 0 {
		return "", false
	}

	services := make(map[string]composeService, len(names))
	for _, dep := range req.Deps {
		if svc, found := dockerComposeServices[dep.ID]; found {
			services[svc.name] = svc
		}
	}

	var b strings.Builder
	b.WriteString("services:\n")
	b.WriteString("  app:\n")
	b.WriteString("    build:\n")
	b.WriteString("      context: .\n")
	b.WriteString("      dockerfile: Dockerfile\n")
	b.WriteString("    ports:\n")
	b.WriteString("      - \"8080:8080\"\n")
	b.WriteString("    depends_on:\n")
	for _, name := range names {
		b.WriteString("      - " + name + "\n")
	}

	for _, name := range names {
		svc := services[name]
		b.WriteString("\n  " + name + ":\n")
		b.WriteString("    image: " + svc.image + "\n")
		if len(svc.ports) > 0 {
			b.WriteString("    ports:\n")
			for _, p := range svc.ports {
				b.WriteString("      - \"" + p + "\"\n")
			}
		}
		if len(svc.environment) > 0 {
			b.WriteString("    environment:\n")
			for _, e := range svc.environment {
				b.WriteString("      - " + e + "\n")
			}
		}
	}

	return b.String(), true
}

// dockerfileContent returns a multistage Dockerfile pinned to goVersion
// (major.minor, e.g. "1.25").
func dockerfileContent(goVersion string) string {
	return fmt.Sprintf(`# syntax=docker/dockerfile:1
FROM golang:%s-alpine AS builder
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/app .

# Placeholder — Genitz only scaffolds a bare main.go with no server code,
# so there's no real port to detect here. Adjust to match what your app
# actually listens on.
EXPOSE 8080

CMD ["./app"]
`, goVersion)
}

func dockerignoreContent() string {
	return `.git
.gitignore
*.md
Dockerfile
docker-compose.yml
.dockerignore
bin/
tmp/
`
}

// readGoVersion reads the "go X.Y[.Z]" directive from go.mod in dir and
// trims it to major.minor (Docker Hub's golang image tags don't carry a
// patch version). Falls back to a recent stable version if go.mod can't be
// read or parsed.
func readGoVersion(dir string) string {
	const fallback = "1.25"

	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return fallback
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "go ") {
			continue
		}
		parts := strings.SplitN(strings.TrimSpace(strings.TrimPrefix(line, "go")), ".", 3)
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1]
		}
	}
	return fallback
}
