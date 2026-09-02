package tui

import "testing"

func TestDependencyRegistryLoadsAndHasNoDuplicateIDs(t *testing.T) {
	if len(DependencyRegistry) == 0 {
		t.Fatal("registry.json failed to load any entries")
	}

	seen := make(map[string]bool, len(DependencyRegistry))
	for _, dep := range DependencyRegistry {
		if dep.ID == "" {
			t.Errorf("entry %q has an empty ID", dep.Name)
		}
		if seen[dep.ID] {
			t.Errorf("duplicate ID %q", dep.ID)
		}
		seen[dep.ID] = true
	}
}

func TestFindByID(t *testing.T) {
	if _, ok := FindByID("fiber"); !ok {
		t.Error("expected to find fiber")
	}
	if _, ok := FindByID("not-a-real-id"); ok {
		t.Error("expected not to find a bogus ID")
	}
}
