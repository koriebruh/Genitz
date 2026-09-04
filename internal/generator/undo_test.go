package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotAndUndoRestoresGoMod(t *testing.T) {
	undoRoot := t.TempDir()
	t.Setenv(undoRootEnvVar, undoRoot)

	projectDir := t.TempDir()
	original := "module example.com/app\n\ngo 1.25.0\n\nrequire github.com/redis/go-redis/v9 v9.5.1\n"
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "go.sum"), []byte("fake sum data\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SnapshotForUndo(projectDir); err != nil {
		t.Fatalf("SnapshotForUndo: %v", err)
	}

	// Simulate an add/remove mutating go.mod after the snapshot.
	mutated := "module example.com/app\n\ngo 1.25.0\n\nrequire github.com/gofiber/fiber/v3 v3.0.0\n"
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(mutated), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Undo(projectDir); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	restored, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != original {
		t.Fatalf("expected go.mod restored to original content, got:\n%s", restored)
	}
}

func TestUndoNothingToUndo(t *testing.T) {
	undoRoot := t.TempDir()
	t.Setenv(undoRootEnvVar, undoRoot)

	projectDir := t.TempDir()
	if err := Undo(projectDir); err == nil {
		t.Fatal("expected an error when no snapshot exists for this project")
	}
}

func TestSnapshotIsPerDirectory(t *testing.T) {
	undoRoot := t.TempDir()
	t.Setenv(undoRootEnvVar, undoRoot)

	projectA := t.TempDir()
	projectB := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectA, "go.mod"), []byte("module a\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SnapshotForUndo(projectA); err != nil {
		t.Fatalf("SnapshotForUndo(A): %v", err)
	}

	// projectB never had a snapshot taken — Undo must fail for it
	// independently of projectA's snapshot existing.
	if err := os.WriteFile(filepath.Join(projectB, "go.mod"), []byte("module b\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Undo(projectB); err == nil {
		t.Fatal("expected Undo(projectB) to fail — no snapshot was ever taken for it")
	}
}

func TestUndoRemovesSnapshotAfterRestoring(t *testing.T) {
	undoRoot := t.TempDir()
	t.Setenv(undoRootEnvVar, undoRoot)

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module app\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SnapshotForUndo(projectDir); err != nil {
		t.Fatalf("SnapshotForUndo: %v", err)
	}
	if err := Undo(projectDir); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	// A second Undo with no new snapshot in between must fail — single-
	// level LIFO, not a history stack.
	if err := Undo(projectDir); err == nil {
		t.Fatal("expected a second consecutive Undo to fail (snapshot consumed by the first)")
	}
}
