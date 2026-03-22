package tui

import "github.com/charmbracelet/lipgloss"

// ── Neon Midnight Palette ─────────────────────────────────────────────────────
// A high-contrast dark theme: violet brand, sky-blue accents, emerald success.
var (
	colorBrand     = lipgloss.Color("#8B5CF6") // violet     — brand identity
	colorAccent    = lipgloss.Color("#38BDF8") // sky-blue   — active / highlight
	colorHighlight = lipgloss.Color("#C084FC") // lavender   — cursor / selected
	colorDone      = lipgloss.Color("#34D399") // emerald    — completed / checked
	colorText      = lipgloss.Color("#F1F5F9") // near-white — primary text
	colorMuted     = lipgloss.Color("#64748B") // slate-500  — hints / meta
	colorSubtle    = lipgloss.Color("#94A3B8") // slate-400  — descriptions
	colorDivider   = lipgloss.Color("#334155") // slate-700  — visible dividers
	colorSurface   = lipgloss.Color("#1E293B") // slate-900  — badge / pill bg
	colorSelected  = lipgloss.Color("#E879F9") // fuchsia    — selected item
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

	// Panel sections — PanelLabel has a violet left-border accent
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
		Brand:       lipgloss.NewStyle().Foreground(colorBrand).Bold(true),
		Name:        lipgloss.NewStyle().Foreground(colorText).Bold(true),
		Selected:    lipgloss.NewStyle().Foreground(colorSelected).Bold(true),
		Description: lipgloss.NewStyle().Foreground(colorSubtle).Italic(true),

		Cursor:   lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		Checkbox: lipgloss.NewStyle().Foreground(colorDone).Bold(true),

		// Active step: sky-blue pill (dark text on bright bg)
		StepActive: lipgloss.NewStyle().
			Foreground(colorSurface).
			Background(colorAccent).
			Bold(true).
			Padding(0, 1),
		// Done step: emerald text
		StepDone: lipgloss.NewStyle().Foreground(colorDone),
		// Pending step: muted slate
		StepPending: lipgloss.NewStyle().Foreground(colorMuted),
		// Separator arrow between steps
		StepSep: lipgloss.NewStyle().Foreground(colorDivider),

		// Panel label with a violet left-border accent bar
		PanelLabel: lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(colorBrand).
			PaddingLeft(1),
		PanelHint: lipgloss.NewStyle().Foreground(colorMuted).Italic(true),

		InputPrompt: lipgloss.NewStyle().Foreground(colorDone).Bold(true),
		InputNote:   lipgloss.NewStyle().Foreground(colorMuted).Italic(true),

		// Key badges: text on a distinct background for visual separation without breaking line height
		KeyBadge: lipgloss.NewStyle().
			Foreground(colorAccent).
			Background(colorSurface).
			Bold(true).
			Padding(0, 1),
		KeyHint: lipgloss.NewStyle().Foreground(colorMuted),

		Divider:   lipgloss.NewStyle().Foreground(colorDivider),
		Container: lipgloss.NewStyle().Padding(0, 4),
	}
}
