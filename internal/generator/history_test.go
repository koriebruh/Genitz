package generator

import (
	"path/filepath"
	"testing"
)

func TestRecordAndReadHistory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(historyPathEnvVar, filepath.Join(dir, "history.jsonl"))

	RecordHistory("init", "my-app", "2 deps")
	RecordHistory("add", ".", "1 deps")

	entries, err := ReadHistory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(entries), entries)
	}
	if entries[0].Command != "init" || entries[1].Command != "add" {
		t.Fatalf("expected entries in append order, got %+v", entries)
	}
}

func TestReadHistoryMissingFileReturnsNil(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(historyPathEnvVar, filepath.Join(dir, "does-not-exist.jsonl"))

	entries, err := ReadHistory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Fatalf("expected nil entries, got %+v", entries)
	}
}
