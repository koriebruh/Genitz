// Package mcpserver exposes genitz's curated registry and lifecycle
// actions (search, info, presets, audit, list, add, remove, scaffold) as
// MCP tools, so an AI coding agent can call them directly with structured
// JSON in/out instead of shelling out to genitz and parsing text output.
package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/koriebruh/Genitz/internal/generator"
	"github.com/koriebruh/Genitz/internal/tui"
)

// NewServer builds the genitz MCP server.
func NewServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "genitz", Version: generator.Version}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_dependencies",
		Description: "Search the curated Go dependency registry by name, ID, category, or description (case-insensitive substring match).",
	}, searchDependencies)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_dependency_info",
		Description: "Get full details (import path, category, description, docs link) for one registry dependency ID.",
	}, getDependencyInfo)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_presets",
		Description: "List built-in and user-saved dependency bundle presets.",
	}, listPresets)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_installed_dependencies",
		Description: "List the direct dependencies already in a Go project's go.mod, cross-referenced against the registry.",
	}, listInstalledDependencies)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "audit_project",
		Description: "Check an existing Go project's dependencies against the curated registry for maintenance-mode packages, plus a govulncheck advisory.",
	}, auditProject)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_dependencies",
		Description: "Add one or more dependencies (by registry ID, optionally id@version) to the Go project in the given directory via go get + go mod tidy.",
	}, addDependencies)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove_dependencies",
		Description: "Remove one or more dependencies (by registry ID) from the Go project in the given directory via go get id@none + go mod tidy.",
	}, removeDependencies)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "scaffold_project",
		Description: "Scaffold a brand-new Go project: go mod init, optional Docker/CI/Makefile/README/community files/LICENSE, and go get for the chosen dependencies.",
	}, scaffoldProject)

	return s
}

// ── search_dependencies ──────────────────────────────────────────────────

type searchArgs struct {
	Query string `json:"query" jsonschema:"substring to match against dependency name, ID, category, or description"`
}

type dependencyResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	ImportPath  string `json:"importPath"`
	Description string `json:"description"`
	DocsURL     string `json:"docsUrl"`
}

func toDependencyResult(d tui.Dependency) dependencyResult {
	return dependencyResult{
		ID:          d.ID,
		Name:        d.Name,
		Category:    d.Category,
		ImportPath:  d.ImportPath,
		Description: d.Description,
		DocsURL:     "https://pkg.go.dev/" + d.ImportPath,
	}
}

type searchResult struct {
	Matches []dependencyResult `json:"matches"`
}

func searchDependencies(_ context.Context, _ *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, searchResult, error) {
	query := strings.ToLower(strings.TrimSpace(args.Query))
	var matches []dependencyResult
	for _, dep := range tui.DependencyRegistry {
		if query == "" ||
			strings.Contains(strings.ToLower(dep.Name), query) ||
			strings.Contains(strings.ToLower(dep.ID), query) ||
			strings.Contains(strings.ToLower(dep.Category), query) ||
			strings.Contains(strings.ToLower(dep.Description), query) {
			matches = append(matches, toDependencyResult(dep))
		}
	}
	return nil, searchResult{Matches: matches}, nil
}

// ── get_dependency_info ──────────────────────────────────────────────────

type infoArgs struct {
	ID string `json:"id" jsonschema:"the registry dependency ID, e.g. \"fiber\""`
}

func getDependencyInfo(_ context.Context, _ *mcp.CallToolRequest, args infoArgs) (*mcp.CallToolResult, dependencyResult, error) {
	dep, ok := tui.FindByID(args.ID)
	if !ok {
		return nil, dependencyResult{}, fmt.Errorf("unknown dependency ID %q", args.ID)
	}
	return nil, toDependencyResult(dep), nil
}

// ── list_presets ──────────────────────────────────────────────────────────

type noArgs struct{}

type presetResult struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DepIDs      []string `json:"depIds"`
}

type listPresetsResult struct {
	Presets []presetResult `json:"presets"`
}

func listPresets(_ context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, listPresetsResult, error) {
	var out []presetResult
	for _, p := range tui.AllPresets() {
		out = append(out, presetResult{ID: p.ID, Name: p.Name, Description: p.Description, DepIDs: p.DepIDs})
	}
	return nil, listPresetsResult{Presets: out}, nil
}

// ── list_installed_dependencies ──────────────────────────────────────────

type dirArgs struct {
	Dir string `json:"dir" jsonschema:"path to the Go project directory (must contain go.mod)"`
}

type installedResult struct {
	ImportPath string `json:"importPath"`
	Version    string `json:"version"`
	Name       string `json:"name,omitempty"`
	Category   string `json:"category,omitempty"`
	Managed    bool   `json:"managed"`
}

type listInstalledResult struct {
	Dependencies []installedResult `json:"dependencies"`
}

