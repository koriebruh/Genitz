package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// userRegistryPath returns $XDG_CONFIG_HOME/genitz/registry.json (or the
// platform equivalent via os.UserConfigDir) — the optional file a user or
// team can drop private/internal dependency entries into without forking
// genitz. Returns "" if the config directory can't be determined.
func userRegistryPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "genitz", "registry.json")
}

// loadUserRegistry merges an optional user/team registry override into
// base: an entry whose ID matches an existing one overrides it in place, a
// new ID is appended. Unlike the embedded registry.json (mustLoadRegistry,
// which panics on invalid JSON since that would be a genitz bug), this file
// is untrusted external input — a missing file, invalid JSON, or an entry
// missing a required field is skipped with a stderr warning rather than
// crashing the CLI.
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

	for _, dep := range overrides {
		if dep.ID == "" || dep.Name == "" || dep.Category == "" || dep.ImportPath == "" {
			fmt.Fprintf(os.Stderr, "genitz: skipping invalid user registry entry (missing ID/Name/Category/ImportPath): %+v\n", dep)
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
