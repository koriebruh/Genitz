package tui

import "github.com/charmbracelet/lipgloss"

type uiStyles struct {
	Brand       lipgloss.Style
	Title       lipgloss.Style
	Selected    lipgloss.Style
	Cursor      lipgloss.Style
	Description lipgloss.Style
	Checkbox    lipgloss.Style
	Name        lipgloss.Style
}

var styles = newUIStyles()

func newUIStyles() uiStyles {
	return uiStyles{
		Brand:       lipgloss.NewStyle().Foreground(lipgloss.Color("#7F56D9")).Bold(true),
		Title:       lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true).MarginBottom(1),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true),
		Cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true),
		Description: lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Italic(true),
		Checkbox:    lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true),
		Name:        lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
	}
}
