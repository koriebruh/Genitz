package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// stepDef defines a single entry in the top step navigation bar.
type stepDef struct{ num, label string }

var initSteps = []stepDef{
	{"①", "Project"},
	{"②", "Dependencies"},
	{"③", "Docker"},
	{"④", "Review"},
	{"⑤", "Installing"},
}

var addSteps = []stepDef{
	{"①", "Dependencies"},
	{"②", "Review"},
	{"③", "Installing"},
}

// keyHint is a keyboard shortcut + action description pair for the footer.
type keyHint struct{ key, action string }

// Model is the root Bubble Tea model for the Genitz TUI wizard.
type Model struct {
	Mode Mode
	Step Step

	// Project info inputs + inline validation errors. Only used in ModeInit.
	FolderInput textinput.Model
	FolderErr   string
	PkgInput    textinput.Model
	PkgErr      string

	// ExistingModule is the module path read from ./go.mod. Only set in ModeAdd.
	ExistingModule string

	Registry []Dependency
	Chosen   map[int]Dependency

	// ExpandedGroup is the depGroups index currently expanded in the
	// dependency picker (-1 = none). Only one group is open at a time.
	ExpandedGroup int

	// Cursor is the visual position inside visibleRows(), NOT a raw Registry
	// index — use activateCursorRow() to act on whatever row it's on.
	Cursor       int
	SearchQuery  string
	SearchActive bool

	// Terminal dimensions — updated via tea.WindowSizeMsg.
	Width  int
	Height int

	// IncludeDocker is the StepDocker toggle. Init mode only.
	IncludeDocker bool
	// LookupComposeServices is injected by main.go (generator.ComposeServiceNames)
	// so StepDocker can preview which selected deps map to a compose service,
	// without the tui package importing generator (the dependency runs the
	// other way already).
	LookupComposeServices func(depIDs []string) []string

	// BuildSteps is injected by main.go: given the confirmed model, it
	// returns the ordered install work to animate on StepInstalling.
	BuildSteps   func(*Model) []InstallStep
	InstallSteps []InstallStep
	InstallIndex int
	InstallErr   error
	Spinner      spinner.Model

	Done bool
}

// InitialModel constructs the model for scaffolding a brand-new project.
func InitialModel() *Model {
	f := textinput.New()
	f.Placeholder = "my-awesome-app"
	f.CharLimit = 64
	f.Focus()

	p := textinput.New()
	p.Placeholder = "github.com/username/repo"
	p.CharLimit = 128

	return &Model{
		Mode:          ModeInit,
		Step:          StepFolder,
		FolderInput:   f,
		PkgInput:      p,
		Registry:      DependencyRegistry,
		Chosen:        make(map[int]Dependency),
		ExpandedGroup: -1,
		Spinner:       newSpinner(),
	}
}

