package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/koriebruh/Genitz/internal/tui"
)

// templateRoot is the path prefix inside templatesFS.
const templateRoot = "templates"

var architectureCatalog = map[string]tui.Architecture{
	tui.ArchMicro: {
		Name:        tui.ArchMicro,
		Description: "Service-per-domain layout with shared pkg folder",
		TemplateDir: templateRoot + "/architecture/microservice",
	},
	tui.ArchClean: {
		Name:        tui.ArchClean,
		Description: "Layered clean architecture with adapters/core split",
		TemplateDir: templateRoot + "/architecture/clean-architecture",
	},
}

// Requirement describes everything needed to scaffold a project.
type Requirement struct {
	ProjectName string
	PackageName string
	Arch        tui.Architecture
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

	arch, err := resolveArchitecture(m.SelectedArch)
	if err != nil {
		return Requirement{}, err
	}

	deps := make(map[int]tui.Dependency, len(m.Chosen))
	for idx, dep := range m.Chosen {
		if dep.Name == "" && idx < len(m.Registry) {
			dep = m.Registry[idx]
		}
		deps[idx] = dep
	}

	return Requirement{
		ProjectName: projectName,
		PackageName: packageName,
		Arch:        arch,
		Deps:        deps,
	}, nil
}

// GenerateNewProject scaffolds a new Go project from the provided requirement.
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

	tmplData := newTemplateData(req)
	// Prepare config data based on dependencies
	enrichConfigData(req, tmplData)

	fmt.Printf("\n📁 Creating project at %s\n", targetPath)
	if err := copyArchitectureTemplate(req, targetPath, tmplData); err != nil {
		return err
	}

	if err := generateConfigFile(targetPath, tmplData); err != nil {
		return err
	}

	if err := applyDependencyTemplates(targetPath, req, tmplData); err != nil {
		return err
	}

	if err := generateProjectFiles(targetPath, req); err != nil {
		return err
	}

	if err := initGoModule(targetPath, req.PackageName); err != nil {
		return err
	}

	if err := installDependencies(targetPath, req.Deps); err != nil {
		return err
	}

	if err := mergeEnvFiles(targetPath, req); err != nil {
		return err
	}

	// Final cleanup: tidy first, then add mandatory env, then fmt
	fmt.Println("✨ Finalizing project...")

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = targetPath
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("final tidy: %w", err)
	}

	// Install mandatory config dependency
	fmt.Println("📦 Installing config loader...")
	envCmd := exec.Command("go", "get", "github.com/caarlos0/env/v11")
	envCmd.Dir = targetPath
	envCmd.Stdout = os.Stdout
	envCmd.Stderr = os.Stderr
	if err := envCmd.Run(); err != nil {
		return fmt.Errorf("install env dependency: %w", err)
	}

	fmtCmd := exec.Command("go", "fmt", "./...")
	fmtCmd.Dir = targetPath
	fmtCmd.Stdout = os.Stdout
	fmtCmd.Stderr = os.Stderr
	if err := fmtCmd.Run(); err != nil {
		return fmt.Errorf("final fmt: %w", err)
	}

	fmt.Println("\n✅ Project scaffold complete! Happy hacking ✨")
	return nil
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

	if r.Arch.TemplateDir == "" {
		arch, err := resolveArchitecture(r.Arch.Name)
		if err != nil {
			return err
		}
		r.Arch = arch
	}

	return nil
}

func resolveArchitecture(name string) (tui.Architecture, error) {
	arch, ok := architectureCatalog[name]
	if !ok {
		return tui.Architecture{}, fmt.Errorf("architecture %q is not supported yet", name)
	}
	return arch, nil
}

// ArchitectureAvailable reports whether an architecture name has a template.
func ArchitectureAvailable(name string) bool {
	_, ok := architectureCatalog[name]
	return ok
}

type FeatureInjection struct {
	TargetDir   string `json:"target_dir"`   // e.g. "config"
	ConfigField string `json:"config_field"` // e.g. "Redis RedisConfig"
	ConfigInit  string `json:"config_init"`  // e.g. "Redis: LoadRedisConfig(),"
}

// titleCase capitalises only the first letter of a string.
// Replaces the deprecated strings.Title.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}

func enrichConfigData(req Requirement, data map[string]any) {
	configFields := []string{}
	configInit := []string{}

	for _, dep := range req.Deps {
		if dep.TemplateDir == "" {
			continue
		}

		// Read inject.json from embedded FS
		injectFile := dep.TemplateDir + "/inject.json"
		content, err := templatesFS.ReadFile(injectFile)
		if err != nil {
			continue
		}

		var inject FeatureInjection
		if err := json.Unmarshal(content, &inject); err != nil {
			fmt.Printf("⚠️ Warn: failed to parse inject.json for %s: %v\n", dep.Name, err)
			continue
		}

		if inject.ConfigField != "" {
			configFields = append(configFields, "\t"+inject.ConfigField)
		}
		if inject.ConfigInit != "" {
			configInit = append(configInit, "\t\t"+inject.ConfigInit)
		}
	}

	data["ConfigFields"] = configFields
	data["ConfigInit"] = configInit
}

func generateConfigFile(target string, data map[string]any) error {
	src := templateRoot + "/feature/config.go.templete"
	dest := filepath.Join(target, "config", "config.go")
	fmt.Printf("🔧 Generating config: %s\n", dest)
	return renderTemplateFromEmbed(src, dest, data)
}

func ensureFreshProjectDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("destination %q already exists", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check destination: %w", err)
	}
	return os.MkdirAll(path, 0o755)
}

