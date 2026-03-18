package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var gradient = []string{
	"#F5D0FE", "#F0ABFC", "#E879F9", "#D946EF",
	"#C026D3", "#A855F7", "#9333EA", "#7C3AED",
	"#6D28D9", "#4F46E5", "#67E8F9", "#22D3EE",
}

var logo = []string{
	`   █████████                      █████  █████                `,
	`  ███░░░░░███                    ░░███  ░░███                 `,
	` ███     ░░░   ██████  ████████   ░███  █████    █████████   `,
	`░███          ███░░███░░███░░███  ░███ ░░░███░    ░█░░░░███   `,
	`░███    █████░███████  ░███ ░███  ░███   ░███     ░   ███░    `,
	`░░███  ░░███ ░███░░░   ░███ ░███  ░███   ░███ ███   ███░  █   `,
	` ░░█████████ ░░██████  ████ █████ █████  ░░█████   █████████  `,
	`  ░░░░░░░░░   ░░░░░░  ░░░░ ░░░░░ ░░░░░   ░░░░░   ░░░░░░░░░   `,
}

// lebar logo (panjang baris terpanjang tanpa trailing space)
const logoWidth = 64

var katSet = map[rune]bool{
	'ゲ': true, 'ェ': true, 'ニ': true,
	'ト': true, 'ズ': true,
}

var (
	versionLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	versionValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22D3EE")).
		Bold(true)

	dividerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2D1B69"))

	hintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151"))

	containerStyle = lipgloss.NewStyle().
		Padding(1, 3)
)

type model struct {
	goVersion string
	ready     bool
}

func (m model) Init() tea.Cmd { return fetchGoVersion }

type goVersionMsg string

func fetchGoVersion() tea.Msg {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return goVersionMsg("not found — make sure Go is installed")
	}
	return goVersionMsg(strings.TrimPrefix(strings.TrimSpace(string(out)), "go version "))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case goVersionMsg:
		m.goVersion = string(msg)
		m.ready = true
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

// renderLogo — gradient per baris, ░ lebih gelap untuk depth
func renderLogo() string {
	var sb strings.Builder

	for i, row := range logo {
		mainColor := gradient[minInt(i, len(gradient)-1)]
		depthColor := gradient[minInt(i+2, len(gradient)-1)]
		for _, ch := range row {
			switch ch {
			case ' ':
				sb.WriteRune(' ')
			case '░':
				sb.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(depthColor)).
					Render(string(ch)))
			default:
				sb.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(mainColor)).
					Bold(true).
					Render(string(ch)))
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// renderTagline — tagline di-center terhadap logoWidth
// format: ░ゲェニト░  go initializer project  ░ズ░
func renderTagline() string {
	// bagian-bagian tagline
	type segment struct {
		text  string
		color string
		faint bool
		bold  bool
	}

	segs := []segment{
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

	// hitung panjang plain text untuk centering
	plain := ""
	for _, s := range segs {
		plain += s.text
	}
	plainLen := len([]rune(plain))
	padding := (logoWidth - plainLen) / 2
	if padding < 0 {
		padding = 0
	}

	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", padding))
	for _, s := range segs {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(s.color))
		if s.bold {
			style = style.Bold(true)
		}
		if s.faint {
			style = style.Faint(true)
		}
		sb.WriteString(style.Render(s.text))
	}
	sb.WriteRune('\n')
	return sb.String()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m model) View() string {
	if !m.ready {
		return "\n  Loading...\n"
	}

	var sb strings.Builder
	sb.WriteString(renderLogo())
	sb.WriteString(renderTagline())
	sb.WriteString("\n" + dividerStyle.Render(strings.Repeat("━", 64)) + "\n\n")
	sb.WriteString(
		versionLabelStyle.Render("  Go version  ") +
			versionValueStyle.Render(m.goVersion) + "\n\n",
	)
	sb.WriteString(hintStyle.Render("  q · esc · ctrl+c  quit") + "\n")

	return containerStyle.Render(sb.String())
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
