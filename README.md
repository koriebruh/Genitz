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
(script: [`docs/demo.tape`](docs/demo.tape)) — nothing staged. `genitz init` →
name the project → module path → apply the "Web API" preset (Fiber + GORM +
Viper + Zap) from the dependency picker → Docker toggle → Extras (CI, README,
community files, MIT license) → review → confirm → the animated Installing
step (with a small animated ASCII cat) runs a real `go get` → doc links
printed → a few of the newer standalone commands (`search`, `doctor`,
`preset list`).*

## Features

- **Two flows, one binary.** No `go.mod` in the current directory → new
  project wizard. `go.mod` already there → straight to the dependency
  picker. `genitz init` / `genitz add` force either flow explicitly.
- **120 dependencies across 16 categories** — web frameworks, ORMs, raw SQL
  and NoSQL/search drivers, caches, message brokers, observability, auth,
  dependency injection, config loaders, CLI-building libraries, background
  job queues, blockchain SDKs, cloud/infra SDKs, payments, testing, desktop
  & game dev, and general utilities. Whatever corner of Go you work in, it
  should be in here — no more tabbing out to search for a package name.
  See the [full list](#dependency-registry) below.
- **Collapsible picker, not a wall of text.** Categories collapse to one
  line by default; open one at a time, with a live "(N selected)" count so a
  choice is never hidden. Type `/` to search across everything at once.
- **Responsive TUI.** Renders cleanly whether your terminal is a small split
  pane or a maximized window — no ASCII-art logo shoving your content off
  screen, no cursor scrolling out of view.
- **Animated install screen, built into the wizard.** Confirming on Review
  moves to its own "Installing" step — a live spinner runs next to whichever
  `go mod`/`go get` command is in flight, each one collapsing to a checkmark
  as it finishes. A failure shows a cross plus the real captured output, no
  raw log spam either way.
- **Doc links, not guesswork.** Every dependency on the Review screen — and
  in the terminal output after install — comes with its
  [pkg.go.dev](https://pkg.go.dev) reference link.
- **Optional Docker setup (new projects only).** Toggle it on and Genitz
  generates a multistage `Dockerfile` + `.dockerignore`, and — if any
  selected dependency has a real backing service (Redis, Postgres, MySQL,
  MongoDB, RabbitMQ, Kafka, and a few more) — a `docker-compose.yml` wiring
  those services up alongside the app. No infra-backed deps selected? No
  compose file gets written; it wouldn't add anything over the Dockerfile
  alone.
- **Optional CI, Makefile, and git init (new projects only).** A checkbox
  screen after Docker: a GitHub Actions workflow mirroring Genitz's own
  build/vet/test/gofmt loop, a Makefile whose docker targets match whatever
  Docker actually produced, and `git init` + a first commit (fully local —
  publishing to GitHub is *suggested*, never run automatically).
- **Automatic config loader stub.** Pick exactly one config library
  (Viper/godotenv/Koanf/envconfig) and Genitz writes a matching
  `config/config.go`. Pick none, or more than one, and nothing is
  generated — no guessing at what you meant.
- **Scriptable.** `genitz init --name x --deps fiber,redis --docker --ci`
  skips the wizard entirely — no TTY assumptions, safe for CI.
- **No template lock-in.** Genitz writes a bare `main.go` and lets `go get`
  do the rest — it doesn't force an opinionated project layout on you.

## Installation

Requires Go 1.25+.

```sh
go install github.com/koriebruh/Genitz@latest
```

This puts a `genitz` (or `Genitz`, depending on your `go install` naming)
binary in `$(go env GOPATH)/bin` — make sure that's on your `PATH`.

Prebuilt binaries for macOS/Linux/Windows are published to
[GitHub Releases](https://github.com/koriebruh/Genitz/releases) via
`.goreleaser.yaml` on every `v*` tag push (`.github/workflows/release.yml`).
A Homebrew tap (`brew install koriebruh/tap/genitz`) is configured in the
same pipeline but not live yet — it needs a separate `homebrew-tap`
repository created first.

## Usage

```text
genitz             Auto-detect: start a new project, or add dependencies if
                    go.mod already exists in the current directory.
genitz init         Always start the new-project wizard.
genitz add          Add dependencies to the project in the current directory
                    (requires an existing go.mod).
genitz remove       Remove dependencies from the current project.
genitz list         List direct dependencies already in go.mod (--json supported).
genitz version      Print the genitz version.
genitz completion   Print a shell completion script (bash|zsh|fish).
genitz doctor       Check the local environment (go/git/docker/gh, GOPROXY, network).
genitz audit        Check your dependencies against the curated registry for
                    maintenance-mode packages (e.g. gorilla/mux -> chi), plus
                    a govulncheck advisory.
genitz undo         Revert go.mod/go.sum to the state before the last add/remove.
genitz config       Get/set persistent defaults (license, author, modulePrefix).
genitz search       Search the dependency registry by name/category/description.
genitz info         Show details for one registry dependency.
genitz preset       List/save presets, or import a team's from a URL with
                    `preset import <url>`.
genitz history      Show a log of past init/add/remove operations.
genitz help         Show usage.
```

### Start a new project

```sh
mkdir my-app && cd my-app   # optional — genitz creates the folder for you
genitz init
```

Walks you through: folder name → module path → dependency picker → optional
Docker setup → optional CI/Makefile/git init → review → scaffold. You end up
with a `go.mod`, a minimal `main.go`, whatever packages you picked already
`go get`-installed, a `Dockerfile` (+ `docker-compose.yml`) if you opted into
Docker, a CI workflow / Makefile / git repo for whichever of those you
checked, and a `config/config.go` stub if you picked exactly one config
library.

### Non-interactive (scripting/CI)

```sh
genitz init --name my-app --module github.com/me/my-app \
  --deps fiber,redis --docker --ci --makefile --git --readme --license mit
genitz add --deps zap,testify@v1.9.0 --preset web-api
genitz remove --deps zap --dry-run
```

Passing `--name` (init) or `--deps`/`--preset` (init/add/remove) skips the
wizard entirely — no TTY, no Bubble Tea program, just plain progress lines.
Unknown dependency IDs fail fast with a clear message instead of guessing.
Pin a version with `id@version` (e.g. `redis@v9.5.1`); apply a starter
bundle with `--preset` (`web-api`, `grpc-service`, `auth-service`,
`cli-tool` — combines with `--deps`, doesn't replace it); add `--dry-run` to
any of `init`/`add`/`remove` to print what would run without doing it.

### Add or remove dependencies in an existing project

```sh
cd my-existing-project   # anywhere with a go.mod
genitz add       # picker, scoped to the whole registry
genitz remove    # picker, scoped to what's already installed
genitz list      # what's in go.mod right now
```

Same picker, straight to it — no folder/module questions since the project
already exists. `add` runs `go get` + `go mod tidy`; `remove` runs
`go get <path>@none` + `go mod tidy` per package when you confirm.

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
<summary>All 120 packages, grouped (click to expand)</summary>

| Group | Packages |
|---|---|
| Web / Routing | Fiber, Gin Gonic, Echo, Chi, gRPC, Protocol Buffers, gqlgen, Twirp, Connect, Gorilla Mux, Gorilla WebSocket, Alice |
| Database | GORM, sqlx, Bun, Ent, Squirrel, MySQL Driver, pgx (PostgreSQL), modernc SQLite, MongoDB Driver, gocql (Cassandra), Elasticsearch, ClickHouse Driver, AWS DynamoDB, etcd, Badger, Neo4j, Bleve, golang-migrate |
| Cache | redis, gomemcache, Ristretto, BigCache, go-cache, groupcache |
| Messaging | kafka-go, Sarama, NATS, RabbitMQ (amqp091), Watermill, Google Pub/Sub |
| Observability | Uber Zap, Logrus, Zerolog, OpenTelemetry, Prometheus client, Sentry, DataDog statsd |
| Security | golang-jwt, golang.org/x/crypto, Casbin, golang.org/x/oauth2, go-oidc, go-jose |
| Dependency Injection | Uber Dig, Google Wire, Uber Fx |
| Configuration | Viper, godotenv, Koanf, envconfig |
| CLI Tools | Cobra, urfave/cli, pflag, promptui, Survey |
| Background Jobs | Asynq, River, gocraft/work |
| Blockchain | go-ethereum, solana-go, btcd, Cosmos SDK, Hyperledger Fabric SDK, CometBFT |
| Cloud & Infra SDKs | AWS SDK v2, Google Cloud SDK, Azure SDK, MinIO client, client-go (Kubernetes), Consul API, Vault API, Docker client |
| Payments | Stripe, PayPal |
| Testing | testify, go.uber.org/mock, Ginkgo, Gomega, gofakeit, go-cmp, httpexpect, Gock |
| Desktop & Game Dev | Fyne, Ebiten, Wails, raylib-go |
| Utilities | go playground validator, Swaggo, Google UUID, shopspring decimal, oklog/run, robfig/cron, golang.org/x/time, sony/gobreaker, Resty, go-money, Excelize, jinzhu/copier, msgpack, golang.org/x/sync, mapstructure, Sprig, gocsv, go-mail, gofpdf, samber/lo, ants, avast/retry-go |

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
