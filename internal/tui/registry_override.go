package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// registryOverrideEnvVar lets tests point at a specific file instead of the
// real $XDG_CONFIG_HOME, so a stray file on a dev/CI box can't change what
// DependencyRegistry's package-init load sees.
const registryOverrideEnvVar = "GENITZ_REGISTRY_OVERRIDE"

// userRegistryPath returns $XDG_CONFIG_HOME/genitz/registry.json (or the
// platform equivalent via os.UserConfigDir), or GENITZ_REGISTRY_OVERRIDE's
// value if set — the optional file a user or team can drop private/internal
// dependency entries into without forking genitz. Returns "" if neither is
// available.
func userRegistryPath() string {
	if p := os.Getenv(registryOverrideEnvVar); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "genitz", "registry.json")
}

// validRegistryCategories returns the categories wired into the picker
// (depGroups) — anything else would resolve via FindByID but never show up
// as a row in the TUI.
func validRegistryCategories() map[string]bool {
	set := make(map[string]bool)
	for _, g := range depGroups {
		for _, c := range g.categories {
			set[c] = true
		}
	}
	return set
}

// loadUserRegistry merges an optional user/team registry override into
// base: a matching ID overrides in place, a new ID is appended. Unlike the
// embedded registry.json (mustLoadRegistry, which panics on bad JSON — a
// genitz bug), this is untrusted external input, so problems are skipped
// with a stderr warning instead of crashing the CLI.
func loadUserRegistry(base []Dependency) []Dependency {
	path := userRegistryPath()
	if path == "" {
		return base
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return base // no override file present — nothing to merge, not an error
	}

	var overrides []Dependency
	if err := json.Unmarshal(content, &overrides); err != nil {
		fmt.Fprintf(os.Stderr, "genitz: ignoring invalid user registry at %s: %v\n", path, err)
		return base
	}

	merged := make([]Dependency, len(base))
	copy(merged, base)
	byID := make(map[string]int, len(merged))
	for i, dep := range merged {
		byID[dep.ID] = i
	}
	validCategories := validRegistryCategories()

	for _, dep := range overrides {
		if dep.ID == "" || dep.Name == "" || dep.Category == "" || dep.ImportPath == "" {
			fmt.Fprintf(os.Stderr, "genitz: skipping invalid user registry entry (missing ID/Name/Category/ImportPath): %+v\n", dep)
			continue
		}
		if !validCategories[dep.Category] {
			fmt.Fprintf(os.Stderr, "genitz: skipping user registry entry %q — unknown category %q (see internal/tui/dependencies.go depGroups for valid categories, or it'll never appear in the picker)\n", dep.ID, dep.Category)
			continue
		}
		if idx, ok := byID[dep.ID]; ok {
			merged[idx] = dep
		} else {
			byID[dep.ID] = len(merged)
			merged = append(merged, dep)
		}
	}
	return merged
}
