package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

func main() {
	// WithAltScreen renders into the terminal's alternate buffer — this prevents
	// the "double logo" effect when the user zooms in/out in their terminal.
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
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
