package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// stepDef defines a single entry in the top step navigation bar.
type stepDef struct{ num, label string }

var wizardSteps = []stepDef{
	{"①", "Project"},
	{"②", "Architecture"},
	{"③", "Dependencies"},
	{"④", "Review"},
}

// keyHint is a keyboard shortcut + action description pair for the footer.
type keyHint struct{ key, action string }

// Model is the root Bubble Tea model for the Genitz TUI wizard.
type Model struct {
	Step Step

	// Project info inputs + inline validation errors.
	FolderInput textinput.Model
	FolderErr   string
	PkgInput    textinput.Model
	PkgErr      string

	ArchOptions  []string
	ArchCursor   int
	SelectedArch string

	Registry []Dependency
	Chosen   map[int]Dependency

	// Cursor is the visual position inside the buildVisibleOrder slice,
	// NOT a raw Registry index. Use currentDepIndex() to get the registry index.
	Cursor       int
	DepsOffset   int // first visible item index for scrollable dep list
	SearchQuery  string
	SearchActive bool

	// Terminal dimensions — updated via tea.WindowSizeMsg.
	Width  int
	Height int

	Done bool
}

// InitialModel constructs the initial model starting at StepFolder.
func InitialModel() *Model {
	f := textinput.New()
	f.Placeholder = "my-awesome-app"
	f.CharLimit = 64
	f.Focus()

	p := textinput.New()
	p.Placeholder = "github.com/username/repo"
	p.CharLimit = 128

	return &Model{
		Step:        StepFolder,
		FolderInput: f,
		PkgInput:    p,
		ArchOptions: []string{ArchMicro, ArchClean, ArchStandard, ArchDDD, ArchCLI},
		Registry:    DependencyRegistry,
		Chosen:      make(map[int]Dependency),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always handle window resize.
	if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width = wsMsg.Width
		m.Height = wsMsg.Height
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if cmd, handled := m.handleGlobalKeys(keyMsg); handled {
			return m, cmd
		}
		if cmd, handled := m.handleStepKeys(keyMsg); handled {
			return m, cmd
		}
	}
	return m, m.updateInputs(msg)
}

// View renders the current state. Every step goes through renderFrame.
func (m *Model) View() string {
	var content string
	switch {
	case m.Done:
		content = m.viewDone()
	case m.Step == StepFolder:
		content = m.viewFolder()
	case m.Step == StepPackage:
		content = m.viewPackage()
	case m.Step == StepArch:
		content = m.viewArchitecture()
	case m.Step == StepDeps:
		content = m.renderDependencyView()
	case m.Step == StepReview:
		content = m.viewReview()
	}
	return m.renderFrame(content)
}

// renderFrame wraps panel content with an adaptive header, step nav, and divider.
// The header shrinks or disappears based on terminal height to prevent double
// rendering when the user zooms in/out (always use tea.WithAltScreen).
func (m *Model) renderFrame(content string) string {
	var b strings.Builder

	switch {
	case m.Height == 0 || m.Height >= 28:
		// Height 0 means we haven't received WindowSizeMsg yet — show full header.
		b.WriteString(RenderHeader(m.Width))
	case m.Height >= 18:
		b.WriteString(RenderHeaderCompact())
		// Below 18: skip header entirely, maximise content space.
	}

	b.WriteString(m.renderStepNav())

	// Full-width divider
	dw := dividerWidth(m.Width)
	b.WriteString(styles.Divider.Render(strings.Repeat("━", dw)))
	b.WriteString("\n\n")

	// Render content at full usable width so panels feel immersive.
	usableW := m.Width - 10
	if usableW < 60 || m.Width == 0 {
		usableW = 0 // let lipgloss use natural width when terminal size unknown
	}
	containerStyle := styles.Container
	if usableW > 0 {
		containerStyle = containerStyle.Width(usableW)
	}
	b.WriteString(containerStyle.Render(content))
	return b.String()
}

// dividerWidth returns the horizontal rule length clamped to terminal width.
func dividerWidth(w int) int {
	const max = splashLogoWidth + 4
	if w <= 0 {
		return max
	}
	n := w - 8
	if n > max {
		return max
	}
	if n < 20 {
		return 20
	}
	return n
}

// renderStepNav renders the step progress bar.
// Done steps show ✔ in emerald; the active step is a sky-blue pill;
// pending steps are muted.
func (m *Model) renderStepNav() string {
	current := m.stepIndex()
	var b strings.Builder
	b.WriteString("  ")
	for i, step := range wizardSteps {
		switch {
		case i < current:
			b.WriteString(styles.StepDone.Render("✔ " + step.label))
		case i == current:
			// Active step: rendered as a filled pill for clear focus
			b.WriteString(styles.StepActive.Render(" " + step.num + " " + step.label + " "))
		default:
			b.WriteString(styles.StepPending.Render("  " + step.label))
		}
		if i < len(wizardSteps)-1 {
			b.WriteString(styles.StepSep.Render("  ›  "))
		}
	}
	b.WriteString("\n\n")
	return b.String()
}

