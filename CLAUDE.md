# Genitz

Go CLI (Bubble Tea TUI, styled like Claude Code) with two modes, auto-detected
from the current directory:

- **Init mode** — no `go.mod` in cwd: wizard asks for folder name, module
  path, and dependencies, then scaffolds a bare `main.go` + `go mod init` +
  `go get` for the chosen deps.
- **Add mode** — `go.mod` already exists in cwd: wizard jumps straight to the
  dependency picker and runs `go get` + `go mod tidy` against the project
  already there.

There is no project-template/architecture-scaffolding feature (removed on
purpose — see git history around "hapus dulu aja fokus untuk cli
depedency"). Only bare project init + dependency install exist today.

## Layout

- `main.go` — entry point. Detects mode via `generator.ReadModulePath(".")`,
  builds the right `tui.Model`, runs the Bubble Tea program, then calls
  `generator.GenerateNewProject` or `generator.AddDependencies`.
- `internal/tui/` — the wizard UI.
  - `main_view.go` — `Model`, step state machine, key handlers, panel views.
  - `dependencies.go` — `Dependency` struct, `DependencyRegistry` (the
    package catalog), category badges, `depGroups` (display grouping).
  - `mode.go` — `Mode` (Init/Add) and `Step` enums.
  - `splash.go` / `styles.go` — ASCII logo, lipgloss styles.
- `internal/generator/` — non-interactive scaffolding logic (`generate.go`),
  no TUI imports beyond the `tui.Dependency`/`tui.Model` types it consumes.

## Responsive TUI

Bubble Tea's alt-screen clips anything wider than the terminal instead of
soft-wrapping it. `renderFrame` in `main_view.go` picks the header (full
logo / compact brand line / none) and step-nav format (full labels / bare
numbers) based on both `Width` and `Height`, and the body `Container` gets a
`Width()` set every render so lipgloss wraps long lines instead of the
terminal clipping them. When touching layout, verify at multiple sizes —
there's no headless TUI test here, so use tmux:

```sh
go build -o /tmp/genitz .
tmux new-session -d -s t -x 40 -y 24 "cd /some/test/dir && /tmp/genitz"
tmux capture-pane -t t -p
```

## Adding a dependency to the registry

Add an entry to `DependencyRegistry` in `internal/tui/dependencies.go`:
`ID`, `Name`, `Category`, `ImportPath` (must be `go get`-able), `Description`.
If the category is new, add a `Cat*` const, a case in `getBadgeStyle`, and
make sure it's listed in `depGroups` — an unlisted category never appears in
the picker even if it's in the registry.

There's no external registry file yet (package metadata is hardcoded Go) —
a data-file-backed registry is a planned future step, not implemented.

## Commands

```sh
go build ./...
go vet ./...
go test ./...
gofmt -l .   # should print nothing
```
