package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — blue/orange/white, consistent across all panels. The
// gradient logo in splash.go is the one deliberate exception (kept as-is
// per explicit request) — everywhere else uses exactly two accent colors
// with one fixed meaning each, not a rotating cast of colors picked per
// screen.
var (
	colorPrimary  = lipgloss.Color("#3B82F6") // blue    — brand / active / structural (border, stamps)
	colorAccent   = lipgloss.Color("#F97316") // orange  — done / confirmed / keyboard hints
	colorDone     = lipgloss.Color("#F97316") // orange  — same accent, "completed" meaning
	colorMuted    = lipgloss.Color("#9CA3AF") // gray    — hints / descriptions
	colorText     = lipgloss.Color("#F8FAFC") // white   — primary text
	colorSelected = lipgloss.Color("#F97316") // orange  — cursor-highlighted text
	colorDivider  = lipgloss.Color("#3B82F6") // blue    — border / inner rules / step sep
	colorDark     = lipgloss.Color("#0B1220") // near-black — stamp foreground text
)

// uiStyles groups all Lipgloss styles used in the TUI.
type uiStyles struct {
	// Text roles
	Brand       lipgloss.Style
	Name        lipgloss.Style
	Selected    lipgloss.Style
	Description lipgloss.Style

	// Interactive elements
	Cursor   lipgloss.Style
	Checkbox lipgloss.Style

	// Step nav bar
	StepActive  lipgloss.Style
	StepDone    lipgloss.Style
	StepPending lipgloss.Style
	StepSep     lipgloss.Style

	// Panel sections
	PanelLabel lipgloss.Style
	PanelHint  lipgloss.Style

	// Input area
	InputPrompt lipgloss.Style
	InputNote   lipgloss.Style

	// Footer key hints
	KeyBadge lipgloss.Style
	KeyHint  lipgloss.Style

	// Layout
	Divider   lipgloss.Style
	Container lipgloss.Style
}

var styles = newUIStyles()

// spinnerStyle colors the install-progress spinner; installFailStyle marks
// a failed install step.
var (
	spinnerStyle     = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	installFailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
)

// newUIStyles — full brutalism: a real bordered Container (see renderFrame
// in main_view.go) instead of an ad-hoc repeated-character divider, plus a
// disciplined two-accent palette. colorPrimary (blue) marks
// "active/selected/labeled" structural state — cursor, active step, panel
// labels, the border itself. colorAccent/colorDone (orange, same value)
// marks "done/confirmed/positive" — checked boxes, completed steps, the
// input prompt, keyboard-shortcut chips. Every other screen reuses these
// same two roles rather than introducing a third or fourth color.
func newUIStyles() uiStyles {
	stamp := lipgloss.NewStyle().Foreground(colorDark).Background(colorPrimary).Bold(true)

	return uiStyles{
		Brand:       lipgloss.NewStyle().Foreground(colorPrimary).Bold(true),
		Name:        lipgloss.NewStyle().Foreground(colorText).Bold(true),
		Selected:    lipgloss.NewStyle().Foreground(colorSelected).Bold(true),
		Description: lipgloss.NewStyle().Foreground(colorMuted),

		Cursor:   stamp,
		Checkbox: lipgloss.NewStyle().Foreground(colorDark).Background(colorDone).Bold(true),

		StepActive:  stamp,
		StepDone:    lipgloss.NewStyle().Foreground(colorDone),
		StepPending: lipgloss.NewStyle().Foreground(colorMuted),
		StepSep:     lipgloss.NewStyle().Foreground(colorDivider),

		PanelLabel: stamp.Padding(0, 1),
		PanelHint:  lipgloss.NewStyle().Foreground(colorMuted),

		InputPrompt: lipgloss.NewStyle().Foreground(colorDone).Bold(true),
		InputNote:   lipgloss.NewStyle().Foreground(colorMuted),

		KeyBadge: lipgloss.NewStyle().
			Foreground(colorDark).
			Background(colorAccent).
			Bold(true).
			Padding(0, 1),
		KeyHint: lipgloss.NewStyle().Foreground(colorMuted),

		Divider:   lipgloss.NewStyle().Foreground(colorPrimary).Bold(true),
		Container: lipgloss.NewStyle().Padding(0, 3),
	}
}
