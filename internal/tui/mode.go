package tui

// Mode picks which wizard flow the CLI runs: a fresh project (no go.mod
// in the current directory yet) or adding dependencies to an existing one.
type Mode int

const (
	ModeInit Mode = iota
	ModeAdd
)

// Step represents a wizard step in the TUI flow.
type Step int

const (
	StepFolder Step = iota
	StepPackage
	StepDeps
	StepReview
	StepInstalling
)
