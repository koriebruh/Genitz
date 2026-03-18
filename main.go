package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

func main() {
	// Jalankan TUI
	p := tea.NewProgram(tui.InitialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	// Casting ke model TUI kita
	finalModel := m.(tui.Model)

	// Cek apakah ada yang dipilih
	if len(finalModel.Chosen) > 0 {
		projectName := "my-new-app"

		fmt.Println("\n🛠️  Sedang memproses project...")

		// Loop semua yang dicentang user
		for index := range finalModel.Chosen {
			dep := finalModel.Registry[index]

			fmt.Printf("   > Installing %s (%s)...\n", dep.Name, dep.Category)

			// Panggil generator untuk tiap dep
			err := generator.CreateProject(projectName, dep.Name)
			if err != nil {
				fmt.Printf("Gagal buat file untuk %s: %v\n", dep.Name, err)
			}
		}

		// Inisialisasi go mod sekali saja di akhir
		generator.InitGoMod(projectName)

		fmt.Printf("\n✅ SELESAI! Project lu ada di folder: ./%s\n", projectName)
	} else {
		fmt.Println("\nBlom ada yang dipilih, project nggak dibuat.")
	}
}
