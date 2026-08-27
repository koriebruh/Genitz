# Genitz

**A terminal CLI, styled like Claude Code, for starting Go projects and
adding dependencies without leaving the keyboard.**

[![Go Reference](https://pkg.go.dev/badge/github.com/koriebruh/Genitz.svg)](https://pkg.go.dev/github.com/koriebruh/Genitz)
[![Go Report Card](https://goreportcard.com/badge/github.com/koriebruh/Genitz)](https://goreportcard.com/report/github.com/koriebruh/Genitz)
[![License: Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Genitz is a single binary with two jobs: scaffold a brand-new Go project, or
pick and install dependencies into the project you're already standing in —
both through the same fast, keyboard-driven picker instead of hand-typing
`go get <import path>` from memory or hunting for package names online.

## Demo

![Genitz demo — scaffolding a new project and picking dependencies](docs/demo.gif)

*Recorded straight from the terminal with [VHS](https://github.com/charmbracelet/vhs)
— nothing staged. `genitz` → name the project → pick Fiber and Echo from the
picker → review → confirm → real `go get` runs → doc links printed.*

## Features

- **Two flows, one binary.** No `go.mod` in the current directory → new
  project wizard. `go.mod` already there → straight to the dependency
  picker. `genitz init` / `genitz add` force either flow explicitly.
- **85 dependencies across 16 categories** — web frameworks, ORMs, raw SQL
  and NoSQL/search drivers, message brokers, observability, auth, dependency
  injection, config loaders, CLI-building libraries, background job queues,
  blockchain SDKs, cloud/infra SDKs, payments, testing, desktop & game dev,
  and general utilities. Whatever corner of Go you work in, it should be in
  here. See the [full list](#dependency-registry) below.
- **Collapsible picker, not a wall of text.** Categories collapse to one
  line by default; open one at a time, with a live "(N selected)" count so a
  choice is never hidden. Type `/` to search across everything at once.
- **Responsive TUI.** Renders cleanly whether your terminal is a small split
  pane or a maximized window — no ASCII-art logo shoving your content off
  screen, no cursor scrolling out of view.
- **Doc links, not guesswork.** Every dependency on the Review screen — and
  in the terminal output after install — comes with its
  [pkg.go.dev](https://pkg.go.dev) reference link.
- **No template lock-in.** Genitz writes a bare `main.go` and lets `go get`
  do the rest — it doesn't force an opinionated project layout on you.

## Installation

Requires Go 1.25+.

```sh
go install github.com/koriebruh/Genitz@latest
```

This puts a `genitz` (or `Genitz`, depending on your `go install` naming)
binary in `$(go env GOPATH)/bin` — make sure that's on your `PATH`.

## Usage

```text
genitz         Auto-detect: start a new project, or add dependencies if
               go.mod already exists in the current directory.
genitz init    Always start the new-project wizard.
genitz add     Add dependencies to the project in the current directory
               (requires an existing go.mod).
genitz help    Show usage.
```

### Start a new project

```sh
mkdir my-app && cd my-app   # optional — genitz creates the folder for you
genitz init
```

Walks you through: folder name → module path → dependency picker → review →
scaffold. You end up with a `go.mod`, a minimal `main.go`, and whatever
packages you picked already `go get`-installed.

### Add dependencies to an existing project

```sh
cd my-existing-project   # anywhere with a go.mod
genitz add
```

Same picker, straight to it — no folder/module questions since the project
already exists. Runs `go get` + `go mod tidy` when you confirm.

### Picker keys

| Key       | Action                                             |
|-----------|-----------------------------------------------------|
| `↑`/`↓`, `j`/`k` | Move the cursor                              |
| `space`   | On a category: expand/collapse it. On a package: toggle it |
| `/`       | Search by name, category, or description            |
| `enter`   | Go to the Review step                                |
| `b`       | Back to the picker (from Review)                     |
| `enter` / `y` | Confirm and generate (from Review)               |
| `q` / `ctrl+c` | Quit                                            |

## Dependency registry

<details>
<summary>All 85 packages, grouped (click to expand)</summary>

| Group | Packages |
|---|---|
| Web / Routing | Fiber, Gin Gonic, Echo, Chi, gRPC, Protocol Buffers, gqlgen, Gorilla Mux, Gorilla WebSocket, Alice |
| Database | GORM, sqlx, Bun, MySQL Driver, pgx (PostgreSQL), MongoDB Driver, gocql (Cassandra), Elasticsearch, AWS DynamoDB, etcd, Badger, Neo4j, Bleve, golang-migrate |
| Cache | redis |
| Messaging | kafka-go, NATS, RabbitMQ (amqp091) |
| Observability | Uber Zap, Logrus, Zerolog, OpenTelemetry, Prometheus client |
| Security | golang-jwt, golang.org/x/crypto, Casbin |
| Dependency Injection | Uber Dig, Google Wire |
| Configuration | Viper, godotenv |
| CLI Tools | Cobra, urfave/cli |
| Background Jobs | Asynq, River |
| Blockchain | go-ethereum, solana-go, btcd, Cosmos SDK, Hyperledger Fabric SDK, CometBFT |
| Cloud & Infra SDKs | AWS SDK v2, Google Cloud SDK, Azure SDK, MinIO client, client-go (Kubernetes), Consul API, Vault API, Docker client |
| Payments | Stripe |
| Testing | testify, go.uber.org/mock, Ginkgo, Gomega, gofakeit |
| Desktop & Game Dev | Fyne, Ebiten |
| Utilities | go playground validator, Swaggo, Google UUID, shopspring decimal, oklog/run, robfig/cron, golang.org/x/time, sony/gobreaker, Resty, go-money, Excelize, jinzhu/copier, msgpack, golang.org/x/sync, mapstructure, Sprig, gocsv, go-mail, gofpdf |

</details>

Every package shows a short description and a live pkg.go.dev doc link in
the picker and on the Review screen — pick what you need, read the one-liner,
follow the link if you want the full docs, no separate search required.

Missing something you reach for often? Open a PR — see
[`internal/tui/dependencies.go`](internal/tui/dependencies.go) for the shape
of a registry entry.

## How it works

Genitz doesn't scaffold an opinionated architecture (that feature existed
early on and was intentionally removed — see the git history). What it does
today is deliberately small:

- **New project**: create the folder, write a one-line `main.go`, `go mod
  init <module>`, then `go get` each selected package.
- **Add dependencies**: `go get` each selected package into the project
  already in the current directory, then `go mod tidy`.

That's the whole contract — Genitz gets your `go.mod` and imports in order
and gets out of the way.

## Development

```sh
go build ./...
go vet ./...
go test ./...
gofmt -l .   # should print nothing
```

The TUI has no headless test harness — verify layout changes across
terminal sizes with `tmux`:

```sh
go build -o /tmp/genitz .
tmux new-session -d -s t -x 40 -y 24 "cd /some/test/dir && /tmp/genitz"
tmux capture-pane -t t -p
```

See [`CLAUDE.md`](CLAUDE.md) for a deeper tour of the codebase layout and
the accordion/scroll-window logic behind the picker.

## License

[Apache License 2.0](LICENSE).
