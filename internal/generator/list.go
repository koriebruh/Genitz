package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/koriebruh/Genitz/internal/tui"
)

// InstalledDep is one entry from go.mod's require block, cross-referenced
// against the dependency registry so `genitz list` can show a friendly name
// and category instead of a bare import path.
type InstalledDep struct {
	ImportPath string
	Version    string
	Name       string // "" when the import path has no registry match
	Category   string
	Managed    bool // true when it matches a registry entry
}

// ListInstalled reads the direct requires from go.mod in dir (indirect ones
// are skipped — they're transitive, not something the user picked) and
// cross-references each import path against tui.DependencyRegistry.
func ListInstalled(dir string) ([]InstalledDep, error) {
	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, err
	}

	byImportPath := make(map[string]tui.Dependency, len(tui.DependencyRegistry))
	for _, dep := range tui.DependencyRegistry {
		byImportPath[dep.ImportPath] = dep
	}

	var deps []InstalledDep
	inRequireBlock := false
	for _, raw := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(raw)

		switch {
		case strings.HasPrefix(trimmed, "require ("):
			inRequireBlock = true
			continue
		case trimmed == ")":
			inRequireBlock = false
			continue
		case strings.HasPrefix(trimmed, "require ") && !strings.Contains(trimmed, "("):
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "require "))
		case !inRequireBlock:
			continue
		}

		if trimmed == "" || strings.Contains(trimmed, "// indirect") {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}
		importPath, version := fields[0], fields[1]

		d := InstalledDep{ImportPath: importPath, Version: version}
		if dep, ok := byImportPath[importPath]; ok {
			d.Name = dep.Name
			d.Category = dep.Category
			d.Managed = true
		}
		deps = append(deps, d)
	}

	sort.Slice(deps, func(i, j int) bool { return deps[i].ImportPath < deps[j].ImportPath })
	return deps, nil
}

// PrintInstalled renders ListInstalled's result as a table — registry
// matches show name/category, unmatched import paths are marked unmanaged
// rather than silently omitted, since they're still real entries in go.mod.
func PrintInstalled(deps []InstalledDep) {
	if len(deps) == 0 {
		fmt.Println("No direct dependencies in go.mod.")
		return
	}

	fmt.Println("📦 Installed dependencies:")
	for _, d := range deps {
		if d.Managed {
			fmt.Printf("   %-24s %-10s %-14s %s\n", d.Name, d.Version, d.Category, d.ImportPath)
		} else {
			fmt.Printf("   %-24s %-10s %-14s %s\n", "(unmanaged)", d.Version, "-", d.ImportPath)
		}
	}
}
