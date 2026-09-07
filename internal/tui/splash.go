package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// splashGradient is the color ramp applied row-by-row across the logo — a
// blue-to-orange sweep matching the same two-color brutalist system as
// the rest of the TUI (styles.go's colorPrimary/colorAccent), so the logo
// no longer clashes with a purple/pink/cyan palette nothing else uses.
var splashGradient = []string{
	"#DBEAFE", "#BFDBFE", "#93C5FD", "#60A5FA",
	"#3B82F6", "#2563EB", "#1D4ED8", "#F97316",
	"#FB923C", "#FDBA74", "#FED7AA", "#FFEDD5",
}

var splashLogo = []string{
	`   █████████                      █████  █████                `,
	`  ███░░░░░███                    ░░███  ░░███                 `,
	` ███     ░░░   ██████  ████████   ░███  █████    █████████   `,
	`░███          ███░░███░░███░░███  ░███ ░░░███░    ░█░░░░███   `,
	`░███    █████░███████  ░███ ░███  ░███   ░███     ░   ███░    `,
	`░░███  ░░███ ░███░░░   ░███ ░███  ░███   ░███ ███   ███░  █   `,
	` ░░█████████ ░░██████  ████ █████ █████  ░░█████   █████████  `,
	`  ░░░░░░░░░   ░░░░░░  ░░░░ ░░░░░ ░░░░░   ░░░░░   ░░░░░░░░░   `,
}

// splashLogoWidth is the character width of each logo row, used for centering.
const splashLogoWidth = 64

// splashRenderLogo renders the ASCII logo with a per-row color gradient.
// Block characters (█) use the main ramp stop; shade characters (░) use
// a slightly deeper stop for a subtle depth effect.
func splashRenderLogo() string {
	var sb strings.Builder
	for i, row := range splashLogo {
		main := splashGradient[min(i, len(splashGradient)-1)]
		depth := splashGradient[min(i+2, len(splashGradient)-1)]
		for _, ch := range row {
			switch ch {
			case ' ':
				sb.WriteRune(' ')
			case '░':
				sb.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(depth)).
					Render(string(ch)))
			default:
				sb.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(main)).
					Bold(true).
					Render(string(ch)))
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// splashRenderTagline renders the centered stylized tagline below the logo.
func splashRenderTagline() string {
	type seg struct {
		text  string
		color string
		faint bool
		bold  bool
	}
	segs := []seg{
		{"░", "#3B82F6", true, false},
		{"ゲ", "#60A5FA", false, true},
		{"ェ", "#3B82F6", false, true},
		{"ニ", "#2563EB", false, true},
		{"ト", "#F97316", false, true},
		{"░", "#3B82F6", true, false},
		{"  ɢᴏ ɪɴɪᴛɪᴀʟɪᴢᴇʀ ᴘʀᴏᴊᴇᴄᴛ   ", "#93C5FD", false, false},
		{"░", "#3B82F6", true, false},
		{"ズ", "#F97316", false, true},
		{"░", "#3B82F6", true, false},
	}

	plain := ""
	for _, s := range segs {
		plain += s.text
	}
	pad := max((splashLogoWidth-len([]rune(plain)))/2, 0)

	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", pad))
	for _, s := range segs {
		st := lipgloss.NewStyle().Foreground(lipgloss.Color(s.color))
		if s.bold {
			st = st.Bold(true)
		}
		if s.faint {
			st = st.Faint(true)
		}
		sb.WriteString(st.Render(s.text))
	}
	sb.WriteRune('\n')
	return sb.String()
}

// RenderHeader returns the full ASCII logo + tagline.
// Used when terminal height >= 28.
func RenderHeader() string {
	var sb strings.Builder
	sb.WriteString(splashRenderLogo())
	sb.WriteString(splashRenderTagline())
	sb.WriteRune('\n')
	return sb.String()
}

// RenderHeaderCompact returns a single branded line without the ASCII art,
// truncated to fit terminal width w. The caller draws its own divider below
// it, so this only ever emits one line — used when the terminal is too
// short or too narrow for the full logo.
func RenderHeaderCompact(w int) string {
	brand := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).Bold(true).
		Render("GENITZ")
	sub := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render("  go project initializer · v0.1.0")
	return lipgloss.NewStyle().MaxWidth(w).Render("  "+brand+sub) + "\n\n"
}