// AddModel constructs the model for adding dependencies to the project
// found in the current directory, whose module path is existingModule.
func AddModel(existingModule string) *Model {
	return &Model{
		Mode:           ModeAdd,
		Step:           StepDeps,
		ExistingModule: existingModule,
		Registry:       DependencyRegistry,
		Chosen:         make(map[int]Dependency),
		ExpandedGroup:  -1,
		Spinner:        newSpinner(),
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

	switch msg := msg.(type) {
	case stepDoneMsg:
		cmd, _ := m.handleStepDone(msg)
		return m, cmd
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
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
	case m.Step == StepDeps:
		content = m.renderDependencyView()
	case m.Step == StepDocker:
		content = m.viewDocker()
	case m.Step == StepReview:
		content = m.viewReview()
	case m.Step == StepInstalling:
		content = m.viewInstalling()
	}
	return m.renderFrame(content)
}

// renderFrame wraps panel content with an adaptive header, step nav, and divider.
// Bubble Tea in alt-screen mode clips anything past the terminal width instead
// of soft-wrapping it, so every piece here is picked or reflowed against the
// real terminal size instead of only reacting to height (always use tea.WithAltScreen).
func (m *Model) renderFrame(content string) string {
	w := m.Width
	if w <= 0 {
		w = 80 // no WindowSizeMsg yet — assume a normal-sized terminal for the first paint.
	}

	var b strings.Builder
	b.WriteString(m.renderHeader(w))
	b.WriteString(m.renderModeLine(w))
	b.WriteString(m.renderStepNav(w))
	b.WriteString(styles.Divider.Render(strings.Repeat("━", dividerWidth(w))))
	b.WriteString("\n\n")

	pad := 3
	if w < 50 {
		pad = 1
	}
	box := styles.Container.Padding(0, pad).Width(w)
	b.WriteString(box.Render(content))
	return b.String()
}

// renderHeader picks the full logo, the one-line compact brand, or nothing,
// based on how much room the terminal actually has. The full ASCII logo is
// reserved for tall terminals so it doesn't eat the room the dependency
// list needs on an ordinary-sized window — compact is the common case.
func (m *Model) renderHeader(w int) string {
	full := m.Height >= 34
	compactOK := m.Height == 0 || m.Height >= 14

	switch {
	case full && w >= splashLogoWidth+6:
		return RenderHeader()
	case compactOK && w >= 30:
		return RenderHeaderCompact(w)
	default:
		return ""
	}
}

// renderModeLine tells the user, up front, which of the two flows they're
// in: a brand-new project wizard, or adding dependencies to the project
// already in the current directory.
func (m *Model) renderModeLine(w int) string {
	style := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	label := "→ New project mode"
	if m.Mode == ModeAdd {
		label = "→ Add dependency mode · " + m.ExistingModule
	}
	return lipgloss.NewStyle().MaxWidth(w).Render("  "+style.Render(label)) + "\n\n"
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

// wizardSteps returns the step nav entries for the model's current mode.
func (m *Model) wizardSteps() []stepDef {
	if m.Mode == ModeAdd {
		return addSteps
	}
	return initSteps
}

// renderStepNav renders the ① Project ── ② Dependencies … bar. Below 55
// columns the text labels are dropped in favour of bare step numbers so the
// bar never gets clipped by the terminal.
func (m *Model) renderStepNav(w int) string {
	steps := m.wizardSteps()
	current := m.stepIndex()
	compact := w < 55

	var line strings.Builder
	line.WriteString("  ")
	for i, step := range steps {
		label := step.num + " " + step.label
		if compact {
			label = step.num
		}
		switch {
		case i < current:
			if compact {
				line.WriteString(styles.StepDone.Render("✓"))
			} else {
				line.WriteString(styles.StepDone.Render("✓ " + step.label))
			}
		case i == current:
			line.WriteString(styles.StepActive.Render(label))
		default:
			line.WriteString(styles.StepPending.Render(label))
		}
		if i < len(steps)-1 {
			sep := "  ──  "
			if compact {
				sep = " ─ "
			}
			line.WriteString(styles.StepSep.Render(sep))
		}
	}
	// Safety net: even the compact form can't overflow the real terminal.
	return lipgloss.NewStyle().MaxWidth(w).Render(line.String()) + "\n\n"
}

// stepIndex maps the current Step to the 0-based index into wizardSteps().
func (m *Model) stepIndex() int {
	if m.Mode == ModeAdd {
		switch m.Step {
		case StepDeps:
			return 0
		case StepReview:
			return 1
		case StepInstalling:
			return 2
		}
		return 0
	}
	switch m.Step {
	case StepFolder, StepPackage:
		return 0
	case StepDeps:
		return 1
	case StepDocker:
		return 2
	case StepReview:
		return 3
	case StepInstalling:
		return 4
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
	case StepDeps:
		return m.handleDepsKeys(msg)
	case StepDocker:
		return m.handleDockerKeys(msg)
	case StepReview:
		return m.handleReviewKeys(msg)
	}
	return nil, false
}

// nextAfterDeps is where "enter" on the dependency picker goes: the Docker
// toggle screen in Init mode (Add mode never has one — installing into an
// existing project's Docker setup, or lack of one, is out of scope).
func (m *Model) nextAfterDeps() Step {
	if m.Mode == ModeInit {
		return StepDocker
	}
	return StepReview
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
		m.Step = StepDeps
		m.Cursor = 0
		return nil, true
	}
	m.PkgErr = ""
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
			return nil, true
		case "down", "j":
			if m.Cursor < len(m.visibleRows())-1 {
				m.Cursor++
			}
			return nil, true
		case " ":
			m.activateCursorRow()
			return nil, true
		case "enter":
			m.SearchActive = false
			m.Step = m.nextAfterDeps()
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
		return nil, true
	case "down", "j":
		if m.Cursor < len(m.visibleRows())-1 {
			m.Cursor++
		}
		return nil, true
	case " ":
		m.activateCursorRow()
		return nil, true
	case "enter":
		m.Step = m.nextAfterDeps()
		return nil, true
	case "q":
		return tea.Quit, true
	}
	return nil, false
}

