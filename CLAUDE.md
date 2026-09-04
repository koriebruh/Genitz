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
- `genitz remove` — the removal wizard (`RemoveModel` in `main_view.go`,
  `RemoveMode` flag on `Model`), scoped to registry-matched deps already in
  `go.mod` (`generator.ListInstalled`). Runs `go get <path>@none` + `go mod
  tidy` per dep (`BuildRemoveSteps` in `generate.go`) — the documented way to
  drop a requirement; if the code still imports it, tidy re-adds it, which is
  correct (genitz doesn't scan source to know).
- `genitz list` — prints direct deps from `go.mod`'s require block
  (`generator.ListInstalled`/`PrintInstalled`), cross-referenced against the
  registry; entries with no registry match print as `(unmanaged)`. Note:
  since scaffolded projects only get a bare, import-free `main.go`, `go mod
  tidy` strips any dep the wizard installs until the user actually imports
  it — an empty `list` right after `init`/`add` is expected, not a bug.
- `genitz version` / `--version` / `-v` — prints `generator.Version`, a
  `var` (not `const` — ldflags `-X` can't target a const) overridden at
  release-build time by `.goreleaser.yaml`'s `-X .../generator.Version=...`;
  `go install`/local builds get the `"0.1.0-dev"` literal default.
- `genitz completion bash|zsh|fish` — static, hand-authored completion
  scripts (no cobra in this codebase) covering `subcommandNames` in
  `main.go`; `main_test.go`'s `TestCompletionScriptsCoverAllSubcommands`
  guards against that list drifting from the real dispatch.
- `genitz doctor` — environment diagnostics (`RunDoctor`/`PrintDoctor` in
  `doctor.go`): go (required)/git/docker/gh presence+version, `GOPROXY`,
  and TCP reachability to `proxy.golang.org:443`.
- `genitz config get|set <key>` — persistent user-level defaults
  (`license`/`author`/`modulePrefix`) at `$XDG_CONFIG_HOME/genitz/config.json`
  (`config.go`). `runInitFlags` applies them when the matching flag isn't
  passed; `license.go`'s `ResolveLicenseHolder` tries `git config user.name`
  first, config's `author` second.
- `genitz search <keyword>` / `genitz info <id>` — registry lookup outside
  the TUI (name/ID/category/description substring match; full entry detail).
- `genitz preset list|save` — `preset save <id> --deps a,b,c` persists a
  user-defined bundle to `$XDG_CONFIG_HOME/genitz/presets.json`
  (`internal/tui/preset_override.go`, same override-by-ID/append-new-ID
  merge shape as the registry override); `tui.AllPresets()` (built-in +
  user) backs the picker overlay, `--preset`, and this listing everywhere.
- `genitz history` — a local JSONL log of past `init`/`add`/`remove`
  operations (`history.go`), best-effort (`RecordHistory` swallows its own
  errors — a logging failure must never break the operation it's recording).
- `genitz help` — usage text.

All of `init`/`add`/`remove` accept `--dry-run` — for `add`/`remove` this
just skips `.Run()` on each `InstallStep` (see `runStepsPlain`'s `dryRun`
param); `init`'s dry-run also calls `generator.CheckPreconditions` first
(the same validate-and-check-destination logic `PrepareNewProject` runs
before actually creating anything), so a dry run can't report a clean
preview for a real run that would immediately fail. `add`/`init` also take
`--preset <id>` to expand a bundle from `internal/tui/presets.go` before
merging with `--deps` (`mergeDeps` dedupes by `ImportPath`; on a collision
the preset's entry wins since it's passed as the first argument — inert
today since both sides resolve an identical `tui.Dependency` via
`FindByID`). `main()` also runs a `generator.CheckBinary("go")` preflight
before any dispatch, and `Requirement.validate()` checks `git` specifically
when `IncludeGitInit` is set — both fail fast with a clear message instead
of a raw exec error surfacing mid-`InstallStep` animation.

There is no project-template/architecture-scaffolding feature (removed on
purpose — see git history around "hapus dulu aja fokus untuk cli
depedency"). Only bare project init + dependency install exist today.

## Logging

`main.go` uses four standardized output helpers instead of ad hoc
`fmt.Printf`/`Println` calls: `logError` (stderr, `✖` prefix), `logSuccess`
(`✔`), `logWarn` (`⚠`), `logInfo` (no prefix). Errors go to stderr
specifically so stdout stays clean for scripting (`genitz list --json`,
`genitz search`, etc.) — use these instead of a bare `fmt.Println` for any
new user-facing message. All user-facing strings are English — this is a
general-purpose tool, not developed only for one locale.

## Release pipeline

`.goreleaser.yaml` + `.github/workflows/release.yml` build/publish on every
`v*` tag push: cross-compiled binaries (linux/darwin/windows,
amd64/arm64), checksums, a GitHub Release, and a Homebrew formula push to a
`koriebruh/homebrew-tap` repo. That tap repo and a
`HOMEBREW_TAP_GITHUB_TOKEN` secret don't exist yet as of writing — the
`brews:` block will fail on a real release until both are set up; the
binaries/checksums/GitHub Release publish as independent goreleaser steps
regardless.

## Community files, .env.example, .editorconfig

- `.editorconfig` is written unconditionally by `PrepareNewProject` — no
  toggle, same treatment as `main.go` itself, since it has no real
  downside (unlike Docker/CI/etc., which are opt-in because they carry
  actual behavioral consequences).
- `.env.example` (`envExampleContent` in `docker.go`) is generated
  alongside `docker-compose.yml` only when a selected service actually
  declares `environment` entries — `docker-compose.yml` itself hardcodes
  real default values (e.g. `POSTGRES_PASSWORD=postgres`) with nothing
  telling a reader to externalize them; this is that nudge.
- `IncludeCommunityFiles` (one `StepExtras` checkbox, `--community` flag)
  bundles `CONTRIBUTING.md`, `SECURITY.md`, `.github/ISSUE_TEMPLATE/*.md`,
  and `.github/dependabot.yml` (`community.go`) — one toggle for all four
  rather than four separate ones.
- `VulnCheckAdvisory` (`vulncheck.go`) runs `govulncheck ./...` after a
  successful `init`/`add` if the binary is on PATH — advisory-only, never
  fatal, since `InstallStep` failures abort the whole run and a known
  vulnerability in a freshly-picked dependency shouldn't block scaffolding.

## Integration tests

`internal/generator/integration_test.go` runs the real `go`/`git` toolchain
end to end (init/add/remove against an actual temp-dir module) — unit
tests alone kept missing bugs that only surface once a real binary touches
disk (a `go.mod` parsing edge case was caught this way before shipping).
Skipped under `-short`; the deps-touching tests additionally skip if
`proxy.golang.org:443` isn't reachable, so they degrade gracefully offline
instead of flaking. genitz's own CI (`go test ./...`, no `-short`) runs
them for real.

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
  - `extras_step.go` — the `StepExtras` checkbox list (CI/Makefile/git
    init/README) plus the trailing License cycle row.
  - `presets.go` — `Preset`/`Presets`, `FindPreset`, `Model.applyPreset`,
    and the preset overlay rendered over `StepDeps` (`p` to open).
  - `registry_override.go` — merges an optional user/team registry file
    into the embedded `DependencyRegistry` at package init.
  - `mode.go` — `Mode` (Init/Add) and `Step` enums.
  - `splash.go` / `styles.go` — ASCII logo, lipgloss styles.
- `internal/generator/` — non-interactive scaffolding logic:
  - `generate.go` — `Requirement`, `PrepareNewProject` (fast/sync: creates
    the dir + `main.go`), `BuildInstallSteps`/`BuildAddSteps`/
    `BuildRemoveSteps` (return `[]tui.InstallStep` for the TUI to run and
    animate), `PrintDocs`.
  - `docker.go` — Dockerfile/`.dockerignore`/`docker-compose.yml` content
    generation, called from `BuildInstallSteps` when `Requirement.IncludeDocker`.
  - `list.go` — `ListInstalled`/`PrintInstalled`, backing `genitz list`.
  - `readme.go` / `license.go` — README.md/LICENSE content generators.
  - `version.go` — `Version` constant.
  - `preflight.go` — `CheckBinary`, used for the `go`/`git` checks above.

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

## Presets, README/LICENSE, version pinning, registry overrides

- **Presets** (`internal/tui/presets.go`) are named bundles of registry IDs
  (`web-api`, `grpc-service`, `auth-service`, `cli-tool` as of writing).
  `Model.applyPreset` unions a preset's deps into `Chosen` — it never
  removes an existing pick, so applying one is always additive. In the
  wizard, `p` on `StepDeps` opens the overlay (`PresetOverlayOpen`); in
  flags mode, `--preset <id>` does the same via `resolvePresetDeps` +
  `mergeDeps` in `main.go`. IDs in `Preset.DepIDs` must exist in
  `registry.json` — verify with `tui.FindByID` before adding one.
- **README/LICENSE** — `IncludeReadme`/`LicenseChoice` on `Model`, cycled on
  the trailing `StepExtras` row (`extraItems` is the checkbox list;
  `licenseChoices`/`nextLicenseChoice` cycle the 3-way License row after
  it). `readmeContent`/`licenseContent` in `internal/generator` are pure
  string generators; `BuildInstallSteps` writes them right after the deps
  loop. `licenseContent` leaves `[COPYRIGHT HOLDER]`/`[year]` as literal
  placeholders rather than guessing a name — same "don't guess" spirit as
  `detectConfigLib`.
- **Version pinning** — flags-only (`--deps id@version`, parsed in
  `resolveDeps` in `main.go`), not exposed in the TUI (no text input on the
  picker). `Requirement.DepVersions` maps import path → version;
  `installArg` in `generate.go` appends `@version` to the `go get` call
  when present, otherwise installs unpinned/latest — used by both
  `BuildInstallSteps` and `BuildAddSteps`.
- **User/team registry override** (`internal/tui/registry_override.go`) —
  an optional `$XDG_CONFIG_HOME/genitz/registry.json` (same JSON shape as
  the embedded `registry.json`) merged into `DependencyRegistry` at package
  init: matching `ID` overrides in place, a new `ID` appends. Unlike the
  embedded file (`mustLoadRegistry`, panics on bad JSON — that'd be a
  genitz bug), this is untrusted external input, so a missing file,
  invalid JSON, or an entry missing a required field is skipped with a
  stderr warning instead of crashing the CLI.
- **golangci-lint scaffold** — whenever `IncludeCI` is set, `BuildInstallSteps`
  also writes `.golangci.yml` (`golangciConfigContent` in `ci.go`) and the
  generated `ci.yml` gets a `golangci-lint-action` step — same "derive from
  what's already selected, no separate toggle" pattern as the config stub.

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
