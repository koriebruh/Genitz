package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleDockerKeysToggle(t *testing.T) {
	m := InitialModel()
	m.Step = StepDocker

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if !m.IncludeDocker {
		t.Fatal("expected \"y\" to set IncludeDocker true")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if m.IncludeDocker {
		t.Fatal("expected \"n\" to set IncludeDocker false")
	}
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !m.IncludeDocker {
		t.Fatal("expected space to toggle IncludeDocker to true")
	}
}

func TestHandleDockerKeysNavigation(t *testing.T) {
	m := InitialModel()
	m.Step = StepDocker

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Step != StepExtras {
		t.Fatalf("expected enter to advance to StepExtras, got %v", m.Step)
	}

	m.Step = StepDocker
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.Step != StepDeps {
		t.Fatalf("expected \"b\" to go back to StepDeps, got %v", m.Step)
	}
}