func listInstalledDependencies(_ context.Context, _ *mcp.CallToolRequest, args dirArgs) (*mcp.CallToolResult, listInstalledResult, error) {
	dir, err := resolveMCPPath(dirOrDot(args.Dir))
	if err != nil {
		return nil, listInstalledResult{}, err
	}
	deps, err := generator.ListInstalled(dir)
	if err != nil {
		return nil, listInstalledResult{}, err
	}
	var out []installedResult
	for _, d := range deps {
		out = append(out, installedResult{
			ImportPath: d.ImportPath, Version: d.Version, Name: d.Name, Category: d.Category, Managed: d.Managed,
		})
	}
	return nil, listInstalledResult{Dependencies: out}, nil
}

// ── audit_project ─────────────────────────────────────────────────────────

type auditFindingResult struct {
	ImportPath    string `json:"importPath"`
	Name          string `json:"name"`
	Detail        string `json:"detail"`
	ReplacementID string `json:"replacementId"`
}

type auditResult struct {
	Findings     []auditFindingResult `json:"findings"`
	VulnAdvisory string               `json:"vulnAdvisory,omitempty"`
}

func auditProject(_ context.Context, _ *mcp.CallToolRequest, args dirArgs) (*mcp.CallToolResult, auditResult, error) {
	dir, err := resolveMCPPath(dirOrDot(args.Dir))
	if err != nil {
		return nil, auditResult{}, err
	}
	findings, advisory, err := generator.AuditProject(dir)
	if err != nil {
		return nil, auditResult{}, err
	}
	var out []auditFindingResult
	for _, f := range findings {
		out = append(out, auditFindingResult{ImportPath: f.ImportPath, Name: f.Name, Detail: f.Detail, ReplacementID: f.ReplacementID})
	}
	return nil, auditResult{Findings: out, VulnAdvisory: advisory}, nil
}

// ── add_dependencies / remove_dependencies ───────────────────────────────

type depsArgs struct {
	Dir  string   `json:"dir" jsonschema:"path to the Go project directory (must contain go.mod)"`
	Deps []string `json:"deps" jsonschema:"registry dependency IDs — add_dependencies accepts an optional id@version pin"`
}

type stepsResult struct {
	Completed []string `json:"completed"`
	Error     string   `json:"error,omitempty"`
}

func addDependencies(_ context.Context, _ *mcp.CallToolRequest, args depsArgs) (*mcp.CallToolResult, stepsResult, error) {
	deps, versions, err := resolveDepIDs(args.Deps)
	if err != nil {
		return nil, stepsResult{}, err
	}
	dir, err := resolveMCPPath(dirOrDot(args.Dir))
	if err != nil {
		return nil, stepsResult{}, err
	}
	completed, runErr := runSteps(generator.BuildAddSteps(dir, deps, versions))
	if runErr != nil {
		return nil, stepsResult{Completed: completed, Error: runErr.Error()}, nil
	}
	return nil, stepsResult{Completed: completed}, nil
}

func removeDependencies(_ context.Context, _ *mcp.CallToolRequest, args depsArgs) (*mcp.CallToolResult, stepsResult, error) {
	deps, _, err := resolveDepIDs(args.Deps)
	if err != nil {
		return nil, stepsResult{}, err
	}
	dir, err := resolveMCPPath(dirOrDot(args.Dir))
	if err != nil {
		return nil, stepsResult{}, err
	}
	completed, runErr := runSteps(generator.BuildRemoveSteps(dir, deps))
	if runErr != nil {
		return nil, stepsResult{Completed: completed, Error: runErr.Error()}, nil
	}
	return nil, stepsResult{Completed: completed}, nil
}

// ── scaffold_project ──────────────────────────────────────────────────────

type scaffoldArgs struct {
	Name           string   `json:"name" jsonschema:"project folder name"`
	Module         string   `json:"module,omitempty" jsonschema:"Go module path (defaults to name)"`
	Deps           []string `json:"deps,omitempty" jsonschema:"registry dependency IDs, optionally id@version"`
	Preset         string   `json:"preset,omitempty" jsonschema:"a starter bundle ID to apply alongside deps"`
	Docker         bool     `json:"docker,omitempty"`
	CI             bool     `json:"ci,omitempty"`
	Makefile       bool     `json:"makefile,omitempty"`
	Readme         bool     `json:"readme,omitempty"`
	CommunityFiles bool     `json:"communityFiles,omitempty"`
	GitInit        bool     `json:"gitInit,omitempty"`
	License        string   `json:"license,omitempty" jsonschema:"one of \"\", \"mit\", or \"apache-2.0\""`
}

type scaffoldResult struct {
	Path      string   `json:"path"`
	Completed []string `json:"completed"`
	Error     string   `json:"error,omitempty"`
}

