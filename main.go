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

// runInit walks the new-project wizard and scaffolds it.
func runInit() {
	finalModel := runProgram(tui.InitialModel())
	if !finalModel.Done {
		fmt.Println("\nCancelled.")
		return
	}

	req, err := generator.NewRequirementFromModel(finalModel)
	if err != nil {
		fmt.Printf("\nInput belum lengkap: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🛠️  Sedang memproses project...")
	if err := generator.GenerateNewProject(req); err != nil {
		// Clean up any partially created directory so a failed run leaves no trace.
		if removeErr := os.RemoveAll(req.ProjectName); removeErr != nil {
			fmt.Printf("Warning: could not remove partial output %q: %v\n", req.ProjectName, removeErr)
		} else {
			fmt.Printf("Cleaned up partial directory: ./%s/\n", req.ProjectName)
		}
		fmt.Printf("\nGagal membuat project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n📂 Project tersedia di: ./%s\n", req.ProjectName)
}

// runAdd walks the dependency picker and installs the chosen packages into
// the Go module found in the current directory.
func runAdd() {
	module, err := generator.ReadModulePath(".")
	if err != nil {
		fmt.Println("\nGak ada go.mod di direktori ini. Jalanin `go mod init <module>` dulu, atau pakai `genitz init` buat project baru.")
		os.Exit(1)
	}

	finalModel := runProgram(tui.AddModel(module))
	if !finalModel.Done {
		fmt.Println("\nCancelled.")
		return
	}

	if err := generator.AddDependencies(".", finalModel.Chosen); err != nil {
		fmt.Printf("\nGagal menambahkan dependency: %v\n", err)
		os.Exit(1)
	}
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
