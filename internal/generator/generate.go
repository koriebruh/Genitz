package generator

import (
	"bytes"
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
	ProjectName   string
	PackageName   string
	Deps          map[int]tui.Dependency
	IncludeDocker bool
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
		ProjectName:   projectName,
		PackageName:   packageName,
		Deps:          deps,
		IncludeDocker: m.IncludeDocker,
	}, nil
}

const bareMainGo = "package main\n\nfunc main() {}\n"

// PrepareNewProject validates req, creates the target directory, and writes
// a bare main.go. It's synchronous and network-free — the animated part
// (go mod / go get) is BuildInstallSteps, run separately so the TUI can
// show progress for it.
func PrepareNewProject(req Requirement) (targetPath string, err error) {
	if err := req.validate(); err != nil {
		return "", err
	}

	targetPath, err = filepath.Abs(req.ProjectName)
	if err != nil {
		return "", fmt.Errorf("resolve project path: %w", err)
	}

	if err := ensureFreshProjectDir(targetPath); err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath.Join(targetPath, "main.go"), []byte(bareMainGo), 0o644); err != nil {
		return "", fmt.Errorf("write main.go: %w", err)
	}

	return targetPath, nil
}

// BuildInstallSteps returns the ordered go mod/go get/go fmt steps for
// scaffolding req at targetPath — each one runnable independently so the
// caller (the TUI) can animate progress as they execute.
func BuildInstallSteps(targetPath string, req Requirement) []tui.InstallStep {
	steps := []tui.InstallStep{
		{
			Label: fmt.Sprintf("Initialising go.mod (%s)", req.PackageName),
			Run:   func() error { return runCaptured(targetPath, "go", "mod", "init", req.PackageName) },
		},
		{Label: "Tidying go.mod", Run: func() error { return runCaptured(targetPath, "go", "mod", "tidy") }},
	}

	if req.IncludeDocker {
		steps = append(steps, tui.InstallStep{
			Label: "Generating Dockerfile",
			Run: func() error {
				goVersion := readGoVersion(targetPath)
				if err := os.WriteFile(filepath.Join(targetPath, "Dockerfile"), []byte(dockerfileContent(goVersion)), 0o644); err != nil {
					return fmt.Errorf("write Dockerfile: %w", err)
				}
				if err := os.WriteFile(filepath.Join(targetPath, ".dockerignore"), []byte(dockerignoreContent()), 0o644); err != nil {
					return fmt.Errorf("write .dockerignore: %w", err)
				}
				return nil
			},
		})

		if content, ok := composeContent(req); ok {
			steps = append(steps, tui.InstallStep{
				Label: "Generating docker-compose.yml",
				Run: func() error {
					if err := os.WriteFile(filepath.Join(targetPath, "docker-compose.yml"), []byte(content), 0o644); err != nil {
						return fmt.Errorf("write docker-compose.yml: %w", err)
					}
					return nil
				},
			})
		}
	}

	for _, importPath := range sortedImportPaths(req.Deps) {
		ip := importPath
		steps = append(steps, tui.InstallStep{
			Label: "Installing " + ip,
			Run:   func() error { return runCaptured(targetPath, "go", "get", ip) },
		})
	}
	steps = append(steps,
		tui.InstallStep{Label: "Tidying go.mod", Run: func() error { return runCaptured(targetPath, "go", "mod", "tidy") }},
		tui.InstallStep{Label: "Formatting code", Run: func() error { return runCaptured(targetPath, "go", "fmt", "./...") }},
	)
	return steps
}

// BuildAddSteps returns the ordered go get/go mod tidy steps for installing
// deps into the existing project at targetDir.
func BuildAddSteps(targetDir string, deps map[int]tui.Dependency) []tui.InstallStep {
	var steps []tui.InstallStep
	for _, importPath := range sortedImportPaths(deps) {
		ip := importPath
		steps = append(steps, tui.InstallStep{
			Label: "Installing " + ip,
			Run:   func() error { return runCaptured(targetDir, "go", "get", ip) },
		})
	}
	steps = append(steps, tui.InstallStep{
		Label: "Tidying go.mod",
		Run:   func() error { return runCaptured(targetDir, "go", "mod", "tidy") },
	})
	return steps
}

// PrintDocs lists what got installed and where to read up on each one —
// pkg.go.dev renders a module's README plus its API docs for any import
// path, so it's a correct reference without hand-curating a link per dep.
func PrintDocs(deps map[int]tui.Dependency) {
	if len(deps) == 0 {
		return
	}

	names := make([]string, 0, len(deps))
	byName := make(map[string]tui.Dependency, len(deps))
	for _, dep := range deps {
		names = append(names, dep.Name)
		byName[dep.Name] = dep
	}
	sort.Strings(names)

	fmt.Println("\n📚 Documentation:")
	for _, name := range names {
		dep := byName[name]
		fmt.Printf("   %-28s %s\n", dep.Name, "https://pkg.go.dev/"+dep.ImportPath)
	}
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

// sortedImportPaths dedupes and sorts the import paths of deps.
func sortedImportPaths(deps map[int]tui.Dependency) []string {
	seen := make(map[string]struct{}, len(deps))
	var paths []string
	for _, dep := range deps {
		if dep.ImportPath == "" {
			continue
		}
		if _, ok := seen[dep.ImportPath]; ok {
			continue
		}
		seen[dep.ImportPath] = struct{}{}
		paths = append(paths, dep.ImportPath)
	}
	sort.Strings(paths)
	return paths
}

// runCaptured runs a command with its output captured instead of streamed —
// the TUI owns the visible progress animation, so a failure folds the
// captured output into the returned error instead of printing it directly.
func runCaptured(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w\n%s", err, strings.TrimSpace(out.String()))
	}
	return nil
}
