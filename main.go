package main

import (
	"fmt"
	"os"

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
			runAdd()
		} else {
			runInit()
		}
	case args[0] == "init":
		runInit()
	case args[0] == "add":
		runAdd()
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
  genitz init    Always start the new-project wizard.
  genitz add     Add dependencies to the project in the current directory
                 (requires an existing go.mod).
  genitz help    Show this message.`)
}

// runInit walks the new-project wizard. Confirming on Review animates the
// install (go mod / go get) inside the TUI itself via model.BuildSteps.
func runInit() {
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
}

// runAdd walks the dependency picker. Confirming on Review animates the
// install inside the TUI itself via model.BuildSteps.
func runAdd() {
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
