package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — consistent across all panels.
var (
	colorPrimary  = lipgloss.Color("#A855F7") // purple  — brand
	colorAccent   = lipgloss.Color("#22D3EE") // cyan    — active / selected
	colorDone     = lipgloss.Color("#10B981") // green   — completed / checked
	colorMuted    = lipgloss.Color("#6B7280") // gray    — hints / descriptions
	colorText     = lipgloss.Color("#E5E7EB") // white   — primary text
	colorSelected = lipgloss.Color("#F0ABFC") // pink    — cursor highlight
	colorDivider  = lipgloss.Color("#2D1B69") // indigo  — dividers / step sep
	colorDark     = lipgloss.Color("#1F2937") // dark    — key badge bg
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

func newUIStyles() uiStyles {
	return uiStyles{
		Brand:       lipgloss.NewStyle().Foreground(colorPrimary).Bold(true),
		Name:        lipgloss.NewStyle().Foreground(colorText).Bold(true),
		Selected:    lipgloss.NewStyle().Foreground(colorSelected).Bold(true),
		Description: lipgloss.NewStyle().Foreground(colorMuted).Italic(true),

		Cursor:   lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		Checkbox: lipgloss.NewStyle().Foreground(colorDone).Bold(true),

		StepActive:  lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		StepDone:    lipgloss.NewStyle().Foreground(colorDone),
		StepPending: lipgloss.NewStyle().Foreground(colorMuted),
		StepSep:     lipgloss.NewStyle().Foreground(colorDivider),

		PanelLabel: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		PanelHint:  lipgloss.NewStyle().Foreground(colorMuted).Italic(true),

		InputPrompt: lipgloss.NewStyle().Foreground(colorDone).Bold(true),
		InputNote:   lipgloss.NewStyle().Foreground(colorMuted).Italic(true),

		KeyBadge: lipgloss.NewStyle().
			Foreground(colorAccent).
			Background(colorDark).
			Padding(0, 1),
		KeyHint: lipgloss.NewStyle().Foreground(colorMuted),

		Divider:   lipgloss.NewStyle().Foreground(colorDivider),
		Container: lipgloss.NewStyle().Padding(0, 3),
	}
}
