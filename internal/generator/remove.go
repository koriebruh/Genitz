package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/koriebruh/Genitz/internal/astparser"
	"github.com/koriebruh/Genitz/internal/registry"
)

// RemoveDependencyHeadless is the engine for `genitz remove <pkg>`.
// It removes the AST import and matching init statements from main.go,
// then runs go mod tidy to clean go.sum.
func RemoveDependencyHeadless(pkg string) error {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found — run this command from your Go project root")
	}

	dep, err := registry.GetDependencyByID(pkg)
	if err != nil {
		return err
	}

	fmt.Printf("🗑️  Removing %s from your project...\n", dep.Name)

	mainFilePath, err := findMainGoFilePath(".")
	if err != nil {
		return fmt.Errorf("could not locate main.go: %w", err)
	}

	// 1. Remove imports defined in inject.json
	if dep.TemplateDir != "" {
		injectFile := dep.TemplateDir + "/inject.json"
		content, readErr := templatesFS.ReadFile(injectFile)
		if readErr == nil {
			var inject FeatureInjection
			if jsonErr := json.Unmarshal(content, &inject); jsonErr == nil {
				for _, imp := range inject.MainImports {
					if err := astparser.RemoveImport(mainFilePath, imp); err != nil {
						fmt.Printf("⚠️  Warn: could not remove import %s: %v\n", imp, err)
					} else {
						fmt.Printf("✅ Removed import: %s\n", imp)
					}
				}
				// Remove init statements by keyword (pkg base name)
				if err := astparser.RemoveStatementsMatching(mainFilePath, "main", filepath.Base(dep.ImportPath)); err != nil {
					fmt.Printf("⚠️  Warn: could not remove init code: %v\n", err)
				} else {
					fmt.Printf("✅ Removed init code for %s\n", dep.Name)
				}
			}
		}
	}

	// 2. Tidy
	fmt.Println("🧹 Running go mod tidy...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	tidyCmd.Run()

	fmt.Printf("✅ Successfully removed %s!\n", dep.Name)
	return nil
}
