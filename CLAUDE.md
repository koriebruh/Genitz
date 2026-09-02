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
- `genitz init` — the init wizard, regardless of cwd. With flags
  (`--name`/`--deps` present) it skips the TUI entirely — see
  `runInitFlags`/`runAddFlags`/`resolveDeps`/`runStepsPlain` in `main.go`,
  which reuse `BuildInstallSteps`/`BuildAddSteps` directly (they only need a
  `Requirement`, no `tui.Model`) and just print plain progress lines.
- `genitz add` — the add wizard; errors out if cwd has no `go.mod` (told to
  run `go mod init` or `genitz init` instead of silently guessing). Flags
  work the same way as init.
- `genitz help` — usage text.

There is no project-template/architecture-scaffolding feature (removed on
purpose — see git history around "hapus dulu aja fokus untuk cli
depedency"). Only bare project init + dependency install exist today.

## Layout

- `main.go` — entry point + subcommand dispatch (`runInit`/`runAdd`/
  `printUsage`). Builds the right `tui.Model` and injects `BuildSteps` (and,
  for Init, `LookupComposeServices`) closures into it — this is how `tui`
  and `generator` talk without an import cycle (`generator` already imports
  `tui` for `tui.Dependency`/`tui.InstallStep`, so `tui` can't import back).
- `internal/tui/` — the wizard UI.
  - `main_view.go` — `Model`, step state machine, key handlers, panel views.
  - `dependencies.go` — `Dependency` struct, `DependencyRegistry` (the
    package catalog), category badges, `depGroups`, the accordion row model
    (`visibleRows`/`activateCursorRow`/`rowWindow`), `docURL`.
  - `install.go` — `InstallStep` (label + `Run func() error`), the animated
    `StepInstalling` screen driven by `stepDoneMsg`/`spinner.TickMsg`.
  - `docker_step.go` — the `StepDocker` toggle screen (Init mode only).
  - `mode.go` — `Mode` (Init/Add) and `Step` enums.
  - `splash.go` / `styles.go` — ASCII logo, lipgloss styles.
- `internal/generator/` — non-interactive scaffolding logic:
  - `generate.go` — `Requirement`, `PrepareNewProject` (fast/sync: creates
    the dir + `main.go`), `BuildInstallSteps`/`BuildAddSteps` (return
    `[]tui.InstallStep` for the TUI to run and animate), `PrintDocs`.
  - `docker.go` — Dockerfile/`.dockerignore`/`docker-compose.yml` content
    generation, called from `BuildInstallSteps` when `Requirement.IncludeDocker`.

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

`DependencyRegistry` in `internal/tui/dependencies.go` is loaded from
`internal/tui/registry.json` via `go:embed` (120 entries as of writing) — add
a JSON object there: `ID`, `Name`, `Category`, `ImportPath` (must be
`go get`-able), `Description`. If the category is new, add a `Cat*` const, a
case in `getBadgeStyle`, and make sure it's listed in `depGroups` — an
unlisted category never appears in the picker even if it's in the registry.
Doc links aren't stored per entry — `docURL(importPath)` derives a
pkg.go.dev link, shown on the Review step and printed after install, so
there's nothing to keep in sync there. `FindByID` looks up a registry entry
by ID — used by the non-interactive `--deps` flag.

## Extras: CI / Makefile / git init / config stub

`StepExtras` (Init mode only, between `StepDocker` and `StepReview`) is a
3-item checkbox list — `Model.IncludeCI`/`IncludeMakefile`/`IncludeGitInit` —
mirroring `StepDocker`'s shape but for multiple items (`extras_step.go`).
`BuildInstallSteps` appends their generation as ordinary `InstallStep`s:
- CI (`ci.go`) and Makefile (`makefile.go`) go in right after the Docker
  block — Makefile's docker targets (`docker-up`/`docker-down` vs
  `docker-build`/`docker-run`) depend on whether Docker actually produced a
  compose file, so that bool is threaded through from the Docker block.
- The config loader stub (`config_stub.go`) has **no toggle at all** — same
  "derive from what's already selected" pattern as Docker's compose mapping.
  `detectConfigLib` only fires when exactly one of
  `viper`/`godotenv`/`koanf`/`envconfig` is selected (ambiguous or absent →
  nothing generated). Runs after the per-dependency `go get` loop, before
  the final tidy, so the newly-referenced import resolves cleanly. Each
  stub's exact API was verified against the real module in a scratch
  `go build` before being written — don't add a fifth library here without
  doing the same (see git history for how koanf's `v2` provider API was
  checked).
- git init (`gitinit.go`) is the very last step — it commits the finished
  scaffold, not a partial one. It's fully local (`git init` + `.gitignore` +
  `git add -A` + `git commit`); **`gh repo create` is never run
  automatically** — `SuggestedPublishCommand()` just returns the command
  text, printed by `main.go` after a successful scaffold. `InstallStep`
  failures are fatal, so anything depending on external auth state
  (a missing/unauthenticated `gh`) can't be allowed to abort an
  otherwise-successful run.

## Docker support

Init mode only (`genitz add` never offers it — no clean way to merge into an
existing Docker setup). `StepDocker` sits between `StepDeps` and `StepReview`
in the wizard; `Model.IncludeDocker` is the toggle. When on,
`BuildInstallSteps` (in `generate.go`) appends two more `InstallStep`s right
after the initial `go mod tidy` — a Dockerfile step always runs, a
docker-compose step only if `composeContent` finds a match — so `go.mod`
already exists by then and the Dockerfile can read the real Go version out
of it (`readGoVersion` in `docker.go`) instead of guessing.

`docker.go`'s `dockerComposeServices` map is what decides which selected
dependencies get a compose service — keyed by `Dependency.ID`, same shape as
the registry itself. Add an entry there when you add a registry dependency
that has an obvious, safe-to-default local container image; embedded/
in-process libraries and anything too opinionated to auto-spin-up (secrets
managers, k8s clients, ...) are deliberately left out. The picker's live
preview (`docker_step.go`'s `dockerPreview`) calls this same table through
`generator.ComposeServiceNames`, injected into `Model.LookupComposeServices`
by `main.go` — same import-cycle-avoidance pattern as `BuildSteps`.

## Commands

```sh
go build ./...
go vet ./...
go test ./...
gofmt -l .   # should print nothing
```
