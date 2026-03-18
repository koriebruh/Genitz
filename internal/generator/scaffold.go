package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CreateProject(projectName string, framework string) error {
	// Bikin Folder
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return err
	}

	// Isi file main.go sederhana
	content := fmt.Sprintf(`package main

import "fmt"

func main() {
    fmt.Println("Project: %s")
    fmt.Println("Framework: %s")
    fmt.Println("Status: Ready to Gas!")
}
`, projectName, framework)

	filePath := filepath.Join(projectName, "main.go")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func InitGoMod(projectName string) {
	fmt.Printf("\n📦 Menjalankan 'go mod init' di ./%s...\n", projectName)
	cmd := exec.Command("go", "mod", "init", projectName)
	cmd.Dir = projectName
	cmd.Run()
}