func scaffoldProject(_ context.Context, _ *mcp.CallToolRequest, args scaffoldArgs) (*mcp.CallToolResult, scaffoldResult, error) {
	if strings.TrimSpace(args.Name) == "" {
		return nil, scaffoldResult{}, fmt.Errorf("name is required")
	}
	if !generator.ValidLicenseKind(args.License) {
		return nil, scaffoldResult{}, fmt.Errorf("invalid license %q — expected \"\", \"mit\", or \"apache-2.0\"", args.License)
	}
	if _, err := resolveMCPPath(args.Name); err != nil {
		return nil, scaffoldResult{}, err
	}

	deps, versions, err := resolveDepIDs(args.Deps)
	if err != nil {
		return nil, scaffoldResult{}, err
	}
	presetDeps, err := resolvePresetIDs(args.Preset)
	if err != nil {
		return nil, scaffoldResult{}, err
	}
	deps = mergeDepMaps(presetDeps, deps)

	module := args.Module
	if module == "" {
		module = args.Name
	}

	req := generator.Requirement{
		ProjectName:           args.Name,
		PackageName:           module,
		Deps:                  deps,
		DepVersions:           versions,
		IncludeDocker:         args.Docker,
		IncludeCI:             args.CI,
		IncludeMakefile:       args.Makefile,
		IncludeReadme:         args.Readme,
		IncludeCommunityFiles: args.CommunityFiles,
		IncludeGitInit:        args.GitInit,
		License:               args.License,
	}

	targetPath, err := generator.PrepareNewProject(req)
	if err != nil {
		return nil, scaffoldResult{}, err
	}

	completed, runErr := runSteps(generator.BuildInstallSteps(targetPath, req))
	if runErr != nil {
		return nil, scaffoldResult{Path: targetPath, Completed: completed, Error: runErr.Error()}, nil
	}
	return nil, scaffoldResult{Path: targetPath, Completed: completed}, nil
}

// ── shared helpers ─────────────────────────────────────────────────────────

func dirOrDot(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return "."
	}
	return dir
}

// resolveMCPPath resolves dir/name to an absolute path and, if
// GENITZ_MCP_ROOT is set, rejects any target that escapes it. Unlike the
// plain CLI (which only ever touches cwd), every MCP tool below accepts an
// arbitrary directory/name from the calling agent — a value that can be
// influenced by untrusted content the agent ingested (indirect prompt
// injection), not just the user's own typed input. GENITZ_MCP_ROOT is
// opt-in and unset by default so a host legitimately targeting other
// projects on disk isn't broken; setting it confines every path-accepting
// tool call to one subtree.
func resolveMCPPath(raw string) (string, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", raw, err)
	}
	root := strings.TrimSpace(os.Getenv("GENITZ_MCP_ROOT"))
	if root == "" {
		return abs, nil
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve GENITZ_MCP_ROOT %q: %w", root, err)
	}
	rel, err := filepath.Rel(absRoot, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes GENITZ_MCP_ROOT %q", abs, absRoot)
	}
	return abs, nil
}

// runSteps runs InstallSteps sequentially, collecting each completed
// label — the MCP-tool equivalent of main.go's runStepsPlain, returning
// structured data instead of printing progress lines.
func runSteps(steps []tui.InstallStep) (completed []string, err error) {
	for _, step := range steps {
		if err := step.Run(); err != nil {
			return completed, fmt.Errorf("%s: %w", step.Label, err)
		}
		completed = append(completed, step.Label)
	}
	return completed, nil
}

// resolveDepIDs mirrors main.go's resolveDeps (id@version parsing) — can't
// import package main, so this is a small intentional duplication rather
// than a cross-package refactor.
func resolveDepIDs(ids []string) (map[int]tui.Dependency, map[string]string, error) {
	deps := make(map[int]tui.Dependency)
	versions := make(map[string]string)
	for i, entry := range ids {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		id, version, _ := strings.Cut(entry, "@")
		id = strings.TrimSpace(id)
		dep, ok := tui.FindByID(id)
		if !ok {
			return nil, nil, fmt.Errorf("unknown dependency ID %q", id)
		}
		deps[i] = dep
		if version != "" {
			versions[dep.ImportPath] = version
		}
	}
	return deps, versions, nil
}

func resolvePresetIDs(presetID string) (map[int]tui.Dependency, error) {
	deps := make(map[int]tui.Dependency)
	presetID = strings.TrimSpace(presetID)
	if presetID == "" {
		return deps, nil
	}
	preset, ok := tui.FindPreset(presetID)
	if !ok {
		return nil, fmt.Errorf("unknown preset %q", presetID)
	}
	for i, id := range preset.DepIDs {
		dep, ok := tui.FindByID(id)
		if !ok {
			return nil, fmt.Errorf("preset %q references unknown dependency ID %q", presetID, id)
		}
		deps[i] = dep
	}
	return deps, nil
}

func mergeDepMaps(a, b map[int]tui.Dependency) map[int]tui.Dependency {
	seen := make(map[string]bool, len(a))
	merged := make(map[int]tui.Dependency, len(a)+len(b))
	i := 0
	for _, dep := range a {
		if seen[dep.ImportPath] {
			continue
		}
		seen[dep.ImportPath] = true
		merged[i] = dep
		i++
	}
	for _, dep := range b {
		if seen[dep.ImportPath] {
			continue
		}
		seen[dep.ImportPath] = true
		merged[i] = dep
		i++
	}
	return merged
}
