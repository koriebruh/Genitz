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
  genitz         Auto-detect: start a new project, or add dependencies if
                 go.mod already exists in the current directory.
  genitz init    Start the new-project wizard.
  genitz add     Add dependencies to the project in the current directory
                 (requires an existing go.mod).
  genitz help    Show this message.

Non-interactive (scripting/CI) — pass --name or --deps to skip the wizard:
  genitz init --name my-app --module github.com/me/my-app --deps fiber,redis \
              [--docker] [--ci] [--makefile] [--git]
  genitz add --deps redis,zap

  --name      folder name (required for non-interactive init)
  --module    Go module path (defaults to --name)
  --deps      comma-separated dependency IDs — see internal/tui/dependencies.go
  --docker    generate a multistage Dockerfile (+ docker-compose.yml if warranted)
  --ci        generate a GitHub Actions CI workflow
  --makefile  generate a Makefile
  --git       git init + first commit (init only)`)
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
	docker := fs.Bool("docker", false, "generate Docker setup")
	ci := fs.Bool("ci", false, "generate GitHub Actions CI workflow")
	makefile := fs.Bool("makefile", false, "generate Makefile")
	gitInit := fs.Bool("git", false, "git init + first commit")
	fs.Parse(args)

	if *name == "" {
		fmt.Println("\n--name is required for non-interactive init")
		os.Exit(1)
	}
	modulePath := *module
	if modulePath == "" {
		modulePath = *name
	}

	deps, err := resolveDeps(*depsFlag)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}

	req := generator.Requirement{
		ProjectName:     *name,
		PackageName:     modulePath,
		Deps:            deps,
		IncludeDocker:   *docker,
		IncludeCI:       *ci,
		IncludeMakefile: *makefile,
		IncludeGitInit:  *gitInit,
	}

	targetPath, err := generator.PrepareNewProject(req)
	if err != nil {
		fmt.Printf("\nGagal membuat project: %v\n", err)
		os.Exit(1)
	}

	if err := runStepsPlain(generator.BuildInstallSteps(targetPath, req)); err != nil {
		if removeErr := os.RemoveAll(req.ProjectName); removeErr == nil {
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
		return generator.BuildAddSteps(".", m.Chosen)
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
	depsFlag := fs.String("deps", "", "comma-separated dependency IDs (required)")
	fs.Parse(args)

	if *depsFlag == "" {
		fmt.Println("\n--deps is required for non-interactive add")
		os.Exit(1)
	}

	if _, err := generator.ReadModulePath("."); err != nil {
		fmt.Println("\nGak ada go.mod di direktori ini. Jalanin `go mod init <module>` dulu, atau pakai `genitz init` buat project baru.")
		os.Exit(1)
	}

	deps, err := resolveDeps(*depsFlag)
	if err != nil {
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}

	if err := runStepsPlain(generator.BuildAddSteps(".", deps)); err != nil {
		fmt.Printf("\nGagal menambahkan dependency: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Dependencies installed ✨")
	generator.PrintDocs(deps)
}

// resolveDeps turns a comma-separated list of registry IDs into a Chosen-shaped
// map, failing fast with a clear message on the first unknown ID.
func resolveDeps(depsFlag string) (map[int]tui.Dependency, error) {
	deps := make(map[int]tui.Dependency)
	depsFlag = strings.TrimSpace(depsFlag)
	if depsFlag == "" {
		return deps, nil
	}
	for i, id := range strings.Split(depsFlag, ",") {
		id = strings.TrimSpace(id)
		dep, ok := tui.FindByID(id)
		if !ok {
			return nil, fmt.Errorf("unknown dependency ID %q — see internal/tui/dependencies.go for valid IDs", id)
		}
		deps[i] = dep
	}
	return deps, nil
}

// runStepsPlain runs InstallSteps sequentially with plain progress lines —
// no Bubble Tea, no TTY assumptions, safe for CI/scripted output.
func runStepsPlain(steps []tui.InstallStep) error {
	for _, step := range steps {
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
