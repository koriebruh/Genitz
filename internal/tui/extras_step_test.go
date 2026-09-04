package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleExtrasKeysToggleCheckbox(t *testing.T) {
	m := InitialModel()
	m.Step = StepExtras
	m.ExtrasCursor = 0 // "GitHub Actions CI"

	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !m.IncludeCI {
		t.Fatal("expected space on the CI row to set IncludeCI true")
	}
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.IncludeCI {
		t.Fatal("expected a second space to toggle IncludeCI back to false")
	}
}

func TestHandleExtrasKeysLicenseCycle(t *testing.T) {
	m := InitialModel()
	m.Step = StepExtras
	m.ExtrasCursor = len(extraItems) // the trailing License row

	if m.LicenseChoice != "" {
		t.Fatalf("expected default LicenseChoice to be empty, got %q", m.LicenseChoice)
	}
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.LicenseChoice != "mit" {
		t.Fatalf("expected first cycle to land on mit, got %q", m.LicenseChoice)
	}
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.LicenseChoice != "apache-2.0" {
		t.Fatalf("expected second cycle to land on apache-2.0, got %q", m.LicenseChoice)
	}
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.LicenseChoice != "" {
		t.Fatalf("expected third cycle to wrap back to none, got %q", m.LicenseChoice)
	}
}

func TestHandleExtrasKeysCursorBounds(t *testing.T) {
	m := InitialModel()
	m.Step = StepExtras
	m.ExtrasCursor = 0

	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.ExtrasCursor != 0 {
		t.Fatalf("expected cursor to stay at 0, got %d", m.ExtrasCursor)
	}

	last := len(extraItems) // the License row is the last valid position
	for i := 0; i < last+5; i++ {
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.ExtrasCursor != last {
		t.Fatalf("expected cursor to clamp at %d, got %d", last, m.ExtrasCursor)
	}
}
