package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/koriebruh/Genitz/internal/tui"
)

const templateRoot = "internal/generator/templetes"

var architectureCatalog = map[string]tui.Architecture{
	tui.ArchMicro: {
		Name:        tui.ArchMicro,
		Description: "Service-per-domain layout with shared pkg folder",
		TemplateDir: filepath.Join(templateRoot, "architecture", "microservice"),
	},
	tui.ArchClean: {
		Name:        tui.ArchClean,
		Description: "Layered clean architecture with adapters/core split",
		TemplateDir: filepath.Join(templateRoot, "architecture", "clean-architecture"),
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

type FeatureInjection struct {
	TargetDir   string `json:"target_dir"`   // e.g. "config"
	ConfigField string `json:"config_field"` // e.g. "Redis RedisConfig"
	ConfigInit  string `json:"config_init"`  // e.g. "Redis: LoadRedisConfig(),"
}

func enrichConfigData(req Requirement, data map[string]any) {
	configFields := []string{}
	configInit := []string{}

	for _, dep := range req.Deps {
		if dep.TemplateDir == "" {
			continue
		}

		// Try to read inject.json from template dir
		injectFile := filepath.Join(dep.TemplateDir, "inject.json")
		content, err := os.ReadFile(injectFile)
		if err != nil {
			// If file doesn't exist, skip injection
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
	src := filepath.Join(templateRoot, "feature", "config.go.templete")
	dest := filepath.Join(target, "config", "config.go")
	fmt.Printf("🔧 Generating config: %s\n", dest)
	return renderTemplateFile(src, dest, data)
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
	if !filepath.IsAbs(src) {
		abs, err := filepath.Abs(src)
		if err != nil {
			return fmt.Errorf("resolve architecture template: %w", err)
		}
		src = abs
	}

	fmt.Printf("🏗️  Applying architecture template: %s\n", req.Arch.Name)
	return copyDirWithTemplates(src, target, data)
}

func applyDependencyTemplates(target string, req Requirement, data map[string]any) error {
	for _, dep := range req.Deps {
		if dep.TemplateDir == "" {
			continue
		}

		src := dep.TemplateDir
		if !filepath.IsAbs(src) {
			abs, err := filepath.Abs(src)
			if err != nil {
				return fmt.Errorf("resolve dependency template for %s: %w", dep.Name, err)
			}
			src = abs
		}

		// Determine destination from inject.json if available
		relDest := "internal/features/" + dep.ID
		injectFile := filepath.Join(src, "inject.json")
		if content, err := os.ReadFile(injectFile); err == nil {
			var inject FeatureInjection
			if json.Unmarshal(content, &inject) == nil && inject.TargetDir != "" {
				relDest = inject.TargetDir
			}
		}

		dest := filepath.Join(target, relDest)
		fmt.Printf("📦 Injecting feature template: %s -> %s\n", dep.Name, dest)
		if err := copyDirWithTemplates(src, dest, data); err != nil {
			return fmt.Errorf("copy dependency template %s: %w", dep.Name, err)
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

		srcEnv := filepath.Join(dep.TemplateDir, ".env")
		content, err := os.ReadFile(srcEnv)
		if err != nil {
			continue // Skip if no .env
		}

		envContent += fmt.Sprintf("\n# %s Configuration\n", strings.Title(dep.Name))

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

func copyDirWithTemplates(src, dest string, data map[string]any) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip inject.json and .env files from being copied to final project
		if info.Name() == "inject.json" || info.Name() == ".env" {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, rel)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if strings.HasSuffix(info.Name(), ".templete") {
			trimmed := strings.TrimSuffix(targetPath, ".templete")
			return renderTemplateFile(path, trimmed, data)
		}

		return copyFile(path, targetPath)
	})
}

func renderTemplateFile(src, dest string, data map[string]any) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	tmpl, err := template.New(filepath.Base(src)).Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
	}).Parse(string(content))
	if err != nil {
		return err
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

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
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
