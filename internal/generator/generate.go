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
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/koriebruh/Genitz/internal/tui"
	"github.com/mattn/go-isatty"
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

	if err := runWithSpinner(targetPath, "Tidying go.mod", "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("final tidy: %w", err)
	}
	if err := runWithSpinner(targetPath, "Formatting code", "go", "fmt", "./..."); err != nil {
		return fmt.Errorf("final fmt: %w", err)
	}

	fmt.Println("\n✅ Project scaffold complete! Happy hacking ✨")
	printDocs(req.Deps)
	return nil
}

// AddDependencies installs the chosen dependencies into the Go module found
// at targetDir (an existing project — go.mod already present).
func AddDependencies(targetDir string, deps map[int]tui.Dependency) error {
	if err := installDependencies(targetDir, deps); err != nil {
		return err
	}
	if err := runWithSpinner(targetDir, "Tidying go.mod", "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	fmt.Println("\n✅ Dependencies installed ✨")
	printDocs(deps)
	return nil
}

// printDocs lists what got installed and where to read up on each one —
// pkg.go.dev renders a module's README plus its API docs for any import
// path, so it's a correct reference without hand-curating a link per dep.
func printDocs(deps map[int]tui.Dependency) {
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

func initGoModule(target, module string) error {
	if err := runWithSpinner(target, fmt.Sprintf("Initialising go.mod (%s)", module), "go", "mod", "init", module); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}
	if err := runWithSpinner(target, "Tidying go.mod", "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func installDependencies(target string, deps map[int]tui.Dependency) error {
	if len(deps) == 0 {
		return nil
	}

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
		label := fmt.Sprintf("Installing %s", importPath)
		if err := runWithSpinner(target, label, "go", "get", importPath); err != nil {
			return fmt.Errorf("go get %s: %w", importPath, err)
		}
	}

	return nil
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#22D3EE")).Bold(true)
	doneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
)

// runWithSpinner runs a command with its output captured (not streamed) and
// an animated spinner shown next to label while it's in flight. On success
// the line collapses to a checkmark; on failure it collapses to a cross and
// the captured output is dumped so the underlying error is still visible.
// Falls back to a single plain line (no animation) when stdout isn't a TTY.
func runWithSpinner(dir, label, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		err := <-done
		if err != nil {
			fmt.Printf("✖ %s\n%s\n", label, out.String())
			return err
		}
		fmt.Printf("✔ %s\n", label)
		return nil
	}

	ticker := time.NewTicker(90 * time.Millisecond)
	defer ticker.Stop()
	frame := 0
	for {
		select {
		case err := <-done:
			fmt.Print("\r\033[K")
			if err != nil {
				fmt.Printf("%s %s\n%s\n", failStyle.Render("✖"), label, out.String())
				return err
			}
			fmt.Printf("%s %s\n", doneStyle.Render("✔"), label)
			return nil
		case <-ticker.C:
			fmt.Printf("\r\033[K%s %s", spinnerStyle.Render(spinnerFrames[frame%len(spinnerFrames)]), label)
			frame++
		}
	}
}
