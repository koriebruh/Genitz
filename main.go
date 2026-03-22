package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

func main() {
	if len(os.Args) >= 2 {
		if os.Args[1] == "add" {
			if len(os.Args) < 3 {
				fmt.Println("Usage: genitz add <package_id>")
				fmt.Println("Example: genitz add redis\nAvailable packages: fiber, gin, gorm, redis, zap, validator")
				os.Exit(1)
			}
			pkgID := os.Args[2]
			if err := generator.AddDependencyHeadless(pkgID); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		} else if os.Args[1] == "clone" {
			if len(os.Args) < 4 {
				fmt.Println("Usage: genitz clone <repo_url> <project_name>")
				fmt.Println("Example: genitz clone https://github.com/koriebruh/my-base-go new-app")
				os.Exit(1)
			}
			repoURL := os.Args[2]
			projName := os.Args[3]
			if err := generator.CloneRemoteTemplate(repoURL, projName); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		} else {
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Available commands: add, clone")
			os.Exit(1)
		}
	}
	genFunc := func(m *tui.Model) error {
		req, err := generator.NewRequirementFromModel(m)
		if err != nil {
			return err
		}
		err = generator.GenerateNewProject(req)
		if err != nil {
			// Silently clean up partial directory if scaffolding fails
			_ = os.RemoveAll(req.ProjectName)
			return err
		}
		return nil
	}

	// WithAltScreen renders into the terminal's alternate buffer — this prevents
	// the "double logo" effect when the user zooms in/out in their terminal.
	p := tea.NewProgram(tui.InitialModel(genFunc), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	finalModel := m.(*tui.Model)

	// User pressed q/ctrl+c without confirming on the Review screen — abort.
	if !finalModel.Done {
		fmt.Println("\nCancelled.")
		return
	}

	if finalModel.GenErr == nil {
		fmt.Printf("\n📂 Project tersedia di: ./%s\n", finalModel.FolderInput.Value())
	}
}
