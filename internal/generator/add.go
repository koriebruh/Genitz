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

// AddDependencyHeadless is the core logic for `genitz add <pkg>`.
// It scans the current directory for a Go project, locates main.go,
// injects the requested dependency via AST, and runs go get.
func AddDependencyHeadless(pkg string) error {
	// 1. Verify we are inside a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Please run 'genitz add' from the root of a Go project")
	}

	// 2. Resolve Dependency
	dep, err := registry.GetDependencyByID(pkg)
	if err != nil {
		return err
	}

	fmt.Printf("📦 Adding %s (%s) to your project layout...\n", dep.Name, dep.ImportPath)

	// 3. Find main.go
	mainFilePath, err := findMainGoFilePath(".")
	if err != nil {
		return fmt.Errorf("could not find main.go to inject into (are you inside the project root?): %w", err)
	}

	// 4. Run AST Injections & Templating
	if dep.TemplateDir != "" {
		injectFile := dep.TemplateDir + "/inject.json"
		content, err := templatesFS.ReadFile(injectFile)
		if err == nil {
			var inject FeatureInjection
			if err := json.Unmarshal(content, &inject); err != nil {
				return fmt.Errorf("failed to parse inject configuration for %s: %w", dep.Name, err)
			}

			// Injects
			for _, imp := range inject.MainImports {
				if err := astparser.AddImport(mainFilePath, imp); err != nil {
					fmt.Printf("⚠️ Warn: failed to inject import %s: %v\n", imp, err)
				}
			}

			if inject.MainInit != "" {
				if err := astparser.InjectToMain(mainFilePath, inject.MainInit); err != nil {
					fmt.Printf("⚠️ Warn: failed to inject initialization code: %v\n", err)
				}
			}

			// Struct Injections
			for _, si := range inject.StructInject {
				structFilePath := filepath.Join(".", si.FilePath)
				if err := astparser.InjectStructField(structFilePath, si.StructName, si.Field); err != nil {
					fmt.Printf("⚠️ Warn: failed to inject struct field into %s: %v\n", si.FilePath, err)
				} else {
					fmt.Printf("✅ Debug: Successfully injected struct field %q into %s\n", si.Field, si.StructName)
				}
			}

			// Modifikasi: Buat auto-test mock file untuk paket yang diretas
			goPkgName := filepath.Base(dep.ImportPath)
			testScaffold(".", goPkgName)

			// Scaffold the boilerplate files for the feature
			if inject.TargetDir != "" {
				// We pass nil data because we don't have a Requirement object here,
				// avoiding any text/template rendering logic since we strictly use AST now.
				err = copyDirFromEmbed(dep.TemplateDir, inject.TargetDir, map[string]any{})
				if err != nil {
					fmt.Printf("⚠️ Warn: failed to scaffold feature files: %v\n", err)
				}
			}
		}
	}

	// 5. Install the package
	fmt.Printf("⬇️  Downloading %s...\n", dep.ImportPath)
	cmd := exec.Command("go", "get", dep.ImportPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go get: %w", err)
	}

	// 6. Tidy and Format
	tCmd := exec.Command("go", "mod", "tidy")
	tCmd.Run()
	
	fCmd := exec.Command("go", "fmt", "./...")
	fCmd.Run()

	fmt.Printf("✅ Successfully added %s!\n", dep.Name)
	return nil
}
