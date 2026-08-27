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