func copyArchitectureTemplate(req Requirement, target string, data map[string]any) error {
	src := req.Arch.TemplateDir
	fmt.Printf("🏗️  Applying architecture template: %s\n", req.Arch.Name)
	return copyDirFromEmbed(src, target, data)
}

func applyDependencyTemplates(target string, req Requirement, data map[string]any) error {
	for _, dep := range req.Deps {
		if dep.TemplateDir == "" {
			continue
		}

		// Determine destination from inject.json if available
		relDest := "internal/features/" + dep.ID
		injectFile := dep.TemplateDir + "/inject.json"
		if content, err := templatesFS.ReadFile(injectFile); err == nil {
			var inject FeatureInjection
			if json.Unmarshal(content, &inject) == nil && inject.TargetDir != "" {
				relDest = inject.TargetDir
			}
		}

		dest := filepath.Join(target, relDest)
		fmt.Printf("📦 Injecting feature template: %s -> %s\n", dep.Name, dest)
		if err := copyDirFromEmbed(dep.TemplateDir, dest, data); err != nil {
			return fmt.Errorf("copy dependency template %s: %w", dep.Name, err)
		}
	}
	return nil
}

// generateProjectFiles writes standard project files (README, .gitignore, Makefile)
// that every scaffold should include.
func generateProjectFiles(target string, req Requirement) error {
	fmt.Println("📝 Generating project files (README, .gitignore, Makefile)...")

	data := newTemplateData(req)

	files := []struct {
		tmplPath string
		destPath string
	}{
		{templateRoot + "/project/README.md.templete", filepath.Join(target, "README.md")},
		{templateRoot + "/project/.gitignore", filepath.Join(target, ".gitignore")},
		{templateRoot + "/project/Makefile.templete", filepath.Join(target, "Makefile")},
		{templateRoot + "/project/.air.toml", filepath.Join(target, ".air.toml")},
	}

	for _, f := range files {
		content, err := templatesFS.ReadFile(f.tmplPath)
		if err != nil {
			// skip optional files that don't exist yet
			continue
		}
		if strings.HasSuffix(f.tmplPath, ".templete") {
			if err := renderTemplateFromEmbed(f.tmplPath, f.destPath, data); err != nil {
				return fmt.Errorf("generate %s: %w", filepath.Base(f.destPath), err)
			}
		} else {
			if err := os.WriteFile(f.destPath, content, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", filepath.Base(f.destPath), err)
			}
		}
	}
	return nil
}

func initGoModule(target, module string) error {
	fmt.Printf("⚙️  Initialising go.mod (%s)\n", module)
	cmd := exec.Command("go", "mod", "init", module)
	cmd.Dir = target
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = target
	tidy.Stdout = os.Stdout
	tidy.Stderr = os.Stderr
	if err := tidy.Run(); err != nil {
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
		cmd := exec.Command("go", "get", importPath)
		cmd.Dir = target
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go get %s: %w", importPath, err)
		}
	}

	return nil
}

func mergeEnvFiles(target string, req Requirement) error {
	envPath := filepath.Join(target, ".env")
	envContent := fmt.Sprintf("APP_NAME=%s\nAPP_ENV=local\n", req.ProjectName)

	fmt.Println("📄 Merging environment variables...")

	seenKeys := make(map[string]bool)
	seenKeys["APP_NAME"] = true
	seenKeys["APP_ENV"] = true

	for _, dep := range req.Deps {
		if dep.TemplateDir == "" {
			continue
		}

		srcEnv := dep.TemplateDir + "/.env"
		content, err := templatesFS.ReadFile(srcEnv)
		if err != nil {
			continue // Skip if no .env in embedded FS
		}

		envContent += fmt.Sprintf("\n# %s Configuration\n", titleCase(dep.Name))

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				if !seenKeys[key] {
					envContent += line + "\n"
					seenKeys[key] = true
				}
			}
		}
	}

	return os.WriteFile(envPath, []byte(envContent), 0644)
}

// copyDirFromEmbed copies a directory from the embedded FS to the filesystem.
func copyDirFromEmbed(src, dest string, data map[string]any) error {
	return fs.WalkDir(templatesFS, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip inject.json and .env files from being copied to final project
		if d.Name() == "inject.json" || d.Name() == ".env" {
			return nil
		}

		rel, err := filepath.Rel(filepath.FromSlash(src), filepath.FromSlash(path))
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if strings.HasSuffix(d.Name(), ".templete") {
			trimmed := strings.TrimSuffix(targetPath, ".templete")
			return renderTemplateFromEmbed(path, trimmed, data)
		}

		return copyFileFromEmbed(path, targetPath)
	})
}

// renderTemplateFromEmbed reads a template from the embedded FS and renders it to dest.
func renderTemplateFromEmbed(src, dest string, data map[string]any) error {
	content, err := templatesFS.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read embedded template %s: %w", src, err)
	}

	tmpl, err := template.New(filepath.Base(src)).Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
		"Title":   titleCase,
	}).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	return tmpl.Execute(out, data)
}

// copyFileFromEmbed copies a single file from the embedded FS to the filesystem.
func copyFileFromEmbed(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	content, err := templatesFS.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read embedded file %s: %w", src, err)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.WriteString(out, string(content))
	return err
}

func newTemplateData(req Requirement) map[string]any {
	return map[string]any{
		"Requirement":  req,
		"ProjectName":  req.ProjectName,
		"PackageName":  req.PackageName,
		"Architecture": req.Arch,
		"Dependencies": req.Deps,
		"ConfigFields": []string{},
		"ConfigInit":   []string{},
	}
}
