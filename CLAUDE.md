# Genitz

Go CLI (Bubble Tea TUI, styled like Claude Code) with two flows:

- **Init** — scaffold a new project: wizard asks for folder name, module
  path, and dependencies, then writes a bare `main.go` + `go mod init` +
  `go get` for the chosen deps.
- **Add** — install dependencies into the project already in the current
  directory: wizard jumps straight to the dependency picker and runs
  `go get` + `go mod tidy` there.

## Commands (`main.go`)

- `genitz` — auto-detect: no `go.mod` in cwd runs init, `go.mod` present
  runs add. Kept for backward compatibility / zero-thought default.
- `genitz init` — always the init wizard, regardless of cwd.
- `genitz add` — always the add wizard; errors out if cwd has no `go.mod`
  (told to run `go mod init` or `genitz init` instead of silently guessing).
- `genitz help` — usage text.

There is no project-template/architecture-scaffolding feature (removed on
purpose — see git history around "hapus dulu aja fokus untuk cli
depedency"). Only bare project init + dependency install exist today.

## Layout

- `main.go` — entry point + subcommand dispatch (`runInit`/`runAdd`/
  `printUsage`), builds the right `tui.Model`, runs the Bubble Tea program,
  then calls `generator.GenerateNewProject` or `generator.AddDependencies`.
- `internal/tui/` — the wizard UI.
  - `main_view.go` — `Model`, step state machine, key handlers, panel views.
  - `dependencies.go` — `Dependency` struct, `DependencyRegistry` (the
    package catalog), category badges, `depGroups`, the accordion row model
    (`visibleRows`/`activateCursorRow`/`rowWindow`), `docURL`.
  - `mode.go` — `Mode` (Init/Add) and `Step` enums.
  - `splash.go` / `styles.go` — ASCII logo, lipgloss styles.
- `internal/generator/` — non-interactive scaffolding logic (`generate.go`),
  no TUI imports beyond the `tui.Dependency`/`tui.Model` types it consumes.

## Dependency picker UX

The picker is an accordion, not a flat list — group headers are collapsed by
default (`Model.ExpandedGroup`, -1 = none), space expands one group at a
time and shows a "(N selected)" count on collapsed headers so a selection is
never hidden. Searching auto-expands every group with a match. The row list
is windowed around the cursor against real terminal height (`rowWindow` in
`dependencies.go`) with "N more above/below" indicators — this is what stops
the cursor running off the bottom of a short terminal. If you add rows to
any per-row-line render path, sanity-check `rowBudget`'s reserved-chrome
constant still leaves no overflow (see Responsive TUI below on how to test).

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
the picker even if it's in the registry. Doc links aren't stored per entry —
`docURL(importPath)` derives a pkg.go.dev link, shown on the Review step and
printed after install, so there's nothing to keep in sync there.

There's no external registry file yet (package metadata is hardcoded Go) —
a data-file-backed registry is a planned future step, not implemented.

## Commands

```sh
go build ./...
go vet ./...
go test ./...
gofmt -l .   # should print nothing
```
