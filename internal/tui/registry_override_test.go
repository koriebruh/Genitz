package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUserRegistryOverridesByID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	content := `[{"ID":"fiber","Name":"Fiber (internal fork)","Category":"framework","ImportPath":"github.com/us/fiber-fork","Description":"internal fork"}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(registryOverrideEnvVar, path)

	base := []Dependency{
		{ID: "fiber", Name: "Fiber", Category: CatFramework, ImportPath: "github.com/gofiber/fiber/v3"},
		{ID: "gin", Name: "Gin", Category: CatFramework, ImportPath: "github.com/gin-gonic/gin"},
	}
	merged := loadUserRegistry(base)

	if len(merged) != 2 {
		t.Fatalf("expected override to replace fiber in place (still 2 entries), got %d: %+v", len(merged), merged)
	}
	var fiber Dependency
	for _, d := range merged {
		if d.ID == "fiber" {
			fiber = d
		}
	}
	if fiber.ImportPath != "github.com/us/fiber-fork" {
		t.Fatalf("expected the override's import path to win, got %q", fiber.ImportPath)
	}
}

func TestLoadUserRegistryAppendsNewID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	content := `[{"ID":"internal-tool","Name":"Internal Tool","Category":"framework","ImportPath":"example.com/internal-tool","Description":"team-internal"}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(registryOverrideEnvVar, path)

	base := []Dependency{{ID: "fiber", Name: "Fiber", Category: CatFramework, ImportPath: "github.com/gofiber/fiber/v3"}}
	merged := loadUserRegistry(base)

	if len(merged) != 2 {
		t.Fatalf("expected the new ID to be appended, got %d entries: %+v", len(merged), merged)
	}
}

func TestLoadUserRegistrySkipsUnknownCategory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	content := `[{"ID":"weird","Name":"Weird","Category":"not-a-real-category","ImportPath":"example.com/weird","Description":"x"}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(registryOverrideEnvVar, path)

	base := []Dependency{{ID: "fiber", Name: "Fiber", Category: CatFramework, ImportPath: "github.com/gofiber/fiber/v3"}}
	merged := loadUserRegistry(base)

	if len(merged) != 1 {
		t.Fatalf("expected the unknown-category entry to be skipped, got %d entries: %+v", len(merged), merged)
	}
}

func TestLoadUserRegistrySkipsMissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	content := `[{"ID":"","Name":"No ID","Category":"framework","ImportPath":"example.com/no-id","Description":"x"}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(registryOverrideEnvVar, path)

	base := []Dependency{{ID: "fiber", Name: "Fiber", Category: CatFramework, ImportPath: "github.com/gofiber/fiber/v3"}}
	merged := loadUserRegistry(base)

	if len(merged) != 1 {
		t.Fatalf("expected the missing-ID entry to be skipped, got %d entries: %+v", len(merged), merged)
	}
}

func TestLoadUserRegistryNoFileReturnsBaseUnchanged(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(registryOverrideEnvVar, filepath.Join(dir, "does-not-exist.json"))

	base := []Dependency{{ID: "fiber", Name: "Fiber", Category: CatFramework, ImportPath: "github.com/gofiber/fiber/v3"}}
	merged := loadUserRegistry(base)

	if len(merged) != 1 || merged[0].ID != "fiber" {
		t.Fatalf("expected base returned unchanged when the override file doesn't exist, got %+v", merged)
	}
}

func TestLoadUserRegistryInvalidJSONReturnsBaseUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(registryOverrideEnvVar, path)

	base := []Dependency{{ID: "fiber", Name: "Fiber", Category: CatFramework, ImportPath: "github.com/gofiber/fiber/v3"}}
	merged := loadUserRegistry(base)

	if len(merged) != 1 || merged[0].ID != "fiber" {
		t.Fatalf("expected base returned unchanged on invalid JSON, got %+v", merged)
	}
}
