package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBeginInstallWithNoBuildStepsQuits(t *testing.T) {
	m := InitialModel()
	cmd := m.beginInstall()
	if !m.Done {
		t.Fatal("expected Done to be true when BuildSteps is nil")
	}
	if cmd == nil {
		t.Fatal("expected a tea.Quit cmd")
	}
}

func TestHandleStepDoneAdvancesOnSuccess(t *testing.T) {
	m := InitialModel()
	m.InstallSteps = []InstallStep{
		{Label: "one", Run: func() error { return nil }},
		{Label: "two", Run: func() error { return nil }},
	}
	m.InstallIndex = 0

	cmd, handled := m.handleStepDone(stepDoneMsg{err: nil})
	if !handled {
		t.Fatal("expected handleStepDone to report handled")
	}
	if m.InstallIndex != 1 {
		t.Fatalf("expected InstallIndex to advance to 1, got %d", m.InstallIndex)
	}
	if cmd == nil {
		t.Fatal("expected a cmd to run the next step")
	}
	if m.Done {
		t.Fatal("expected Done to still be false with a step remaining")
	}
}

func TestHandleStepDoneFinishesOnLastStep(t *testing.T) {
	m := InitialModel()
	m.InstallSteps = []InstallStep{{Label: "only", Run: func() error { return nil }}}
	m.InstallIndex = 0

	m.handleStepDone(stepDoneMsg{err: nil})
	if !m.Done {
		t.Fatal("expected Done to be true after the last step succeeds")
	}
}

func TestHandleStepDoneRecordsFailure(t *testing.T) {
	m := InitialModel()
	m.InstallSteps = []InstallStep{
		{Label: "one", Run: func() error { return nil }},
		{Label: "two", Run: func() error { return nil }},
	}
	m.InstallIndex = 0

	wantErr := errors.New("boom")
	cmd, handled := m.handleStepDone(stepDoneMsg{err: wantErr})
	if !handled {
		t.Fatal("expected handleStepDone to report handled")
	}
	if m.InstallErr != wantErr {
		t.Fatalf("expected InstallErr to be set to %v, got %v", wantErr, m.InstallErr)
	}
	if m.InstallIndex != 0 {
		t.Fatalf("expected InstallIndex to stay at the failed step, got %d", m.InstallIndex)
	}
	if cmd == nil {
		t.Fatal("expected a tea.Quit cmd on failure")
	}
	// Confirm it really is tea.Quit-shaped by invoking it — tea.Quit
	// returns a tea.QuitMsg.
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected the returned cmd to produce a tea.QuitMsg")
	}
}
