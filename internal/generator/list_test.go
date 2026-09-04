package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListInstalledParsesRequireBlockSkipsIndirect(t *testing.T) {
	dir := t.TempDir()
	content := `module example.com/app

go 1.25.0

require (
	github.com/redis/go-redis/v9 v9.5.1
	golang.org/x/sys v0.41.0 // indirect
)

require github.com/spf13/viper v1.19.0
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps, err := ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 direct deps (indirect skipped), got %d: %+v", len(deps), deps)
	}

	byPath := make(map[string]InstalledDep, len(deps))
	for _, d := range deps {
		byPath[d.ImportPath] = d
	}

	redis, ok := byPath["github.com/redis/go-redis/v9"]
	if !ok {
		t.Fatal("expected redis import path present")
	}
	if !redis.Managed || redis.Name == "" {
		t.Fatalf("expected redis to resolve against the registry, got %+v", redis)
	}

	viper, ok := byPath["github.com/spf13/viper"]
	if !ok {
		t.Fatal("expected single-line require for viper to be parsed")
	}
	if viper.Version != "v1.19.0" {
		t.Fatalf("expected viper version v1.19.0, got %q", viper.Version)
	}
}

func TestListInstalledUnmanagedImport(t *testing.T) {
	dir := t.TempDir()
	content := "module example.com/app\n\ngo 1.25.0\n\nrequire example.com/not-in-registry v1.0.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	deps, err := ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 || deps[0].Managed {
		t.Fatalf("expected one unmanaged entry, got %+v", deps)
	}
}

func TestListInstalledMissingGoMod(t *testing.T) {
	dir := t.TempDir()
	if _, err := ListInstalled(dir); err == nil {
		t.Fatal("expected an error for missing go.mod")
	}
}
