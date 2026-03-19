package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// splashGradient is the color ramp applied row-by-row across the logo.
var splashGradient = []string{
	"#F5D0FE", "#F0ABFC", "#E879F9", "#D946EF",
	"#C026D3", "#A855F7", "#9333EA", "#7C3AED",
	"#6D28D9", "#4F46E5", "#67E8F9", "#22D3EE",
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
		{"░", "#4F46E5", true, false},
		{"ゲ", "#F0ABFC", false, true},
		{"ェ", "#E879F9", false, true},
		{"ニ", "#D946EF", false, true},
		{"ト", "#C026D3", false, true},
		{"░", "#4F46E5", true, false},
		{"  ɢᴏ ɪɴɪᴛɪᴀʟɪᴢᴇʀ ᴘʀᴏᴊᴇᴄᴛ   ", "#818CF8", false, false},
		{"░", "#4F46E5", true, false},
		{"ズ", "#9333EA", false, true},
		{"░", "#4F46E5", true, false},
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

// RenderHeaderCompact returns a single branded line without the ASCII art.
// Used when terminal height is between 20–27.
func RenderHeaderCompact() string {
	brand := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A855F7")).Bold(true).
		Render("GENITZ")
	sub := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render("  go project initializer · v0.1.0")
	div := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2D1B69")).
		Render(strings.Repeat("━", splashLogoWidth))
	return "  " + brand + sub + "\n" + div + "\n\n"
}
