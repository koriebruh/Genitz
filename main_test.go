package main

import (
	"testing"

	"github.com/koriebruh/Genitz/internal/tui"
)

func TestResolveDepsPlain(t *testing.T) {
	deps, versions, err := resolveDeps("fiber,redis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}
	if len(versions) != 0 {
		t.Fatalf("expected no pinned versions, got %v", versions)
	}
}

func TestResolveDepsWithVersionPin(t *testing.T) {
	deps, versions, err := resolveDeps("redis@v9.5.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dep, got %d", len(deps))
	}
	dep := deps[0]
	if versions[dep.ImportPath] != "v9.5.1" {
		t.Fatalf("expected pinned version v9.5.1 for %s, got %q", dep.ImportPath, versions[dep.ImportPath])
	}
}

func TestResolveDepsUnknownID(t *testing.T) {
	if _, _, err := resolveDeps("not-a-real-id"); err == nil {
		t.Fatal("expected an error for an unknown dependency ID")
	}
}

func TestResolvePresetDeps(t *testing.T) {
	preset, ok := tui.FindPreset("web-api")
	if !ok {
		t.Fatal("expected the web-api preset to exist")
	}

	deps, err := resolvePresetDeps("web-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Every DepID must resolve — a gap here would mean a preset silently
	// applies fewer dependencies than it advertises.
	if len(deps) != len(preset.DepIDs) {
		t.Fatalf("expected all %d web-api DepIDs to resolve, got %d", len(preset.DepIDs), len(deps))
	}
}

func TestAllBuiltinPresetsFullyResolve(t *testing.T) {
	for _, preset := range tui.Presets {
		deps, err := resolvePresetDeps(preset.ID)
		if err != nil {
			t.Errorf("preset %q: %v", preset.ID, err)
			continue
		}
		if len(deps) != len(preset.DepIDs) {
			t.Errorf("preset %q: expected %d resolved deps, got %d", preset.ID, len(preset.DepIDs), len(deps))
		}
	}
}

func TestResolvePresetDepsUnknown(t *testing.T) {
	if _, err := resolvePresetDeps("not-a-real-preset"); err == nil {
		t.Fatal("expected an error for an unknown preset")
	}
}

func TestMergeDepsDedupesByImportPath(t *testing.T) {
	presetDeps, err := resolvePresetDeps("web-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	flatDeps, _, err := resolveDeps("fiber") // fiber is also in the web-api preset
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	merged := mergeDeps(presetDeps, flatDeps)

	seen := make(map[string]bool, len(merged))
	for _, dep := range merged {
		if seen[dep.ImportPath] {
			t.Fatalf("duplicate import path %q in merged result", dep.ImportPath)
		}
		seen[dep.ImportPath] = true
	}
	if len(merged) != len(presetDeps) {
		t.Fatalf("expected merge of an overlapping dep to add nothing new, preset had %d, merged has %d", len(presetDeps), len(merged))
	}
}

// TestCompletionScriptsCoverAllSubcommands guards against subcommandNames
// drifting from main()'s actual dispatch — there's no cobra auto-gen here,
// so this is the only thing keeping the two in sync.
func TestCompletionScriptsCoverAllSubcommands(t *testing.T) {
	for _, script := range []string{completionBash(), completionZsh(), completionFish()} {
		for _, name := range subcommandNames {
			if !contains(script, name) {
				t.Errorf("completion script missing subcommand %q:\n%s", name, script)
			}
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