func (m *Model) handleReviewKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "enter", "y":
		return m.beginInstall(), true
	case "b":
		if m.Mode == ModeInit {
			m.Step = StepDocker
		} else {
			m.Step = StepDeps
		}
		return nil, true
	case "q":
		return tea.Quit, true
	}
	return nil, false
}

// clampCursor ensures Cursor stays within visibleRows() after a search change.
func (m *Model) clampCursor() {
	n := len(m.visibleRows())
	if n == 0 {
		m.Cursor = 0
		return
	}
	if m.Cursor >= n {
		m.Cursor = n - 1
	}
}

func (m *Model) toggleDependency(registryIndex int) {
	if _, ok := m.Chosen[registryIndex]; ok {
		delete(m.Chosen, registryIndex)
		return
	}
	m.Chosen[registryIndex] = m.Registry[registryIndex]
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

// viewReview renders a clean two-section summary before generation.
func (m *Model) viewReview() string {
	var b strings.Builder

	b.WriteString(styles.PanelLabel.Render("REVIEW") + "\n")
	b.WriteString(styles.PanelHint.Render("Confirm your configuration before scaffolding") + "\n\n")

	const keyW = 16
	hrWidth := m.Width - 8
	if hrWidth > 52 {
		hrWidth = 52
	} else if hrWidth < 20 {
		hrWidth = 20
	}
	hr := styles.Divider.Render(strings.Repeat("─", hrWidth)) + "\n"

	summaryRow := func(key, value string, valStyle lipgloss.Style) string {
		k := lipgloss.NewStyle().Foreground(colorMuted).Width(keyW).Render(key)
		return fmt.Sprintf("  %s  %s\n", k, valStyle.Render(value))
	}

	b.WriteString(styles.Description.Render("  PROJECT") + "\n")
	b.WriteString(hr)
	if m.Mode == ModeAdd {
		b.WriteString(summaryRow("Module", m.ExistingModule, styles.StepActive))
		b.WriteString(summaryRow("Target", "current directory", styles.StepPending))
	} else {
		folder := m.FolderInput.Value()
		if folder == "" {
			folder = "(not set)"
		}
		pkg := m.PkgInput.Value()
		if pkg == "" {
			pkg = "(not set)"
		}
		b.WriteString(summaryRow("Folder", folder, styles.StepActive))
		b.WriteString(summaryRow("Module", pkg, styles.StepActive))
		b.WriteString(summaryRow("Output", "./"+folder+"/", styles.StepPending))
	}
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
				b.WriteString(fmt.Sprintf("      %s\n", styles.Description.Render("docs: "+docURL(dep.ImportPath))))
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
	b.WriteString(styles.Checkbox.Render("✔ Done!") + "\n\n")
	if m.Mode == ModeAdd {
		b.WriteString(styles.Name.Render("  Dependencies installed.") + "\n")
	} else {
		b.WriteString(styles.Name.Render("  Project scaffolded.") + "\n")
	}
	return b.String()
}
