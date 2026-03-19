package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.InitialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	finalModel := m.(*tui.Model)

	req, err := generator.NewRequirementFromModel(finalModel)
	if err != nil {
		fmt.Printf("\nInput belum lengkap: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🛠️  Sedang memproses project...")
	if err := generator.GenerateNewProject(req); err != nil {
		fmt.Printf("\nGagal membuat project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n📂 Project tersedia di: ./%s\n", req.ProjectName)
}
