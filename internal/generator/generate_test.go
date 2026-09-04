package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadModulePath(t *testing.T) {
	dir := t.TempDir()
	content := "module github.com/example/proj\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadModulePath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "github.com/example/proj"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReadModulePathMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := ReadModulePath(dir); err == nil {
		t.Fatal("expected error for missing go.mod, got nil")
	}
}

func TestCheckPreconditionsRejectsExistingDestination(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "already-here")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := CheckPreconditions(Requirement{ProjectName: existing}); err == nil {
		t.Fatal("expected an error when the target directory already exists")
	}
}

func TestCheckPreconditionsDoesNotCreateAnything(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "not-yet-created")

	if _, err := CheckPreconditions(Requirement{ProjectName: target}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected CheckPreconditions not to create %q, stat error: %v", target, err)
	}
}

func TestCheckPreconditionsRejectsEmptyProjectName(t *testing.T) {
	if _, err := CheckPreconditions(Requirement{}); err == nil {
		t.Fatal("expected an error for an empty project name")
	}
}

func TestPrepareNewProjectRejectsExistingDestination(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "already-here")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := PrepareNewProject(Requirement{ProjectName: existing}); err == nil {
		t.Fatal("expected an error when the target directory already exists")
	}
}
