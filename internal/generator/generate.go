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
	ProjectName     string
	PackageName     string
	Deps            map[int]tui.Dependency
	IncludeDocker   bool
	IncludeCI       bool
	IncludeMakefile bool
	IncludeGitInit  bool
	IncludeReadme   bool
	// IncludeCommunityFiles bundles CONTRIBUTING.md, SECURITY.md, GitHub
	// issue templates, and (if IncludeCI is also on) .github/dependabot.yml
	// — one toggle rather than four, same "keep the extras list short"
	// spirit as the other checkboxes.
	IncludeCommunityFiles bool
	// License is one of "", "none", "mit", "apache-2.0" — "" and "none" both
	// mean "generate nothing", kept as two spellings so the zero value (flag
	// not passed / TUI default) behaves the same as an explicit "none".
	License string
	// DepVersions maps an import path to a pinned version (e.g. "v2.50.0"),
	// used by --deps id@version in the non-interactive flag flow. Nil/missing
	// entries install unpinned (latest), same as before this field existed.
	DepVersions map[string]string
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
		ProjectName:           projectName,
		PackageName:           packageName,
		Deps:                  deps,
		IncludeDocker:         m.IncludeDocker,
		IncludeCI:             m.IncludeCI,
		IncludeMakefile:       m.IncludeMakefile,
		IncludeGitInit:        m.IncludeGitInit,
		IncludeReadme:         m.IncludeReadme,
		IncludeCommunityFiles: m.IncludeCommunityFiles,
		License:               m.LicenseChoice,
	}, nil
}

const bareMainGo = "package main\n\nfunc main() {}\n"

const editorconfigContent = `root = true

[*]
indent_style = space
indent_size = 4
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.go]
indent_style = tab

[Makefile]
indent_style = tab
`

// CheckPreconditions validates req (including the git-on-PATH check when
// IncludeGitInit is set) and confirms the target directory doesn't already
// exist — without creating anything. It's what makes --dry-run an accurate
// preview instead of one that can report success for a run that would
// immediately fail; PrepareNewProject calls it too, so the real run and the
// dry-run share exactly one precondition check.
func CheckPreconditions(req Requirement) (Requirement, error) {
	if err := req.validate(); err != nil {
		return req, err
	}

	targetPath, err := filepath.Abs(req.ProjectName)
	if err != nil {
		return req, fmt.Errorf("resolve project path: %w", err)
	}

	if err := statFreshProjectDir(targetPath); err != nil {
		return req, err
	}

	return req, nil
}

