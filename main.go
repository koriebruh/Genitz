package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

func main() {
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
