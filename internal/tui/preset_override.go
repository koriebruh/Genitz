package tui

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
