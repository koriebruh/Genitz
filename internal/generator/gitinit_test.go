package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunGitInit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Ensure commit identity exists even in a bare CI environment.
	cmd := exec.Command("git", "config", "--global", "user.email")
	if out, _ := cmd.Output(); len(out) == 0 {
		t.Setenv("GIT_AUTHOR_NAME", "Test")
		t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
		t.Setenv("GIT_COMMITTER_NAME", "Test")
		t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	}

	if err := runGitInit(dir); err != nil {
		t.Fatalf("runGitInit: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Error(".git directory not created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err != nil {
		t.Error(".gitignore not created")
	}

	out, err := exec.Command("git", "-C", dir, "log", "--oneline").Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected at least one commit")
	}
}

func TestSuggestedPublishCommand(t *testing.T) {
	if got := SuggestedPublishCommand(); got == "" {
		t.Fatal("expected a non-empty suggested command")
	}
}
