package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

func main() {
	args := os.Args[1:]

	if err := generator.CheckBinary("go"); err != nil {
		fmt.Printf("\n%v — genitz shells out to the go toolchain for every command.\n", err)
		os.Exit(1)
	}

	switch {
	case len(args) == 0:
		// No subcommand — auto-detect: go.mod present in cwd means add
		// dependencies, otherwise scaffold a new project.
		if _, err := generator.ReadModulePath("."); err == nil {
			runAdd(nil)
		} else {
			runInit(nil)
		}
	case args[0] == "init":
		runInit(args[1:])
	case args[0] == "add":
		runAdd(args[1:])
	case args[0] == "remove":
		runRemove(args[1:])
	case args[0] == "list":
		runList()
	case args[0] == "version", args[0] == "--version", args[0] == "-v":
		fmt.Println("genitz version " + generator.Version)
	case args[0] == "completion":
		runCompletion(args[1:])
	case args[0] == "help", args[0] == "-h", args[0] == "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command %q\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Genitz — Go project starter & dependency picker

Usage:
  genitz             Auto-detect: start a new project, or add dependencies if
                      go.mod already exists in the current directory.
  genitz init         Start the new-project wizard.
  genitz add          Add dependencies to the project in the current directory
                      (requires an existing go.mod).
  genitz remove       Remove dependencies from the project in the current
                      directory (requires an existing go.mod).
  genitz list         List direct dependencies already in go.mod.
  genitz version      Print the genitz version.
  genitz completion   Print a shell completion script (bash|zsh|fish).
  genitz help         Show this message.

Non-interactive (scripting/CI) — pass --name or --deps to skip the wizard:
  genitz init --name my-app --module github.com/me/my-app --deps fiber,redis \
              [--docker] [--ci] [--makefile] [--git] [--readme] [--license mit] \
              [--preset web-api] [--dry-run]
  genitz add --deps redis,zap@v9.5.1 [--preset web-api] [--dry-run]
  genitz remove --deps redis,zap [--dry-run]

  --name      folder name (required for non-interactive init)
  --module    Go module path (defaults to --name)
  --deps      comma-separated dependency IDs — see internal/tui/dependencies.go
              — pin a version with id@version (e.g. redis@v9.5.1)
  --preset    apply a starter bundle (web-api, grpc-service, auth-service,
              cli-tool) — combines with --deps rather than replacing it
  --docker    generate a multistage Dockerfile (+ docker-compose.yml if warranted)
  --ci        generate a GitHub Actions CI workflow (+ .golangci.yml)
  --makefile  generate a Makefile
  --git       git init + first commit (init only)
  --readme    generate README.md (init only)
  --license   generate a LICENSE file: mit or apache-2.0 (init only)
  --dry-run   print the steps that would run without executing them`)
}

// runInit walks the new-project wizard, or — when flags are passed — scaffolds
// non-interactively with no TUI at all.
func runInit(args []string) {
	if len(args) > 0 {
		runInitFlags(args)
		return
	}

	model := tui.InitialModel()
	model.LookupComposeServices = generator.ComposeServiceNames
	model.BuildSteps = func(m *tui.Model) []tui.InstallStep {
		req, err := generator.NewRequirementFromModel(m)
		if err != nil {
			return []tui.InstallStep{{Label: "Validate input", Run: func() error { return err }}}
		}
		targetPath, err := generator.PrepareNewProject(req)
		if err != nil {
			return []tui.InstallStep{{Label: "Create project directory", Run: func() error { return err }}}
		}
		return generator.BuildInstallSteps(targetPath, req)
	}

	finalModel := runProgram(model)

	if finalModel.InstallErr != nil {
		folder := finalModel.FolderInput.Value()
		// Only clean up if we actually created the directory (InstallIndex
		// advanced past the "create project directory" step) — never touch
		// a pre-existing folder we merely failed to use as a target.
		if finalModel.InstallIndex > 0 {
			if removeErr := os.RemoveAll(folder); removeErr != nil {
				fmt.Printf("Warning: could not remove partial output %q: %v\n", folder, removeErr)
			} else {
				fmt.Printf("Cleaned up partial directory: ./%s/\n", folder)
			}
		}
		fmt.Printf("\nGagal membuat project: %v\n", finalModel.InstallErr)
		os.Exit(1)
	}

	if !finalModel.Done {
		fmt.Println("\nCancelled.")
		return
	}

	fmt.Printf("\n📂 Project tersedia di: ./%s\n", finalModel.FolderInput.Value())
	generator.PrintDocs(finalModel.Chosen)
	if finalModel.IncludeGitInit {
		fmt.Printf("\n📦 To publish: %s\n", generator.SuggestedPublishCommand())
	}
}

// runInitFlags scaffolds a project directly from flags — no Bubble Tea
// program at all, safe for CI/scripts (no TTY assumptions).
func runInitFlags(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	name := fs.String("name", "", "project folder name (required)")
	module := fs.String("module", "", "Go module path (defaults to --name)")
	depsFlag := fs.String("deps", "", "comma-separated dependency IDs")
	presetFlag := fs.String("preset", "", "starter bundle to apply (combines with --deps)")
	docker := fs.Bool("docker", false, "generate Docker setup")
	ci := fs.Bool("ci", false, "generate GitHub Actions CI workflow")
	makefile := fs.Bool("makefile", false, "generate Makefile")
	gitInit := fs.Bool("git", false, "git init + first commit")
	readme := fs.Bool("readme", false, "generate README.md")
	license := fs.String("license", "", "LICENSE to generate: mit or apache-2.0")
	dryRun := fs.Bool("dry-run", false, "print steps without running them")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("\n--name is required for non-interactive init")
		os.Exit(1)
	}
	if !generator.ValidLicenseKind(*license) {
		fmt.Printf("\nunknown --license %q — expected \"mit\" or \"apache-2.0\"\n", *license)
		os.Exit(1)
	}
	modulePath := *module
	if modulePath == "" {
		modulePath = *name
	}

	deps, versions, err := resolveDeps(*depsFlag)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
	presetDeps, err := resolvePresetDeps(*presetFlag)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
	deps = mergeDeps(presetDeps, deps)

	req := generator.Requirement{
		ProjectName:     *name,
		PackageName:     modulePath,
		Deps:            deps,
		IncludeDocker:   *docker,
		IncludeCI:       *ci,
		IncludeMakefile: *makefile,
		IncludeGitInit:  *gitInit,
		IncludeReadme:   *readme,
		License:         *license,
		DepVersions:     versions,
	}

	if *dryRun {
		req, err = generator.CheckPreconditions(req)
		if err != nil {
			fmt.Printf("\n%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[dry-run] would create ./%s/ (module %s)\n", req.ProjectName, req.PackageName)
		if err := runStepsPlain(generator.BuildInstallSteps(req.ProjectName, req), true); err != nil {
			fmt.Printf("\n%v\n", err)
			os.Exit(1)
		}
		return
	}

	targetPath, err := generator.PrepareNewProject(req)
	if err != nil {
		fmt.Printf("\nGagal membuat project: %v\n", err)
		os.Exit(1)
	}

	if err := runStepsPlain(generator.BuildInstallSteps(targetPath, req), false); err != nil {
		if removeErr := os.RemoveAll(req.ProjectName); removeErr != nil {
			fmt.Printf("Warning: could not remove partial output %q: %v\n", req.ProjectName, removeErr)
		} else {
			fmt.Printf("Cleaned up partial directory: ./%s/\n", req.ProjectName)
		}
		fmt.Printf("\nGagal membuat project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n📂 Project tersedia di: ./%s\n", req.ProjectName)
	generator.PrintDocs(req.Deps)
	if req.IncludeGitInit {
		fmt.Printf("\n📦 To publish: %s\n", generator.SuggestedPublishCommand())
	}
}

// runAdd walks the dependency picker, or — when flags are passed — installs
// non-interactively with no TUI at all.
func runAdd(args []string) {
	if len(args) > 0 {
		runAddFlags(args)
		return
	}

	module, err := generator.ReadModulePath(".")
	if err != nil {
		fmt.Println("\nGak ada go.mod di direktori ini. Jalanin `go mod init <module>` dulu, atau pakai `genitz init` buat project baru.")
		os.Exit(1)
	}

	model := tui.AddModel(module)
	model.BuildSteps = func(m *tui.Model) []tui.InstallStep {
		return generator.BuildAddSteps(".", m.Chosen, nil)
	}

	finalModel := runProgram(model)

	if finalModel.InstallErr != nil {
		fmt.Printf("\nGagal menambahkan dependency: %v\n", finalModel.InstallErr)
		os.Exit(1)
	}

	if !finalModel.Done {
		fmt.Println("\nCancelled.")
		return
	}

	fmt.Println("\n✅ Dependencies installed ✨")
	generator.PrintDocs(finalModel.Chosen)
}

// runAddFlags installs dependencies directly from flags — no TUI.
func runAddFlags(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	depsFlag := fs.String("deps", "", "comma-separated dependency IDs (required unless --preset is set)")
	presetFlag := fs.String("preset", "", "starter bundle to apply (combines with --deps)")
	dryRun := fs.Bool("dry-run", false, "print steps without running them")
	fs.Parse(args)

	if *depsFlag == "" && *presetFlag == "" {
		fmt.Println("\n--deps or --preset is required for non-interactive add")
		os.Exit(1)
	}

	if _, err := generator.ReadModulePath("."); err != nil {
		fmt.Println("\nGak ada go.mod di direktori ini. Jalanin `go mod init <module>` dulu, atau pakai `genitz init` buat project baru.")
		os.Exit(1)
	}

	deps, versions, err := resolveDeps(*depsFlag)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
	presetDeps, err := resolvePresetDeps(*presetFlag)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
	deps = mergeDeps(presetDeps, deps)

	if err := runStepsPlain(generator.BuildAddSteps(".", deps, versions), *dryRun); err != nil {
		fmt.Printf("\nGagal menambahkan dependency: %v\n", err)
		os.Exit(1)
	}
	if *dryRun {
		return
	}

	fmt.Println("\n✅ Dependencies installed ✨")
	generator.PrintDocs(deps)
}

// runRemove walks the interactive removal picker (scoped to already-
// installed, registry-matched deps), or — when flags are passed — removes
// non-interactively with no TUI at all.
func runRemove(args []string) {
	if len(args) > 0 {
		runRemoveFlags(args)
		return
	}

	module, err := generator.ReadModulePath(".")
	if err != nil {
		fmt.Println("\nGak ada go.mod di direktori ini. Jalanin `go mod init <module>` dulu, atau pakai `genitz init` buat project baru.")
		os.Exit(1)
	}

	installed, err := generator.ListInstalled(".")
	if err != nil {
		fmt.Printf("\nGagal membaca go.mod: %v\n", err)
		os.Exit(1)
	}
	var removable []tui.Dependency
	for _, d := range installed {
		if !d.Managed {
			continue
		}
		if dep, ok := tui.FindByID(idFromImportPath(d.ImportPath)); ok {
			removable = append(removable, dep)
		}
	}
	if len(removable) == 0 {
		fmt.Println("\nNo registry-managed dependencies to remove.")
		return
	}

	model := tui.RemoveModel(module, removable)
	model.BuildSteps = func(m *tui.Model) []tui.InstallStep {
		return generator.BuildRemoveSteps(".", m.Chosen)
	}

	finalModel := runProgram(model)

	if finalModel.InstallErr != nil {
		fmt.Printf("\nGagal menghapus dependency: %v\n", finalModel.InstallErr)
		os.Exit(1)
	}
	if !finalModel.Done {
		fmt.Println("\nCancelled.")
		return
	}
	fmt.Println("\n✅ Dependencies removed")
}

// idFromImportPath recovers a registry ID from an import path for the
// removal picker — FindByID resolves the other direction, so this scans
// the registry for the matching entry rather than duplicating IDs.
func idFromImportPath(importPath string) string {
	for _, dep := range tui.DependencyRegistry {
		if dep.ImportPath == importPath {
			return dep.ID
		}
	}
	return ""
}

// runRemoveFlags removes dependencies directly from flags — no TUI.
func runRemoveFlags(args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	depsFlag := fs.String("deps", "", "comma-separated dependency IDs (required)")
	dryRun := fs.Bool("dry-run", false, "print steps without running them")
	fs.Parse(args)

	if *depsFlag == "" {
		fmt.Println("\n--deps is required for non-interactive remove")
		os.Exit(1)
	}

	if _, err := generator.ReadModulePath("."); err != nil {
		fmt.Println("\nGak ada go.mod di direktori ini. Jalanin `go mod init <module>` dulu, atau pakai `genitz init` buat project baru.")
		os.Exit(1)
	}

	deps, _, err := resolveDeps(*depsFlag)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}

	if err := runStepsPlain(generator.BuildRemoveSteps(".", deps), *dryRun); err != nil {
		fmt.Printf("\nGagal menghapus dependency: %v\n", err)
		os.Exit(1)
	}
	if *dryRun {
		return
	}

	fmt.Println("\n✅ Dependencies removed")
}

// runList prints the direct dependencies already in ./go.mod.
func runList() {
	if _, err := generator.ReadModulePath("."); err != nil {
		fmt.Println("\nGak ada go.mod di direktori ini.")
		os.Exit(1)
	}
	deps, err := generator.ListInstalled(".")
	if err != nil {
		fmt.Printf("\nGagal membaca go.mod: %v\n", err)
		os.Exit(1)
	}
	generator.PrintInstalled(deps)
}

// runCompletion prints a static shell completion script for the requested
// shell — hand-authored (no cobra in this codebase), covering every
// top-level subcommand so a new one added to main() should also be added
// to subcommandNames below (see main_test.go for the drift check).
func runCompletion(args []string) {
	if len(args) == 0 {
		fmt.Println("\nUsage: genitz completion <bash|zsh|fish>")
		os.Exit(1)
	}
	switch args[0] {
	case "bash":
		fmt.Print(completionBash())
	case "zsh":
		fmt.Print(completionZsh())
	case "fish":
		fmt.Print(completionFish())
	default:
		fmt.Printf("\nUnknown shell %q — expected bash, zsh, or fish\n", args[0])
		os.Exit(1)
	}
}

// subcommandNames are genitz's top-level subcommands, single source of
// truth for both the completion scripts below and main()'s dispatch.
var subcommandNames = []string{"init", "add", "remove", "list", "version", "completion", "help"}

func completionBash() string {
	return fmt.Sprintf(`_genitz_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=($(compgen -W "%s" -- "$cur"))
    fi
}
complete -F _genitz_completions genitz
`, strings.Join(subcommandNames, " "))
}

func completionZsh() string {
	return fmt.Sprintf(`#compdef genitz
_genitz() {
    local -a subcommands
    subcommands=(%s)
    _describe 'command' subcommands
}
_genitz
`, strings.Join(subcommandNames, " "))
}

func completionFish() string {
	var b strings.Builder
	for _, name := range subcommandNames {
		fmt.Fprintf(&b, "complete -c genitz -n '__fish_use_subcommand' -a %s\n", name)
	}
	return b.String()
}

// resolveDeps turns a comma-separated list of registry IDs (optionally
// pinned as id@version) into a Chosen-shaped map plus an import-path-keyed
// version map, failing fast with a clear message on the first unknown ID.
func resolveDeps(depsFlag string) (map[int]tui.Dependency, map[string]string, error) {
	deps := make(map[int]tui.Dependency)
	versions := make(map[string]string)
	depsFlag = strings.TrimSpace(depsFlag)
	if depsFlag == "" {
		return deps, versions, nil
	}
	for i, entry := range strings.Split(depsFlag, ",") {
		entry = strings.TrimSpace(entry)
		id, version, _ := strings.Cut(entry, "@")
		id = strings.TrimSpace(id)
		dep, ok := tui.FindByID(id)
		if !ok {
			return nil, nil, fmt.Errorf("unknown dependency ID %q — see internal/tui/dependencies.go for valid IDs", id)
		}
		deps[i] = dep
		if version != "" {
			versions[dep.ImportPath] = version
		}
	}
	return deps, versions, nil
}

// resolvePresetDeps expands presetFlag (a Preset.ID) into a Chosen-shaped
// map, failing fast on an unknown preset ID.
func resolvePresetDeps(presetFlag string) (map[int]tui.Dependency, error) {
	deps := make(map[int]tui.Dependency)
	presetFlag = strings.TrimSpace(presetFlag)
	if presetFlag == "" {
		return deps, nil
	}
	preset, ok := tui.FindPreset(presetFlag)
	if !ok {
		return nil, fmt.Errorf("unknown preset %q — see internal/tui/presets.go for valid presets", presetFlag)
	}
	for i, id := range preset.DepIDs {
		dep, ok := tui.FindByID(id)
		if !ok {
			// Presets are code-owned and built-in — an ID that doesn't
			// resolve means the registry changed out from under a preset,
			// a genitz bug, not user input to guess around.
			return nil, fmt.Errorf("preset %q references unknown dependency ID %q — this is a genitz bug, please report it", presetFlag, id)
		}
		deps[i] = dep
	}
	return deps, nil
}

// mergeDeps unions a and b by ImportPath (so a preset and an overlapping
// --deps entry never produce a duplicate); on a collision the entry from a
// wins since a is inserted first — call sites pass presets as a, so a
// preset's dependency wins over an identical one from --deps. This has no
// observable effect today: both sides resolve the same tui.Dependency via
// FindByID for a given ID, so "which one wins" only matters if that ever
// stops being true.
func mergeDeps(a, b map[int]tui.Dependency) map[int]tui.Dependency {
	seen := make(map[string]bool, len(a))
	merged := make(map[int]tui.Dependency, len(a)+len(b))
	i := 0
	for _, dep := range a {
		if seen[dep.ImportPath] {
			continue
		}
		seen[dep.ImportPath] = true
		merged[i] = dep
		i++
	}
	for _, dep := range b {
		if seen[dep.ImportPath] {
			continue
		}
		seen[dep.ImportPath] = true
		merged[i] = dep
		i++
	}
	return merged
}

// runStepsPlain runs InstallSteps sequentially with plain progress lines —
// no Bubble Tea, no TTY assumptions, safe for CI/scripted output. When
// dryRun is true, steps are printed but never executed.
func runStepsPlain(steps []tui.InstallStep, dryRun bool) error {
	for _, step := range steps {
		if dryRun {
			fmt.Printf("[dry-run] would run: %s\n", step.Label)
			continue
		}
		fmt.Printf("▶ %s...", step.Label)
		if err := step.Run(); err != nil {
			fmt.Println(" failed")
			return fmt.Errorf("%s: %w", step.Label, err)
		}
		fmt.Println(" done")
	}
	return nil
}

// runProgram runs the Bubble Tea wizard to completion and returns the final model.
func runProgram(model *tui.Model) *tui.Model {
	// WithAltScreen renders into the terminal's alternate buffer — this prevents
	// the "double logo" effect when the user zooms in/out in their terminal.
	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return m.(*tui.Model)
}
