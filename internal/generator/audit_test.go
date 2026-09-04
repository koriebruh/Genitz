package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/koriebruh/Genitz/internal/tui"
)

func TestAuditProjectFlagsDeprecatedDependency(t *testing.T) {
	dir := t.TempDir()
	content := "module example.com/app\n\ngo 1.25.0\n\nrequire github.com/gorilla/mux v1.8.1\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, _, err := AuditProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for gorilla/mux, got %d: %+v", len(findings), findings)
	}
	if findings[0].ReplacementID != "chi" {
		t.Fatalf("expected chi as the suggested replacement, got %q", findings[0].ReplacementID)
	}
}

func TestAuditProjectCleanStack(t *testing.T) {
	dir := t.TempDir()
	content := "module example.com/app\n\ngo 1.25.0\n\nrequire github.com/gofiber/fiber/v3 v3.0.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, _, err := AuditProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings for a non-deprecated dependency, got %+v", findings)
	}
}

func TestDeprecatedReplacementsPointToRealRegistryEntries(t *testing.T) {
	// Guards against the audit map referencing an ID that doesn't (or no
	// longer) exists in the registry — a dangling replacement suggestion.
	for id, replacement := range deprecatedReplacements {
		if _, ok := tui.FindByID(id); !ok {
			t.Errorf("deprecatedReplacements key %q is not itself a real registry ID", id)
		}
		if _, ok := tui.FindByID(replacement.ReplacementID); !ok {
			t.Errorf("deprecatedReplacements[%q].ReplacementID %q is not a real registry ID", id, replacement.ReplacementID)
		}
	}
}
