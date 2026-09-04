package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const presetOverrideEnvVar = "GENITZ_PRESET_OVERRIDE"

func userPresetPath() string {
	if p := os.Getenv(presetOverrideEnvVar); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "genitz", "presets.json")
}

func loadUserPresets() []Preset {
	path := userPresetPath()
	if path == "" {
		return nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var presets []Preset
	if err := json.Unmarshal(content, &presets); err != nil {
		return nil
	}
	return presets
}

// AllPresets returns the built-in Presets plus any user-saved ones (see
// SavePreset) — a user preset with the same ID as a built-in overrides it
// in place, same merge shape as loadUserRegistry.
func AllPresets() []Preset {
	user := loadUserPresets()
	if len(user) == 0 {
		return Presets
	}

	merged := make([]Preset, len(Presets))
	copy(merged, Presets)
	byID := make(map[string]int, len(merged))
	for i, p := range merged {
		byID[p.ID] = i
	}
	for _, p := range user {
		if idx, ok := byID[p.ID]; ok {
			merged[idx] = p
		} else {
			byID[p.ID] = len(merged)
			merged = append(merged, p)
		}
	}
	return merged
}

// SavePreset appends preset to the user preset file (overwriting any
// existing entry with the same ID), creating the file/directory if needed —
// backs `genitz preset save`.
func SavePreset(preset Preset) error {
	path := userPresetPath()
	if path == "" {
		return errors.New("could not determine user config directory")
	}

	presets := loadUserPresets()
	replaced := false
	for i, p := range presets {
		if p.ID == preset.ID {
			presets[i] = preset
			replaced = true
			break
		}
	}
	if !replaced {
		presets = append(presets, preset)
	}

	content, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

// importPresetHTTPTimeout bounds how long a `genitz preset import` fetch
// can hang — a stalled remote shouldn't hang the CLI indefinitely.
const importPresetHTTPTimeout = 10 * time.Second

// ImportPresetFromURL fetches a single Preset (JSON object, not an array)
// from url and saves it via SavePreset — same trust model as the user
// running `curl <url>` themselves (their own explicit input), but the
// fetched JSON is still validated against the same required-field and
// DepIDs-must-resolve checks as any other preset source before being
// persisted, so a malformed or malicious response can't silently corrupt
// the local preset file.
func ImportPresetFromURL(url string) (Preset, error) {
	client := &http.Client{Timeout: importPresetHTTPTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return Preset{}, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Preset{}, fmt.Errorf("fetch %s: unexpected status %s", url, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MiB cap — a preset is a few KB at most
	if err != nil {
		return Preset{}, fmt.Errorf("read response from %s: %w", url, err)
	}

	var preset Preset
	if err := json.Unmarshal(body, &preset); err != nil {
		return Preset{}, fmt.Errorf("%s did not return a valid preset JSON object: %w", url, err)
	}
	if preset.ID == "" || preset.Name == "" || len(preset.DepIDs) == 0 {
		return Preset{}, fmt.Errorf("%s: preset is missing ID/Name/DepIDs", url)
	}

	var unresolved []string
	for _, id := range preset.DepIDs {
		if _, ok := FindByID(id); !ok {
			unresolved = append(unresolved, id)
		}
	}
	if len(unresolved) > 0 {
		return Preset{}, fmt.Errorf("%s: preset references unknown dependency IDs: %v", url, unresolved)
	}

	if err := SavePreset(preset); err != nil {
		return Preset{}, err
	}
	return preset, nil
}
