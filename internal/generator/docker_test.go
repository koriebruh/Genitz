package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koriebruh/Genitz/internal/tui"
)

func TestComposeContentWithServices(t *testing.T) {
	req := Requirement{Deps: map[int]tui.Dependency{
		0: {ID: "redis"},
		1: {ID: "postgres-driver"},
		2: {ID: "fiber"}, // not infra-backed — should not produce a service block
	}}

	content, ok := composeContent(req)
	if !ok {
		t.Fatal("expected ok=true when an infra-backed dependency is selected")
	}
	for _, want := range []string{"app:", "redis:", "image: redis:7-alpine", "postgres:", "image: postgres:16-alpine"} {
		if !strings.Contains(content, want) {
			t.Errorf("compose content missing %q\n---\n%s", want, content)
		}
	}
	if strings.Contains(content, "fiber") {
		t.Error("compose content should not mention a non-infra dependency")
	}
}

func TestEnvExampleContentWithEnvVars(t *testing.T) {
	req := Requirement{Deps: map[int]tui.Dependency{
		0: {ID: "postgres-driver"},
		1: {ID: "redis"}, // no environment — shouldn't produce its own section
	}}

	content, ok := envExampleContent(req)
	if !ok {
		t.Fatal("expected ok=true when a service declares environment vars")
	}
	if !strings.Contains(content, "POSTGRES_PASSWORD=postgres") {
		t.Errorf("expected postgres env vars in output, got:\n%s", content)
	}
	if strings.Contains(content, "# redis") {
		t.Errorf("expected no section for a service with no env vars, got:\n%s", content)
	}
}

func TestEnvExampleContentNoEnvVars(t *testing.T) {
	req := Requirement{Deps: map[int]tui.Dependency{0: {ID: "redis"}}}
	if _, ok := envExampleContent(req); ok {
		t.Fatal("expected ok=false when the only selected service has no environment vars")
	}
}

func TestEnvExampleContentNoServices(t *testing.T) {
	req := Requirement{Deps: map[int]tui.Dependency{0: {ID: "fiber"}}}
	if _, ok := envExampleContent(req); ok {
		t.Fatal("expected ok=false when nothing selected maps to a compose service")
	}
}

func TestComposeContentDedupesSharedService(t *testing.T) {
	req := Requirement{Deps: map[int]tui.Dependency{
		0: {ID: "kafka-go"},
		1: {ID: "sarama"},
	}}

	content, ok := composeContent(req)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if n := strings.Count(content, "kafka:\n"); n != 1 {
		t.Errorf("expected exactly one kafka service block, got %d\n---\n%s", n, content)
	}
}

func TestComposeContentNoServices(t *testing.T) {
	req := Requirement{Deps: map[int]tui.Dependency{
		0: {ID: "fiber"},
		1: {ID: "testify"},
	}}

	if _, ok := composeContent(req); ok {
		t.Fatal("expected ok=false when nothing selected maps to a compose service")
	}
}

func TestDockerfileContentIncludesGoVersion(t *testing.T) {
	content := dockerfileContent("1.25")
	if !strings.Contains(content, "golang:1.25-alpine") {
		t.Errorf("dockerfile content missing pinned go version:\n%s", content)
	}
}

func TestReadGoVersion(t *testing.T) {
	dir := t.TempDir()
	content := "module github.com/example/proj\n\ngo 1.23.4\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if got, want := readGoVersion(dir), "1.23"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReadGoVersionMissingFallsBack(t *testing.T) {
	dir := t.TempDir()
	if got, want := readGoVersion(dir), "1.25"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