// stepIndex maps the current Step to the 0-based index into wizardSteps.
func (m *Model) stepIndex() int {
	switch m.Step {
	case StepFolder, StepPackage:
		return 0
	case StepArch:
		return 1
	case StepDeps:
		return 2
	case StepReview:
		return 3
	}
	return 0
}

// renderKeyHints renders a footer row of keyboard shortcut badges.
func renderKeyHints(hints []keyHint) string {
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, styles.KeyBadge.Render(h.key)+" "+styles.KeyHint.Render(h.action))
	}
	return strings.Join(parts, "   ") + "\n"
}

// errLine renders an inline validation error.
func errLine(msg string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true).
		Render("  ✗ "+msg) + "\n"
}

// ── Validation ────────────────────────────────────────────────────────────────

// validateFolder returns a non-empty error string when the folder name is invalid.
func validateFolder(v string) string {
	if strings.TrimSpace(v) == "" {
		return "folder name cannot be empty"
	}
	for _, ch := range v {
		if unicode.IsSpace(ch) || strings.ContainsRune(`/\:*?"<>|`, ch) {
			return `no spaces or special characters ( / \ : * ? " < > | )`
		}
	}
	return ""
}

// validateModulePath returns a non-empty error string when the module path is invalid.
func validateModulePath(v string) string {
	if strings.TrimSpace(v) == "" {
		return "module path cannot be empty"
	}
	if strings.ContainsAny(v, " \t\n") {
		return "module path cannot contain spaces"
	}
	return ""
}

// ── Key handlers ──────────────────────────────────────────────────────────────

func (m *Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.String() == "ctrl+c" {
		return tea.Quit, true
	}
	return nil, false
}

func (m *Model) handleStepKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch m.Step {
	case StepFolder:
		return m.handleFolderKeys(msg)
	case StepPackage:
		return m.handlePackageKeys(msg)
	case StepArch:
		return m.handleArchKeys(msg)
	case StepDeps:
		return m.handleDepsKeys(msg)
	case StepReview:
		return m.handleReviewKeys(msg)
	}
	return nil, false
}

// updateInputs forwards non-key messages to the active textinput.
func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	switch m.Step {
	case StepFolder:
		var cmd tea.Cmd
		m.FolderInput, cmd = m.FolderInput.Update(msg)
		return cmd
	case StepPackage:
		var cmd tea.Cmd
		m.PkgInput, cmd = m.PkgInput.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) handleFolderKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.String() == "enter" {
		if e := validateFolder(m.FolderInput.Value()); e != "" {
			m.FolderErr = e
			return nil, true
		}
		m.FolderErr = ""
		m.FolderInput.Blur()
		m.PkgInput.Focus()
		m.Step = StepPackage
		return nil, true
	}
	m.FolderErr = "" // clear error on any other key
	return nil, false
}

func (m *Model) handlePackageKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.String() == "enter" {
		if e := validateModulePath(m.PkgInput.Value()); e != "" {
			m.PkgErr = e
			return nil, true
		}
		m.PkgErr = ""
		m.PkgInput.Blur()
		m.Step = StepArch
		return nil, true
	}
	m.PkgErr = ""
	return nil, false
}

func (m *Model) handleArchKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "up", "k":
		if m.ArchCursor > 0 {
			m.ArchCursor--
		}
		return nil, true
	case "down", "j":
		if m.ArchCursor < len(m.ArchOptions)-1 {
			m.ArchCursor++
		}
		return nil, true
	case "enter":
		// Block selection of templates that are not yet implemented.
		if !AvailableArchitectures[m.ArchOptions[m.ArchCursor]] {
			return nil, true // consume key, message already shown in view
		}
		m.SelectedArch = m.ArchOptions[m.ArchCursor]
		m.Step = StepDeps
		m.Cursor = 0
		return nil, true
	case "q":
		return tea.Quit, true
	}
	return nil, false
}

