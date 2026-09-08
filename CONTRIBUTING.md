# Contributing to Genitz

Thanks for considering a contribution to Genitz — a Go CLI that scaffolds new
projects and manages dependencies through a curated registry.

## Getting started

1. Fork the repository and create a branch off `main`.
2. Make your change, with tests where it makes sense.
3. Run the same checks CI runs before opening a PR:

   ```sh
   go build ./...
   go vet ./...
   go test ./...
   gofmt -l .   # should print nothing
   ```

4. Open a pull request describing what changed and why.

## Adding a dependency to the registry

`internal/tui/registry.json` is the curated catalog the picker reads from —
see the "Adding a dependency to the registry" section in `CLAUDE.md` for the
exact fields required and where else a new category needs wiring up.

## Changing the TUI

If your change touches rendering (`internal/tui/*.go`), verify it at a few
terminal sizes before opening a PR — there's no headless TUI test here. See
"Responsive TUI" in `CLAUDE.md` for a quick `tmux` recipe.

## Reporting a bug

Open an issue with the command you ran, what you expected, and what
happened instead — `genitz doctor` output is often useful context for
environment-related issues.
