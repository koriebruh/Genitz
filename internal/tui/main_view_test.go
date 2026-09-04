package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestFolderStepValidation(t *testing.T) {
	m := InitialModel()

	// Enter on an empty folder name stays on StepFolder with an error set.
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Step != StepFolder {
		t.Fatalf("expected to stay on StepFolder, got %v", m.Step)
	}
	if m.FolderErr == "" {
		t.Fatal("expected FolderErr to be set for an empty folder name")
	}

	// A valid folder name advances to StepPackage.
	m.FolderInput.SetValue("my-app")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Step != StepPackage {
		t.Fatalf("expected StepPackage, got %v", m.Step)
	}
}

func TestWindowSizeMsgUpdatesDimensions(t *testing.T) {
	m := InitialModel()
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if m.Width != 100 || m.Height != 40 {
		t.Fatalf("expected Width=100 Height=40, got Width=%d Height=%d", m.Width, m.Height)
	}
}

func TestDepsStepCursorBounds(t *testing.T) {
	m := InitialModel()
	m.Step = StepDeps
	m.Cursor = 0

	// "up" at the top stays at 0.
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Cursor != 0 {
		t.Fatalf("expected cursor to stay at 0, got %d", m.Cursor)
	}

	// "down" moves forward without panicking or going negative.
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor < 0 {
		t.Fatalf("cursor went negative: %d", m.Cursor)
	}
}

func TestPresetOverlayOpenApplyAddsDeps(t *testing.T) {
	m := InitialModel()
	m.Step = StepDeps

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if !m.PresetOverlayOpen {
		t.Fatal("expected \"p\" to open the preset overlay")
	}

	before := len(m.Chosen)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.PresetOverlayOpen {
		t.Fatal("expected enter to close the preset overlay")
	}
	if len(m.Chosen) <= before {
		t.Fatalf("expected applying a preset to add dependencies, had %d now %d", before, len(m.Chosen))
	}
}

func TestPresetOverlayEscCancelsWithoutApplying(t *testing.T) {
	m := InitialModel()
	m.Step = StepDeps

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	before := len(m.Chosen)
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.PresetOverlayOpen {
		t.Fatal("expected esc to close the preset overlay")
	}
	if len(m.Chosen) != before {
		t.Fatalf("expected esc not to apply a preset, had %d now %d", before, len(m.Chosen))
	}
}

func TestRemoveModelScopesRegistryToInstalled(t *testing.T) {
	installed := []Dependency{
		{ID: "redis", Name: "Redis", Category: CatCache, ImportPath: "github.com/redis/go-redis/v9"},
	}
	m := RemoveModel("example.com/app", installed)
	if !m.RemoveMode {
		t.Fatal("expected RemoveMode to be true")
	}
	if len(m.Registry) != 1 || m.Registry[0].ID != "redis" {
		t.Fatalf("expected Registry scoped to the single installed dep, got %+v", m.Registry)
	}
}