func (m *Model) handleDepsKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	// Search mode: most keys still work, printable chars go to query.
	if m.SearchActive {
		switch msg.String() {
		case "esc":
			m.SearchQuery = ""
			m.SearchActive = false
			m.clampCursor()
			return nil, true
		case "backspace":
			if len(m.SearchQuery) > 0 {
				runes := []rune(m.SearchQuery)
				m.SearchQuery = string(runes[:len(runes)-1])
				m.clampCursor()
			}
			return nil, true
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
			m.syncDepsOffset()
			return nil, true
		case "down", "j":
			if m.Cursor < len(m.buildVisibleOrder())-1 {
				m.Cursor++
			}
			m.syncDepsOffset()
			return nil, true
		case " ":
			if idx := m.currentDepIndex(); idx >= 0 {
				m.toggleDependency(idx)
			}
			return nil, true
		case "enter":
			m.SearchActive = false
			m.Step = StepReview
			return nil, true
		default:
			if len(msg.Runes) == 1 && unicode.IsPrint(msg.Runes[0]) {
				m.SearchQuery += string(msg.Runes)
				m.clampCursor()
			}
			return nil, true
		}
	}

	// Normal mode.
	switch msg.String() {
	case "/":
		m.SearchActive = true
		return nil, true
	case "esc":
		if m.SearchQuery != "" {
			m.SearchQuery = ""
			m.clampCursor()
		}
		return nil, true
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
		m.syncDepsOffset()
		return nil, true
	case "down", "j":
		if m.Cursor < len(m.buildVisibleOrder())-1 {
			m.Cursor++
		}
		return nil, true
	case " ":
		if idx := m.currentDepIndex(); idx >= 0 {
			m.toggleDependency(idx)
		}
		return nil, true
	case "enter":
		m.Step = StepReview
		return nil, true
	case "q":
		return tea.Quit, true
	}
	return nil, false
}

func (m *Model) handleReviewKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter", "y":
		m.Done = true
		return tea.Quit, true
	case "b":
		m.Step = StepDeps
		return nil, true
	case "q":
		return tea.Quit, true
	}
	return nil, false
}

// clampCursor ensures Cursor stays within the visible order after a search change,
// then syncs the scroll offset so the cursor stays visible.
func (m *Model) clampCursor() {
	order := m.buildVisibleOrder()
	if len(order) == 0 {
		m.Cursor = 0
		m.DepsOffset = 0
		return
	}
	if m.Cursor >= len(order) {
		m.Cursor = len(order) - 1
	}
	m.syncDepsOffset()
}

// depsMaxVisible returns how many dependency rows fit on-screen at once.
func (m *Model) depsMaxVisible() int {
	if m.Height == 0 {
		return 10 // fallback before first WindowSizeMsg
	}
	// Fixed rows: header(0-10) + step nav(2) + divider(2) + panel header(4) + search(2) + footer(2)
	fixed := 14
	if m.Height >= 28 {
		fixed += 10 // full logo
	} else if m.Height >= 18 {
		fixed += 2 // compact header
	}
	avail := m.Height - fixed
	if avail < 2 {
		return 2
	}
	return avail / 2 // each item occupies 2 lines (name + description)
}

// syncDepsOffset adjusts DepsOffset so the cursor is always visible.
func (m *Model) syncDepsOffset() {
	maxVis := m.depsMaxVisible()
	if maxVis <= 0 {
		return
	}
	if m.Cursor < m.DepsOffset {
		m.DepsOffset = m.Cursor
	}
	if m.Cursor >= m.DepsOffset+maxVis {
		m.DepsOffset = m.Cursor - maxVis + 1
	}
	if m.DepsOffset < 0 {
		m.DepsOffset = 0
	}
}

func (m *Model) toggleDependency(registryIndex int) {
	if _, ok := m.Chosen[registryIndex]; ok {
		delete(m.Chosen, registryIndex)
		return
	}
	m.Chosen[registryIndex] = m.Registry[registryIndex]
}

// currentDepIndex returns the Registry index of the highlighted dep, or -1.
func (m *Model) currentDepIndex() int {
	order := m.buildVisibleOrder()
	if len(order) == 0 || m.Cursor >= len(order) {
		return -1
	}
	return order[m.Cursor]
}

// ── Panel views ───────────────────────────────────────────────────────────────

