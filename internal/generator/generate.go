package generator

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/koriebruh/Genitz/internal/tui"
)

// Requirement describes everything needed to scaffold a new project.
type Requirement struct {
	ProjectName string
	PackageName string
	Deps        map[int]tui.Dependency
}

// NewRequirementFromModel converts the interactive model into a concrete Requirement.
func NewRequirementFromModel(m *tui.Model) (Requirement, error) {
	if m == nil {
		return Requirement{}, errors.New("model is nil")
	}

	projectName := strings.TrimSpace(m.FolderInput.Value())
	if projectName == "" {
		return Requirement{}, errors.New("project/folder name cannot be empty")
	}

	packageName := strings.TrimSpace(m.PkgInput.Value())
	if packageName == "" {
		packageName = projectName
	}

	deps := make(map[int]tui.Dependency, len(m.Chosen))
	for idx, dep := range m.Chosen {
		deps[idx] = dep
	}

	return Requirement{
		ProjectName: projectName,
		PackageName: packageName,
		Deps:        deps,
	}, nil
}

const bareMainGo = "package main\n\nfunc main() {}\n"

// GenerateNewProject scaffolds a bare new Go project from the provided requirement.
func GenerateNewProject(req Requirement) error {
	if err := req.validate(); err != nil {
		return err
	}

	targetPath, err := filepath.Abs(req.ProjectName)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	if err := ensureFreshProjectDir(targetPath); err != nil {
		return err
	}

	fmt.Printf("\n📁 Creating project at %s\n", targetPath)
	if err := os.WriteFile(filepath.Join(targetPath, "main.go"), []byte(bareMainGo), 0o644); err != nil {
		return fmt.Errorf("write main.go: %w", err)
	}

	if err := initGoModule(targetPath, req.PackageName); err != nil {
		return err
	}

	if err := installDependencies(targetPath, req.Deps); err != nil {
		return err
	}

	fmt.Println("✨ Finalizing project...")
	if err := runIn(targetPath, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("final tidy: %w", err)
	}
	if err := runIn(targetPath, "go", "fmt", "./..."); err != nil {
		return fmt.Errorf("final fmt: %w", err)
	}

	fmt.Println("\n✅ Project scaffold complete! Happy hacking ✨")
	return nil
}

// AddDependencies installs the chosen dependencies into the Go module found
// at targetDir (an existing project — go.mod already present).
func AddDependencies(targetDir string, deps map[int]tui.Dependency) error {
	if err := installDependencies(targetDir, deps); err != nil {
		return err
	}
	fmt.Println("✨ Tidying go.mod...")
	if err := runIn(targetDir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	fmt.Println("\n✅ Dependencies installed ✨")
	return nil
}

// ReadModulePath reads the module directive from go.mod in dir.
func ReadModulePath(dir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", errors.New("module directive not found in go.mod")
}

func (r *Requirement) validate() error {
	r.ProjectName = strings.TrimSpace(r.ProjectName)
	if r.ProjectName == "" {
		return errors.New("project name is required")
	}

	r.PackageName = strings.TrimSpace(r.PackageName)
	if r.PackageName == "" {
		r.PackageName = r.ProjectName
	}

	return nil
}

func ensureFreshProjectDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("destination %q already exists", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check destination: %w", err)
	}
	return os.MkdirAll(path, 0o755)
}

func initGoModule(target, module string) error {
	fmt.Printf("⚙️  Initialising go.mod (%s)\n", module)
	if err := runIn(target, "go", "mod", "init", module); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}
	if err := runIn(target, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func installDependencies(target string, deps map[int]tui.Dependency) error {
	if len(deps) == 0 {
		return nil
	}

	fmt.Println("📚 Installing selected dependencies")
	seen := make(map[string]struct{})
	var toInstall []string
	for _, dep := range deps {
		if dep.ImportPath == "" {
			continue
		}
		if _, ok := seen[dep.ImportPath]; ok {
			continue
		}
		seen[dep.ImportPath] = struct{}{}
		toInstall = append(toInstall, dep.ImportPath)
	}

	sort.Strings(toInstall)
	for _, importPath := range toInstall {
		if err := runIn(target, "go", "get", importPath); err != nil {
			return fmt.Errorf("go get %s: %w", importPath, err)
		}
	}

	return nil
}

func runIn(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
