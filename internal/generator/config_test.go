package generator

import (
	"path/filepath"
	"testing"
)

func TestConfigSetGetRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(configPathEnvVar, filepath.Join(dir, "config.json"))

	var cfg Config
	if err := cfg.Set("author", "Jane Doe"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cfg.Set("license", "mit"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Get("author") != "Jane Doe" || loaded.Get("license") != "mit" {
		t.Fatalf("expected round-tripped config, got %+v", loaded)
	}
}

func TestConfigSetInvalidLicense(t *testing.T) {
	var cfg Config
	if err := cfg.Set("license", "bsd"); err == nil {
		t.Fatal("expected an error for an invalid license value")
	}
}

func TestConfigSetUnknownKey(t *testing.T) {
	var cfg Config
	if err := cfg.Set("not-a-real-key", "x"); err == nil {
		t.Fatal("expected an error for an unknown config key")
	}
}

func TestLoadConfigMissingFileReturnsZeroValue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(configPathEnvVar, filepath.Join(dir, "does-not-exist.json"))

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.License != "" || cfg.Author != "" || cfg.ModulePrefix != "" {
		t.Fatalf("expected zero-value config, got %+v", cfg)
	}
}
