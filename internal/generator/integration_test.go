package generator

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/koriebruh/Genitz/internal/tui"
)

// networkReachable does a quick, best-effort check for internet access so
// the deps-touching integration test below can skip cleanly instead of
// flaking in offline/sandboxed environments.
func networkReachable() bool {
	conn, err := net.DialTimeout("tcp", "proxy.golang.org:443", 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// TestIntegrationScaffoldWithoutDeps exercises the real init path — go mod
// init, go mod tidy, go fmt — against the actual go toolchain, with zero
// dependencies so it needs no network and stays fast enough to always run.
// This is the class of bug unit tests alone kept missing this session: the
// pieces that only fail once a real `go` binary actually touches disk.
func TestIntegrationScaffoldWithoutDeps(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	if err := CheckBinary("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	dir := t.TempDir()
	projectPath := filepath.Join(dir, "integration-app")

	req := Requirement{ProjectName: projectPath, PackageName: "integration-app"}
	targetPath, err := PrepareNewProject(req)
	if err != nil {
		t.Fatalf("PrepareNewProject: %v", err)
	}

	for _, step := range BuildInstallSteps(targetPath, req) {
		if err := step.Run(); err != nil {
			t.Fatalf("step %q failed: %v", step.Label, err)
		}
	}

	if _, err := os.Stat(filepath.Join(targetPath, "go.mod")); err != nil {
		t.Fatalf("expected go.mod to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetPath, "main.go")); err != nil {
		t.Fatalf("expected main.go to exist: %v", err)
	}
}

// TestIntegrationScaffoldWithDependency additionally exercises a real `go
// get` against the network — skipped if the network isn't reachable so it
// doesn't flake offline.
func TestIntegrationScaffoldWithDependency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	if err := CheckBinary("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	if !networkReachable() {
		t.Skip("network not reachable, skipping go get integration test")
	}

	redis, ok := tui.FindByID("redis")
	if !ok {
		t.Fatal("expected \"redis\" to exist in the registry")
	}

	dir := t.TempDir()
	projectPath := filepath.Join(dir, "integration-app-deps")

	req := Requirement{
		ProjectName: projectPath,
		PackageName: "integration-app-deps",
		Deps:        map[int]tui.Dependency{0: redis},
	}
	targetPath, err := PrepareNewProject(req)
	if err != nil {
		t.Fatalf("PrepareNewProject: %v", err)
	}

	for _, step := range BuildInstallSteps(targetPath, req) {
		if err := step.Run(); err != nil {
			t.Fatalf("step %q failed: %v", step.Label, err)
		}
	}

	if _, err := os.Stat(filepath.Join(targetPath, "go.mod")); err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	// The bare scaffolded main.go never imports redis, so go mod tidy
	// correctly strips it back out — this test's job is confirming `go get`
	// itself succeeded against the real module proxy, not that it lingers.
	if _, err := ListInstalled(targetPath); err != nil {
		t.Fatalf("ListInstalled: %v", err)
	}
}

// TestIntegrationRemoveDependency exercises the real BuildRemoveSteps path
// (go get pkg@none + go mod tidy) against an actually-imported dependency,
// so tidy can't silently strip it out from under the test the way it does
// in TestIntegrationScaffoldWithDependency.
func TestIntegrationRemoveDependency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	if err := CheckBinary("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	if !networkReachable() {
		t.Skip("network not reachable, skipping go get integration test")
	}

	redis, ok := tui.FindByID("redis")
	if !ok {
		t.Fatal("expected \"redis\" to exist in the registry")
	}

	dir := t.TempDir()
	projectPath := filepath.Join(dir, "integration-app-remove")

	req := Requirement{ProjectName: projectPath, PackageName: "integration-app-remove"}
	targetPath, err := PrepareNewProject(req)
	if err != nil {
		t.Fatalf("PrepareNewProject: %v", err)
	}
	for _, step := range BuildInstallSteps(targetPath, req) {
		if err := step.Run(); err != nil {
			t.Fatalf("init step %q failed: %v", step.Label, err)
		}
	}

	mainGo := "package main\n\nimport \"github.com/redis/go-redis/v9\"\n\nfunc main() {\n\t_ = redis.NewClient(nil)\n}\n"
	if err := os.WriteFile(filepath.Join(targetPath, "main.go"), []byte(mainGo), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	for _, step := range BuildAddSteps(targetPath, map[int]tui.Dependency{0: redis}, nil) {
		if err := step.Run(); err != nil {
			t.Fatalf("add step %q failed: %v", step.Label, err)
		}
	}

	before, err := ListInstalled(targetPath)
	if err != nil {
		t.Fatalf("ListInstalled before remove: %v", err)
	}
	if len(before) != 1 || before[0].ImportPath != redis.ImportPath {
		t.Fatalf("expected redis to be installed and imported, got %+v", before)
	}

	// Drop the import so `go get redis@none` + tidy can actually remove it
	// (still-referenced code would just get tidy'd back in).
	if err := os.WriteFile(filepath.Join(targetPath, "main.go"), []byte(bareMainGo), 0o644); err != nil {
		t.Fatalf("rewrite main.go: %v", err)
	}
	for _, step := range BuildRemoveSteps(targetPath, map[int]tui.Dependency{0: redis}) {
		if err := step.Run(); err != nil {
			t.Fatalf("remove step %q failed: %v", step.Label, err)
		}
	}

	after, err := ListInstalled(targetPath)
	if err != nil {
		t.Fatalf("ListInstalled after remove: %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("expected no direct deps left after remove, got %+v", after)
	}
}