func (m *Model) viewFolder() string {
	var b strings.Builder
	b.WriteString(styles.PanelLabel.Render("FOLDER NAME") + "\n")
	b.WriteString(styles.PanelHint.Render("Name of the project directory — no spaces") + "\n\n")
	b.WriteString(styles.InputPrompt.Render("$ ") + m.FolderInput.View() + "\n")
	if m.FolderErr != "" {
		b.WriteString(errLine(m.FolderErr))
	} else {
		b.WriteString(styles.InputNote.Render("  will be created at ./<folder>/") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(renderKeyHints([]keyHint{{"enter", "next"}, {"ctrl+c", "quit"}}))
	return b.String()
}

func (m *Model) viewPackage() string {
	folder := m.FolderInput.Value()
	if folder == "" {
		folder = "my-app"
	}
	var b strings.Builder
	b.WriteString(styles.PanelLabel.Render("MODULE PATH") + "\n")
	b.WriteString(styles.PanelHint.Render("Go module path written to go.mod — no spaces") + "\n\n")
	b.WriteString(styles.InputPrompt.Render("$ ") + m.PkgInput.View() + "\n")
	if m.PkgErr != "" {
		b.WriteString(errLine(m.PkgErr))
	} else {
		b.WriteString(styles.InputNote.Render("  e.g. github.com/username/"+folder) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(renderKeyHints([]keyHint{{"enter", "next"}, {"ctrl+c", "quit"}}))
	return b.String()
}

func (m *Model) viewArchitecture() string {
	var b strings.Builder
	b.WriteString(styles.PanelLabel.Render("ARCHITECTURE") + "\n")
	b.WriteString(styles.PanelHint.Render("Choose the project structure template") + "\n\n")

	for i, opt := range m.ArchOptions {
		isActive := m.ArchCursor == i
		available := AvailableArchitectures[opt]

		cursor := "   "
		var name string
		switch {
		case isActive && available:
			cursor = styles.Cursor.Render(" ▶ ")
			name = styles.Selected.Render(opt)
		case isActive && !available:
			cursor = styles.Cursor.Render(" ▶ ")
			name = styles.StepPending.Render(opt)
		default:
			if available {
				name = styles.Name.Render(opt)
			} else {
				name = styles.StepPending.Render(opt)
			}
		}

		var badge string
		if !available {
			badge = " " + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 1).
				Render("coming soon")
		}

		b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, name, badge))
		b.WriteString(fmt.Sprintf("     %s\n\n", styles.Description.Render(archDescriptions[opt])))
	}

	if !AvailableArchitectures[m.ArchOptions[m.ArchCursor]] {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Italic(true).
			Render("  This template is not yet available, please choose another.") + "\n\n")
	}

	b.WriteString(renderKeyHints([]keyHint{
		{"↑↓ / jk", "navigate"},
		{"enter", "select"},
		{"q", "quit"},
	}))
	return b.String()
}

// viewReview renders a clean two-section summary before generation.
func (m *Model) viewReview() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("REVIEW") + "\n")
	b.WriteString(styles.PanelHint.Render("Confirm your configuration before scaffolding") + "\n\n")

	folder := m.FolderInput.Value()
	if folder == "" {
		folder = "(not set)"
	}
	pkg := m.PkgInput.Value()
	if pkg == "" {
		pkg = "(not set)"
	}
	arch := m.SelectedArch
	if arch == "" {
		arch = "(not set)"
	}

	const keyW = 16
	hr := styles.Divider.Render(strings.Repeat("─", 52)) + "\n"

	summaryRow := func(key, value string, valStyle lipgloss.Style) string {
		k := lipgloss.NewStyle().Foreground(colorMuted).Width(keyW).Render(key)
		return fmt.Sprintf("  %s  %s\n", k, valStyle.Render(value))
	}

	b.WriteString(styles.Description.Render("  PROJECT") + "\n")
	b.WriteString(hr)
	b.WriteString(summaryRow("Folder", folder, styles.StepActive))
	b.WriteString(summaryRow("Module", pkg, styles.StepActive))
	b.WriteString(summaryRow("Architecture", arch, styles.Checkbox))
	b.WriteString(summaryRow("Output", "./"+folder+"/", styles.StepPending))
	b.WriteString("\n")

	b.WriteString(styles.Description.Render("  DEPENDENCIES") + "\n")
	b.WriteString(hr)

	if len(m.Chosen) == 0 {
		b.WriteString(fmt.Sprintf("  %s\n\n", styles.Description.Render("none selected")))
	} else {
		for _, group := range depGroups {
			var hits []Dependency
			for _, dep := range m.Chosen {
				for _, cat := range group.categories {
					if dep.Category == cat {
						hits = append(hits, dep)
						break
					}
				}
			}
			if len(hits) == 0 {
				continue
			}
			b.WriteString(groupHeaderStyle.Render("  ▸ "+group.label) + "\n")
			for _, dep := range hits {
				badge := getBadgeStyle(dep.Category).Render(strings.ToUpper(dep.Category))
				k := lipgloss.NewStyle().Foreground(colorText).Width(keyW + 2).Render(dep.Name)
				b.WriteString(fmt.Sprintf("    %s %s  %s\n",
					k, badge, styles.Description.Render(dep.ImportPath),
				))
			}
			b.WriteRune('\n')
		}
	}

	b.WriteString(hr)
	b.WriteString(renderKeyHints([]keyHint{
		{"enter / y", "generate"},
		{"b", "back"},
		{"q", "quit"},
	}))
	return b.String()
}

// viewDone is shown briefly before tea.Quit takes effect.
func (m *Model) viewDone() string {
	var b strings.Builder
	b.WriteString(styles.Checkbox.Render("✔ Ready to scaffold!") + "\n\n")
	b.WriteString(styles.Name.Render("  Generating "+m.SelectedArch+" project...") + "\n")
	return b.String()
}
