package generator

import (
	"strings"
	"testing"

	"github.com/koriebruh/Genitz/internal/tui"
)

func TestArchitectureDiagramNoDeps(t *testing.T) {
	if _, ok := architectureDiagram(Requirement{ProjectName: "app"}); ok {
		t.Fatal("expected ok=false with zero dependencies")
	}
}

func TestArchitectureDiagramWithDeps(t *testing.T) {
	req := Requirement{
		ProjectName: "shoply-api",
		Deps: map[int]tui.Dependency{
			0: {ID: "fiber", Name: "Fiber", Category: "framework"},
			1: {ID: "postgres-driver", Name: "pgx (PostgreSQL)", Category: "driver"},
		},
	}
	diagram, ok := architectureDiagram(req)
	if !ok {
		t.Fatal("expected ok=true with dependencies selected")
	}
	if !strings.HasPrefix(diagram, "```mermaid\nflowchart LR\n") {
		t.Fatalf("expected a mermaid flowchart fence, got:\n%s", diagram)
	}
	if !strings.HasSuffix(diagram, "```\n") {
		t.Fatalf("expected the diagram to close its code fence, got:\n%s", diagram)
	}
	for _, want := range []string{"Fiber", "pgx (PostgreSQL)", "shoply_api"} {
		if !strings.Contains(diagram, want) {
			t.Errorf("expected diagram to mention %q, got:\n%s", want, diagram)
		}
	}
	// Node IDs must never contain a raw hyphen — postgres-driver's ID has
	// one, and an un-sanitized hyphen is unreliable across Mermaid parsers.
	if strings.Contains(diagram, "postgres-driver[") {
		t.Errorf("expected the hyphenated ID to be sanitized in the node identifier, got:\n%s", diagram)
	}
}

func TestReadmeContentIncludesArchitectureSectionOnlyWithDeps(t *testing.T) {
	withDeps := readmeContent(Requirement{
		ProjectName: "app",
		PackageName: "app",
		Deps:        map[int]tui.Dependency{0: {ID: "fiber", Name: "Fiber", Category: "framework"}},
	})
	if !strings.Contains(withDeps, "## Architecture") {
		t.Error("expected an Architecture section when deps are selected")
	}

	withoutDeps := readmeContent(Requirement{ProjectName: "app", PackageName: "app"})
	if strings.Contains(withoutDeps, "## Architecture") {
		t.Error("expected no Architecture section with zero deps")
	}
}
