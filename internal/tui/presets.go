package tui

import "strings"

// Preset is a named bundle of registry dependency IDs a user can apply in
// one action instead of hand-picking every package for a common shape of
// project. IDs are verified to exist in registry.json — see FindByID.
type Preset struct {
	ID          string
	Name        string
	Description string
	DepIDs      []string
}

// Presets are the built-in starter bundles, offered from the StepDeps
// picker (press "p") and via the --preset flag.
var Presets = []Preset{
	{
		ID:          "web-api",
		Name:        "Web API",
		Description: "Fiber + GORM + Viper + Zap",
		DepIDs:      []string{"fiber", "gorm", "viper", "zap"},
	},
	{
		ID:          "grpc-service",
		Name:        "gRPC Service",
		Description: "gRPC + Protobuf + Viper + Zap",
		DepIDs:      []string{"grpc", "protobuf", "viper", "zap"},
	},
	{
		ID:          "auth-service",
		Name:        "Auth Service",
		Description: "Gin + golang-jwt + Viper",
		DepIDs:      []string{"gin", "jwt", "viper"},
	},
	{
		ID:          "cli-tool",
		Name:        "CLI Tool",
		Description: "Cobra + Viper + Zap",
		DepIDs:      []string{"cobra", "viper", "zap"},
	},
}

// FindPreset returns the preset with the given ID (built-in or
// user-saved), if any.
func FindPreset(id string) (Preset, bool) {
	for _, p := range AllPresets() {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}

// renderPresetOverlay renders the preset bundle list shown over StepDeps.
func (m *Model) renderPresetOverlay() string {
	var b strings.Builder
	b.WriteString(styles.PanelLabel.Render("PRESETS") + "\n")
	b.WriteString(styles.PanelHint.Render("apply a starter bundle — adds to your current picks, doesn't replace them") + "\n\n")

	for i, p := range AllPresets() {
		cursor := "  "
		name := styles.Name.Render(p.Name)
		if m.PresetCursor == i {
			cursor = styles.Cursor.Render("▶ ")
			name = styles.Selected.Render(p.Name)
		}
		b.WriteString(cursor + name + "\n")
		b.WriteString("    " + styles.Description.Render(p.Description) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderKeyHints([]keyHint{
		{"↑↓ / jk", "navigate"},
		{"enter", "apply"},
		{"esc", "cancel"},
	}))
	return b.String()
}

// applyPreset unions preset's dependencies into m.Chosen — it never removes
// an existing selection, so applying a second preset (or hand-picking
// afterward) only ever adds.
func (m *Model) applyPreset(preset Preset) {
	byID := make(map[string]int, len(m.Registry))
	for i, dep := range m.Registry {
		byID[dep.ID] = i
	}
	for _, id := range preset.DepIDs {
		if idx, ok := byID[id]; ok {
			m.Chosen[idx] = m.Registry[idx]
		}
	}
}