// PrepareNewProject validates req, creates the target directory, and writes
// a bare main.go. It's synchronous and network-free — the animated part
// (go mod / go get) is BuildInstallSteps, run separately so the TUI can
// show progress for it.
func PrepareNewProject(req Requirement) (targetPath string, err error) {
	req, err = CheckPreconditions(req)
	if err != nil {
		return "", err
	}

	targetPath, err = filepath.Abs(req.ProjectName)
	if err != nil {
		return "", fmt.Errorf("resolve project path: %w", err)
	}

	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return "", fmt.Errorf("create project directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(targetPath, "main.go"), []byte(bareMainGo), 0o644); err != nil {
		return "", fmt.Errorf("write main.go: %w", err)
	}

	// .editorconfig is universally useful and has no toggle — same
	// unconditional treatment as main.go itself, unlike Docker/CI/etc.
	// which are opt-in because they carry real behavioral consequences.
	if err := os.WriteFile(filepath.Join(targetPath, ".editorconfig"), []byte(editorconfigContent), 0o644); err != nil {
		return "", fmt.Errorf("write .editorconfig: %w", err)
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

	hasCompose := false
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
			hasCompose = true
			steps = append(steps, tui.InstallStep{
				Label: "Generating docker-compose.yml",
				Run: func() error {
					if err := os.WriteFile(filepath.Join(targetPath, "docker-compose.yml"), []byte(content), 0o644); err != nil {
						return fmt.Errorf("write docker-compose.yml: %w", err)
					}
					return nil
				},
			})

			if envContent, ok := envExampleContent(req); ok {
				steps = append(steps, tui.InstallStep{
					Label: "Generating .env.example",
					Run: func() error {
						return os.WriteFile(filepath.Join(targetPath, ".env.example"), []byte(envContent), 0o644)
					},
				})
			}
		}
	}

	if req.IncludeCI {
		steps = append(steps, tui.InstallStep{
			Label: "Generating GitHub Actions CI workflow",
			Run: func() error {
				dir := filepath.Join(targetPath, ".github", "workflows")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("create .github/workflows: %w", err)
				}
				content := ciWorkflowContent(readGoVersion(targetPath))
				if err := os.WriteFile(filepath.Join(dir, "ci.yml"), []byte(content), 0o644); err != nil {
					return fmt.Errorf("write ci.yml: %w", err)
				}
				return nil
			},
		}, tui.InstallStep{
			Label: "Generating .golangci.yml",
			Run: func() error {
				return os.WriteFile(filepath.Join(targetPath, ".golangci.yml"), []byte(golangciConfigContent()), 0o644)
			},
		})
	}

	if req.IncludeMakefile {
		hasDockerfile := req.IncludeDocker
		steps = append(steps, tui.InstallStep{
			Label: "Generating Makefile",
			Run: func() error {
				content := makefileContent(hasDockerfile, hasCompose)
				if err := os.WriteFile(filepath.Join(targetPath, "Makefile"), []byte(content), 0o644); err != nil {
					return fmt.Errorf("write Makefile: %w", err)
				}
				return nil
			},
		})
	}

	if req.IncludeReadme {
		steps = append(steps, tui.InstallStep{
			Label: "Generating README.md",
			Run: func() error {
				return os.WriteFile(filepath.Join(targetPath, "README.md"), []byte(readmeContent(req)), 0o644)
			},
		})
	}

	if req.IncludeCommunityFiles {
		steps = append(steps, tui.InstallStep{
			Label: "Generating community files (CONTRIBUTING, SECURITY, issue templates)",
			Run: func() error {
				if err := os.WriteFile(filepath.Join(targetPath, "CONTRIBUTING.md"), []byte(contributingContent(req.ProjectName)), 0o644); err != nil {
					return fmt.Errorf("write CONTRIBUTING.md: %w", err)
				}
				if err := os.WriteFile(filepath.Join(targetPath, "SECURITY.md"), []byte(securityContent()), 0o644); err != nil {
					return fmt.Errorf("write SECURITY.md: %w", err)
				}
				issueDir := filepath.Join(targetPath, ".github", "ISSUE_TEMPLATE")
				if err := os.MkdirAll(issueDir, 0o755); err != nil {
					return fmt.Errorf("create .github/ISSUE_TEMPLATE: %w", err)
				}
				if err := os.WriteFile(filepath.Join(issueDir, "bug_report.md"), []byte(bugReportIssueTemplate), 0o644); err != nil {
					return fmt.Errorf("write bug_report.md: %w", err)
				}
				if err := os.WriteFile(filepath.Join(issueDir, "feature_request.md"), []byte(featureRequestIssueTemplate), 0o644); err != nil {
					return fmt.Errorf("write feature_request.md: %w", err)
				}
				if err := os.WriteFile(filepath.Join(targetPath, ".github", "dependabot.yml"), []byte(dependabotContent(req.IncludeCI)), 0o644); err != nil {
					return fmt.Errorf("write dependabot.yml: %w", err)
				}
				return nil
			},
		})
	}

	if content, ok := licenseContent(req.License, ResolveLicenseHolder()); ok {
		steps = append(steps, tui.InstallStep{
			Label: "Generating LICENSE",
			Run: func() error {
				return os.WriteFile(filepath.Join(targetPath, "LICENSE"), []byte(content), 0o644)
			},
		})
	}

	for _, importPath := range sortedImportPaths(req.Deps) {
		arg := installArg(importPath, req.DepVersions)
		steps = append(steps, tui.InstallStep{
			Label: "Installing " + arg,
			Run:   func() error { return runCaptured(targetPath, "go", "get", arg) },
		})
	}

	if dep, ok := detectConfigLib(req.Deps); ok {
		if content, ok := configStubContent(dep); ok {
			steps = append(steps, tui.InstallStep{
				Label: "Generating config/config.go (" + dep.Name + ")",
				Run: func() error {
					dir := filepath.Join(targetPath, "config")
					if err := os.MkdirAll(dir, 0o755); err != nil {
						return fmt.Errorf("create config dir: %w", err)
					}
					if err := os.WriteFile(filepath.Join(dir, "config.go"), []byte(content), 0o644); err != nil {
						return fmt.Errorf("write config/config.go: %w", err)
					}
					return nil
				},
			})
		}
	}

	steps = append(steps,
		tui.InstallStep{Label: "Tidying go.mod", Run: func() error { return runCaptured(targetPath, "go", "mod", "tidy") }},
		tui.InstallStep{Label: "Formatting code", Run: func() error { return runCaptured(targetPath, "go", "fmt", "./...") }},
	)

	if req.IncludeGitInit {
		steps = append(steps, tui.InstallStep{
			Label: "Initialising git repository",
			Run:   func() error { return runGitInit(targetPath) },
		})
	}
	return steps
}

