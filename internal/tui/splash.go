package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Version is injected at build time via ldflags:
//
//	go build -ldflags "-X github.com/koriebruh/Genitz/internal/tui.Version=v1.0.0"
var Version = "dev"

// splashGradient is the Neon Midnight color ramp: violet → indigo → sky-blue.
var splashGradient = []string{
	"#F5D0FE", "#E879F9", "#D946EF", "#C026D3",
	"#A855F7", "#8B5CF6", "#7C3AED", "#6D28D9",
	"#4F46E5", "#3B82F6", "#38BDF8", "#22D3EE",
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

// leftPad returns a left-padding string so the logo/tagline is centered in the terminal.
func leftPad(terminalWidth int) string {
	if terminalWidth <= splashLogoWidth+8 {
		return "  " // minimal padding for narrow terminals
	}
	pad := (terminalWidth - splashLogoWidth) / 2
	return strings.Repeat(" ", pad)
}

// splashRenderLogo renders the ASCII logo with a per-row color gradient.
// Block characters (█) use the main ramp stop; shade characters (░) use
// a slightly deeper stop for a subtle depth effect.
// If terminalWidth > 0, the logo is centered horizontally.
func splashRenderLogo(terminalWidth int) string {
	padding := leftPad(terminalWidth)
	var sb strings.Builder
	for i, row := range splashLogo {
		main := splashGradient[min(i, len(splashGradient)-1)]
		depth := splashGradient[min(i+2, len(splashGradient)-1)]
		sb.WriteString(padding)
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
func splashRenderTagline(terminalWidth int) string {
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
	runeLen := len([]rune(plain))

	padding := leftPad(terminalWidth)
	// Fine-tune: center tagline relative to logo itself
	finePad := max((splashLogoWidth-runeLen)/2, 0)

	var sb strings.Builder
	sb.WriteString(padding)
	sb.WriteString(strings.Repeat(" ", finePad))
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
func RenderHeader(terminalWidth int) string {
	var sb strings.Builder
	sb.WriteString(splashRenderLogo(terminalWidth))
	sb.WriteString(splashRenderTagline(terminalWidth))
	sb.WriteRune('\n')
	return sb.String()
}

// RenderHeaderCompact returns a single branded line without the ASCII art.
// Used when terminal height is between 18–27.
func RenderHeaderCompact() string {
	brand := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A855F7")).Bold(true).
		Render("GENITZ")
	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render("  go project initializer · " + Version)
	div := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2D1B69")).
		Render(strings.Repeat("━", splashLogoWidth))
	return "  " + brand + version + "\n" + div + "\n\n"
}
