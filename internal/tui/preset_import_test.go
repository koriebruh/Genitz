package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestImportPresetFromURLSavesValidPreset(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(presetOverrideEnvVar, filepath.Join(dir, "presets.json"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Preset{
			ID:          "team-standard",
			Name:        "Team Standard",
			Description: "our team's default stack",
			DepIDs:      []string{"fiber", "zap"},
		})
	}))
	defer srv.Close()

	preset, err := ImportPresetFromURL(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.ID != "team-standard" || len(preset.DepIDs) != 2 {
		t.Fatalf("unexpected preset: %+v", preset)
	}

	found, ok := FindPreset("team-standard")
	if !ok {
		t.Fatal("expected the imported preset to be findable via FindPreset (persisted + merged)")
	}
	if found.Name != "Team Standard" {
		t.Fatalf("unexpected persisted preset: %+v", found)
	}
}

func TestImportPresetFromURLRejectsUnknownDepID(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(presetOverrideEnvVar, filepath.Join(dir, "presets.json"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Preset{
			ID:     "bad-preset",
			Name:   "Bad Preset",
			DepIDs: []string{"not-a-real-dependency-id"},
		})
	}))
	defer srv.Close()

	if _, err := ImportPresetFromURL(srv.URL); err == nil {
		t.Fatal("expected an error for a preset referencing an unknown dependency ID")
	}
	if _, err := os.Stat(filepath.Join(dir, "presets.json")); !os.IsNotExist(err) {
		t.Fatal("expected nothing to be saved when validation fails")
	}
}

func TestImportPresetFromURLRejectsMissingFields(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(presetOverrideEnvVar, filepath.Join(dir, "presets.json"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Preset{ID: "no-name-or-deps"})
	}))
	defer srv.Close()

	if _, err := ImportPresetFromURL(srv.URL); err == nil {
		t.Fatal("expected an error for a preset missing Name/DepIDs")
	}
}

func TestImportPresetFromURLRejectsNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := ImportPresetFromURL(srv.URL); err == nil {
		t.Fatal("expected an error for a non-200 response")
	}
}

func TestImportPresetFromURLRejectsInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	if _, err := ImportPresetFromURL(srv.URL); err == nil {
		t.Fatal("expected an error for an invalid JSON response")
	}
}
