package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

// ── Logging ───────────────────────────────────────────────────────────────
// Standardized output helpers used throughout instead of ad hoc
// fmt.Printf/Println calls: errors go to stderr (idiomatic CLI behavior —
// keeps stdout clean for scripting/piping), everything else to stdout,
// all through one consistent prefix style.

func logError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "✖ "+format+"\n", args...)
}

func logSuccess(format string, args ...any) {
	fmt.Printf("✔ "+format+"\n", args...)
}

func logWarn(format string, args ...any) {
	fmt.Printf("⚠ "+format+"\n", args...)
}

func logInfo(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func main() {
	args := os.Args[1:]

	if err := generator.CheckBinary("go"); err != nil {
		logError("%v — genitz shells out to the go toolchain for every command.", err)
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
		runList(args[1:])
	case args[0] == "version", args[0] == "--version", args[0] == "-v":
		logInfo("genitz version %s", generator.Version)
	case args[0] == "completion":
		runCompletion(args[1:])
	case args[0] == "doctor":
		generator.PrintDoctor(generator.RunDoctor())
	case args[0] == "config":
		runConfig(args[1:])
	case args[0] == "search":
		runSearch(args[1:])
	case args[0] == "info":
		runInfo(args[1:])
	case args[0] == "preset":
		runPreset(args[1:])
	case args[0] == "history":
		history, err := generator.ReadHistory()
		if err != nil {
			logError("Failed to read history: %v", err)
			os.Exit(1)
		}
		generator.PrintHistory(history)
	case args[0] == "help", args[0] == "-h", args[0] == "--help":
		printUsage()
	default:
		logError("Unknown command %q", args[0])
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
  genitz list         List direct dependencies already in go.mod (--json for
                      machine-readable output).
  genitz version      Print the genitz version.
  genitz completion   Print a shell completion script (bash|zsh|fish).
  genitz doctor       Check the local environment (go/git/docker/gh, GOPROXY,
                      network reachability).
  genitz config       Get/set persistent defaults (license, author,
                      modulePrefix) — see below.
  genitz search       Search the dependency registry by name/category/description.
  genitz info         Show details for one registry dependency.
  genitz preset       List or save dependency-bundle presets.
  genitz history      Show a log of past init/add/remove operations.
  genitz help         Show this message.

Non-interactive (scripting/CI) — pass --name or --deps to skip the wizard:
  genitz init --name my-app --module github.com/me/my-app --deps fiber,redis \
              [--docker] [--ci] [--makefile] [--git] [--readme] [--community] \
              [--license mit] [--preset web-api] [--dry-run]
  genitz add --deps redis,zap@v9.5.1 [--preset web-api] [--dry-run]
  genitz remove --deps redis,zap [--dry-run]

  --name       folder name (required for non-interactive init)
  --module     Go module path (defaults to --name, or config's modulePrefix+name)
  --deps       comma-separated dependency IDs — see internal/tui/dependencies.go
               — pin a version with id@version (e.g. redis@v9.5.1)
  --preset     apply a starter bundle (see ` + "`genitz preset list`" + `) — combines
               with --deps rather than replacing it
  --docker     generate a multistage Dockerfile (+ docker-compose.yml, +
               .env.example if any service needs credentials)
  --ci         generate a GitHub Actions CI workflow (+ .golangci.yml)
  --makefile   generate a Makefile
  --git        git init + first commit (init only)
  --readme     generate README.md (init only)
  --community  generate CONTRIBUTING.md/SECURITY.md/issue templates/
               dependabot.yml (init only)
  --license    generate a LICENSE file: mit or apache-2.0 (init only) —
               defaults to config's license if set
  --dry-run    print the steps that would run without executing them

genitz config keys (genitz config set <key> <value> / genitz config get <key>):
  license        default --license value
  author         default LICENSE copyright holder (git config user.name wins
                 if set)
  modulePrefix   prepended to --name when --module isn't passed (e.g.
                 github.com/you/)`)
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
				logWarn("Could not remove partial output %q: %v", folder, removeErr)
			} else {
				logInfo("Cleaned up partial directory: ./%s/", folder)
			}
		}
		logError("Failed to create project: %v", finalModel.InstallErr)
		os.Exit(1)
	}

	if !finalModel.Done {
		logInfo("Cancelled.")
		return
	}

	logSuccess("Project available at: ./%s", finalModel.FolderInput.Value())
	generator.PrintDocs(finalModel.Chosen)
	if finalModel.IncludeGitInit {
		logInfo("📦 To publish: %s", generator.SuggestedPublishCommand())
	}
	generator.RecordHistory("init", finalModel.FolderInput.Value(), fmt.Sprintf("%d deps", len(finalModel.Chosen)))
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
	community := fs.Bool("community", false, "generate CONTRIBUTING/SECURITY/issue templates/dependabot.yml")
	license := fs.String("license", "", "LICENSE to generate: mit or apache-2.0")
	dryRun := fs.Bool("dry-run", false, "print steps without running them")
	fs.Parse(args)

	if *name == "" {
		logError("--name is required for non-interactive init")
		os.Exit(1)
	}

	cfg, err := generator.LoadConfig()
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}
	if *license == "" {
		*license = cfg.License
	}
	if !generator.ValidLicenseKind(*license) {
		logError("unknown --license %q — expected \"mit\" or \"apache-2.0\"", *license)
		os.Exit(1)
	}
	modulePath := *module
	if modulePath == "" && cfg.ModulePrefix != "" {
		modulePath = strings.TrimSuffix(cfg.ModulePrefix, "/") + "/" + *name
	}
	if modulePath == "" {
		modulePath = *name
	}

	deps, versions, err := resolveDeps(*depsFlag)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}
	presetDeps, err := resolvePresetDeps(*presetFlag)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}
	deps = mergeDeps(presetDeps, deps)

	req := generator.Requirement{
		ProjectName:           *name,
		PackageName:           modulePath,
		Deps:                  deps,
		IncludeDocker:         *docker,
		IncludeCI:             *ci,
		IncludeMakefile:       *makefile,
		IncludeGitInit:        *gitInit,
		IncludeReadme:         *readme,
		IncludeCommunityFiles: *community,
		License:               *license,
		DepVersions:           versions,
	}

	if *dryRun {
		req, err = generator.CheckPreconditions(req)
		if err != nil {
			logError("%v", err)
			os.Exit(1)
		}
		logInfo("[dry-run] would create ./%s/ (module %s)", req.ProjectName, req.PackageName)
		if err := runStepsPlain(generator.BuildInstallSteps(req.ProjectName, req), true); err != nil {
			logError("%v", err)
			os.Exit(1)
		}
		return
	}

	targetPath, err := generator.PrepareNewProject(req)
	if err != nil {
		logError("Failed to create project: %v", err)
		os.Exit(1)
	}

	if err := runStepsPlain(generator.BuildInstallSteps(targetPath, req), false); err != nil {
		if removeErr := os.RemoveAll(req.ProjectName); removeErr != nil {
			logWarn("Could not remove partial output %q: %v", req.ProjectName, removeErr)
		} else {
			logInfo("Cleaned up partial directory: ./%s/", req.ProjectName)
		}
		logError("Failed to create project: %v", err)
		os.Exit(1)
	}

	logSuccess("Project available at: ./%s", req.ProjectName)
	generator.PrintDocs(req.Deps)
	if req.IncludeGitInit {
		logInfo("📦 To publish: %s", generator.SuggestedPublishCommand())
	}
	if advisory := generator.VulnCheckAdvisory(targetPath); advisory != "" {
		logWarn("%s", advisory)
	}
	generator.RecordHistory("init", req.ProjectName, fmt.Sprintf("%d deps", len(req.Deps)))
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
		logError("No go.mod found in this directory. Run `go mod init <module>` first, or use `genitz init` to scaffold a new project.")
		os.Exit(1)
	}

	model := tui.AddModel(module)
	model.BuildSteps = func(m *tui.Model) []tui.InstallStep {
		return generator.BuildAddSteps(".", m.Chosen, nil)
	}

	finalModel := runProgram(model)

	if finalModel.InstallErr != nil {
		logError("Failed to add dependency: %v", finalModel.InstallErr)
		os.Exit(1)
	}

	if !finalModel.Done {
		logInfo("Cancelled.")
		return
	}

	logSuccess("Dependencies installed")
	generator.PrintDocs(finalModel.Chosen)
	generator.RecordHistory("add", module, fmt.Sprintf("%d deps", len(finalModel.Chosen)))
}

// runAddFlags installs dependencies directly from flags — no TUI.
func runAddFlags(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	depsFlag := fs.String("deps", "", "comma-separated dependency IDs (required unless --preset is set)")
	presetFlag := fs.String("preset", "", "starter bundle to apply (combines with --deps)")
	dryRun := fs.Bool("dry-run", false, "print steps without running them")
	fs.Parse(args)

	if *depsFlag == "" && *presetFlag == "" {
		logError("--deps or --preset is required for non-interactive add")
		os.Exit(1)
	}

	if _, err := generator.ReadModulePath("."); err != nil {
		logError("No go.mod found in this directory. Run `go mod init <module>` first, or use `genitz init` to scaffold a new project.")
		os.Exit(1)
	}

	deps, versions, err := resolveDeps(*depsFlag)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}
	presetDeps, err := resolvePresetDeps(*presetFlag)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}
	deps = mergeDeps(presetDeps, deps)

	if err := runStepsPlain(generator.BuildAddSteps(".", deps, versions), *dryRun); err != nil {
		logError("Failed to add dependency: %v", err)
		os.Exit(1)
	}
	if *dryRun {
		return
	}

	logSuccess("Dependencies installed")
	generator.PrintDocs(deps)
	if advisory := generator.VulnCheckAdvisory("."); advisory != "" {
		logWarn("%s", advisory)
	}
	generator.RecordHistory("add", ".", fmt.Sprintf("%d deps", len(deps)))
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
		logError("No go.mod found in this directory. Run `go mod init <module>` first, or use `genitz init` to scaffold a new project.")
		os.Exit(1)
	}

	installed, err := generator.ListInstalled(".")
	if err != nil {
		logError("Failed to read go.mod: %v", err)
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
		logInfo("No registry-managed dependencies to remove.")
		return
	}

	model := tui.RemoveModel(module, removable)
	model.BuildSteps = func(m *tui.Model) []tui.InstallStep {
		return generator.BuildRemoveSteps(".", m.Chosen)
	}

	finalModel := runProgram(model)

	if finalModel.InstallErr != nil {
		logError("Failed to remove dependency: %v", finalModel.InstallErr)
		os.Exit(1)
	}
	if !finalModel.Done {
		logInfo("Cancelled.")
		return
	}
	logSuccess("Dependencies removed")
	generator.RecordHistory("remove", module, fmt.Sprintf("%d deps", len(finalModel.Chosen)))
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
		logError("--deps is required for non-interactive remove")
		os.Exit(1)
	}

	if _, err := generator.ReadModulePath("."); err != nil {
		logError("No go.mod found in this directory. Run `go mod init <module>` first, or use `genitz init` to scaffold a new project.")
		os.Exit(1)
	}

	deps, _, err := resolveDeps(*depsFlag)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}

	if err := runStepsPlain(generator.BuildRemoveSteps(".", deps), *dryRun); err != nil {
		logError("Failed to remove dependency: %v", err)
		os.Exit(1)
	}
	if *dryRun {
		return
	}

	logSuccess("Dependencies removed")
	generator.RecordHistory("remove", ".", fmt.Sprintf("%d deps", len(deps)))
}

// runList prints the direct dependencies already in ./go.mod — as a table,
// or as JSON with --json for scripting/CI consumption.
func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "print as JSON")
	fs.Parse(args)

	if _, err := generator.ReadModulePath("."); err != nil {
		logError("No go.mod found in this directory.")
		os.Exit(1)
	}
	deps, err := generator.ListInstalled(".")
	if err != nil {
		logError("Failed to read go.mod: %v", err)
		os.Exit(1)
	}

	if *jsonOut {
		b, err := json.MarshalIndent(deps, "", "  ")
		if err != nil {
			logError("%v", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
		return
	}
	generator.PrintInstalled(deps)
}

// runCompletion prints a static shell completion script for the requested
// shell — hand-authored (no cobra in this codebase), covering every
// top-level subcommand so a new one added to main() should also be added
// to subcommandNames below (see main_test.go for the drift check).
func runCompletion(args []string) {
	if len(args) == 0 {
		logError("Usage: genitz completion <bash|zsh|fish>")
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
		logError("Unknown shell %q — expected bash, zsh, or fish", args[0])
		os.Exit(1)
	}
}

// subcommandNames are genitz's top-level subcommands, single source of
// truth for both the completion scripts below and main()'s dispatch.
var subcommandNames = []string{
	"init", "add", "remove", "list", "version", "completion",
	"doctor", "config", "search", "info", "preset", "history", "help",
}

// runConfig implements `genitz config` (show all), `genitz config get
// <key>`, and `genitz config set <key> <value>`.
func runConfig(args []string) {
	cfg, err := generator.LoadConfig()
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		for _, key := range generator.ConfigKeys {
			fmt.Printf("%-14s %s\n", key, cfg.Get(key))
		}
		return
	}

	switch args[0] {
	case "get":
		if len(args) < 2 {
			logError("Usage: genitz config get <key>")
			os.Exit(1)
		}
		fmt.Println(cfg.Get(args[1]))
	case "set":
		if len(args) < 3 {
			logError("Usage: genitz config set <key> <value>")
			os.Exit(1)
		}
		if err := cfg.Set(args[1], args[2]); err != nil {
			logError("%v", err)
			os.Exit(1)
		}
		if err := cfg.Save(); err != nil {
			logError("Failed to save config: %v", err)
			os.Exit(1)
		}
		logSuccess("%s = %q", args[1], args[2])
	default:
		logError("Unknown config subcommand %q — expected get or set", args[0])
		os.Exit(1)
	}
}

// runSearch prints registry entries whose ID/name/category/description
// contain the given query (case-insensitive) — an alternative to opening
// the full TUI just to check whether a package is registered.
func runSearch(args []string) {
	if len(args) == 0 {
		logError("Usage: genitz search <keyword>")
		os.Exit(1)
	}
	query := strings.ToLower(strings.Join(args, " "))

	var matches []tui.Dependency
	for _, dep := range tui.DependencyRegistry {
		if strings.Contains(strings.ToLower(dep.Name), query) ||
			strings.Contains(strings.ToLower(dep.ID), query) ||
			strings.Contains(strings.ToLower(dep.Category), query) ||
			strings.Contains(strings.ToLower(dep.Description), query) {
			matches = append(matches, dep)
		}
	}
	if len(matches) == 0 {
		logInfo("No matches for %q", strings.Join(args, " "))
		return
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Name < matches[j].Name })
	for _, dep := range matches {
		fmt.Printf("   %-20s %-14s %s\n", dep.ID, dep.Category, dep.Name)
	}
}

// runInfo prints the full registry entry for one dependency ID.
func runInfo(args []string) {
	if len(args) == 0 {
		logError("Usage: genitz info <dependency-id>")
		os.Exit(1)
	}
	dep, ok := tui.FindByID(args[0])
	if !ok {
		logError("Unknown dependency ID %q — try `genitz search %s`", args[0], args[0])
		os.Exit(1)
	}
	fmt.Printf("\n%s (%s)\n", dep.Name, dep.ID)
	fmt.Printf("  Category:    %s\n", dep.Category)
	fmt.Printf("  Import path: %s\n", dep.ImportPath)
	fmt.Printf("  Description: %s\n", dep.Description)
	fmt.Printf("  Docs:        https://pkg.go.dev/%s\n", dep.ImportPath)
}

// runPreset implements `genitz preset list` and `genitz preset save`.
func runPreset(args []string) {
	if len(args) == 0 {
		logError("Usage: genitz preset list | genitz preset save <id> --deps id1,id2 [--name ...] [--description ...]")
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		for _, p := range tui.AllPresets() {
			fmt.Printf("   %-16s %-24s %s\n", p.ID, p.Name, p.Description)
		}
	case "save":
		runPresetSave(args[1:])
	default:
		logError("Unknown preset subcommand %q — expected list or save", args[0])
		os.Exit(1)
	}
}

// runPresetSave saves the current selection (given via --deps) as a
// reusable preset — `genitz preset save <id> --deps a,b,c`.
func runPresetSave(args []string) {
	if len(args) == 0 {
		logError("Usage: genitz preset save <id> --deps id1,id2 [--name ...] [--description ...]")
		os.Exit(1)
	}
	id := args[0]

	fs := flag.NewFlagSet("preset save", flag.ExitOnError)
	depsFlag := fs.String("deps", "", "comma-separated dependency IDs (required)")
	name := fs.String("name", "", "human-readable preset name (defaults to the ID)")
	description := fs.String("description", "", "short description shown in the picker")
	fs.Parse(args[1:])

	if *depsFlag == "" {
		logError("--deps is required")
		os.Exit(1)
	}
	deps, _, err := resolveDeps(*depsFlag)
	if err != nil {
		logError("%v", err)
		os.Exit(1)
	}

	ids := make([]string, 0, len(deps))
	for _, d := range deps {
		ids = append(ids, d.ID)
	}
	sort.Strings(ids)

	humanName := *name
	if humanName == "" {
		humanName = id
	}
	preset := tui.Preset{ID: id, Name: humanName, Description: *description, DepIDs: ids}
	if err := tui.SavePreset(preset); err != nil {
		logError("Failed to save preset: %v", err)
		os.Exit(1)
	}
	logSuccess("Preset %q saved (%d dependencies)", id, len(ids))
}

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
			logInfo("[dry-run] would run: %s", step.Label)
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
		logError("%v", err)
		os.Exit(1)
	}
	return m.(*tui.Model)
}
