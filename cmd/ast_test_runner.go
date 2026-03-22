package main

import (
	"fmt"
	"log"

	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

func main() {
	req := generator.Requirement{
		ProjectName: "test-ast-project",
		PackageName: "github.com/koriebruh/test",
		Arch: tui.Architecture{
			Name:        tui.ArchMicro,
			TemplateDir: "templates/architecture/microservice",
		},
		Deps: map[int]tui.Dependency{
			0: tui.DependencyRegistry[2], // Gin
			1: tui.DependencyRegistry[0], // Gorm
		},
	}
	fmt.Println("Generating AST injection test project...")
	if err := generator.GenerateNewProject(req); err != nil {
		log.Fatalf("FAILED: %v", err)
	}
	fmt.Println("SUCCESS. Please check the injected files in ./test-ast-project")
}
