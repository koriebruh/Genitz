package tui

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

const splashLogoWidth = 64

var (
	splashVersionLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)
	splashVersionValue = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22D3EE")).
		Bold(true)
	splashDivider = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2D1B69"))
	splashHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151"))
	splashContainer = lipgloss.NewStyle().
		Padding(1, 3)
)

func splashFetchGoVersion() string {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return "not found — make sure Go is installed"
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "go version ")
}

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

func splashRenderContent(goVersion string) string {
	var sb strings.Builder
	sb.WriteString(splashRenderLogo())
	sb.WriteString(splashRenderTagline())
	sb.WriteString("\n" + splashDivider.Render(strings.Repeat("━", splashLogoWidth)) + "\n\n")
	sb.WriteString(
		splashVersionLabel.Render("  Go version  ") +
			splashVersionValue.Render(goVersion) + "\n\n",
	)
	sb.WriteString(splashHint.Render("  q · esc · enter  continue") + "\n")
	return splashContainer.Render(sb.String())
}

type splashModel struct {
	goVersion string
	ready     bool
}

type splashVersionMsg string

func (m splashModel) Init() tea.Cmd {
	return func() tea.Msg {
		return splashVersionMsg(splashFetchGoVersion())
	}
}

func (m splashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case splashVersionMsg:
		m.goVersion = string(msg)
		m.ready = true
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m splashModel) View() string {
	if !m.ready {
		return "\n  Loading...\n"
	}
	return splashRenderContent(m.goVersion)
}

// RenderSplashView fetches the Go version and renders the splash screen.
// Use this for non-interactive / static rendering contexts.
func RenderSplashView() string {
	return splashRenderContent(splashFetchGoVersion())
}