// BuildAddSteps returns the ordered go get/go mod tidy steps for installing
// deps into the existing project at targetDir. versions optionally pins an
// import path to a specific version (see installArg) — pass nil for
// unpinned/latest installs (the interactive TUI never collects a version).
func BuildAddSteps(targetDir string, deps map[int]tui.Dependency, versions map[string]string) []tui.InstallStep {
	var steps []tui.InstallStep
	for _, importPath := range sortedImportPaths(deps) {
		arg := installArg(importPath, versions)
		steps = append(steps, tui.InstallStep{
			Label: "Installing " + arg,
			Run:   func() error { return runCaptured(targetDir, "go", "get", arg) },
		})
	}
	steps = append(steps, tui.InstallStep{
		Label: "Tidying go.mod",
		Run:   func() error { return runCaptured(targetDir, "go", "mod", "tidy") },
	})
	return steps
}

// BuildRemoveSteps returns the ordered go get <path>@none/go mod tidy steps
// for dropping deps from the project at targetDir. `go get <path>@none` is
// the documented way to remove a requirement — go mod tidy then cleans up
// anything left unreferenced (if the code still imports it, tidy will
// re-add it, which is correct: genitz can't know that without a full
// source scan, so it doesn't try).
func BuildRemoveSteps(targetDir string, deps map[int]tui.Dependency) []tui.InstallStep {
	var steps []tui.InstallStep
	for _, importPath := range sortedImportPaths(deps) {
		steps = append(steps, tui.InstallStep{
			Label: "Removing " + importPath,
			Run:   func() error { return runCaptured(targetDir, "go", "get", importPath+"@none") },
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

	if r.IncludeGitInit {
		if err := CheckBinary("git"); err != nil {
			return fmt.Errorf("git init requested: %w", err)
		}
	}

	return nil
}

// statFreshProjectDir confirms path doesn't already exist, without creating
// it — shared by CheckPreconditions (dry-run-safe) and PrepareNewProject
// (which creates the directory itself right after calling this).
func statFreshProjectDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("destination %q already exists", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check destination: %w", err)
	}
	return nil
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

// installArg returns the "go get" argument for importPath — pinned to
// versions[importPath] (e.g. "example.com/pkg@v1.2.3") when present,
// otherwise the bare import path (unpinned/latest).
func installArg(importPath string, versions map[string]string) string {
	if v, ok := versions[importPath]; ok && v != "" {
		return importPath + "@" + v
	}
	return importPath
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
