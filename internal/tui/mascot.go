package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// mascotStyle is deliberately plain foreground-only (no background block,
// unlike Cursor/Checkbox after the brutalist pass) — a background fill
// would paint solid rectangles behind the cat's whitespace, which looks
// broken rather than blocky.
var mascotStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

// mascotFrames are small ASCII cat frames cycled on the Installing screen
// — a contained delight detail, confined to StepInstalling only so it
// never competes with the InstallStep list, which is the actual content.
var mascotFrames = []string{
	" /\\_/\\\n( o.o )\n > ^ <",
	" /\\_/\\\n( -.- )\n > ^ <",
	" /\\_/\\\n( o.o )\n >   <",
}

// mascotDoneFrame is shown once every InstallStep has finished successfully.
const mascotDoneFrame = " /\\_/\\\n( ^.^ )\n > w <"

// mascotTicksPerFrame throttles the frame advance — the spinner ticks
// every ~100-150ms (MiniDot), which would flicker the cat if it advanced
// every tick.
const mascotTicksPerFrame = 4

// renderMascot returns the current frame, colored, or the "done" pose once
// install finishes without error.
func (m *Model) renderMascot() string {
	frame := mascotDoneFrame
	if !m.Done && m.InstallErr == nil {
		frame = mascotFrames[(m.MascotTick/mascotTicksPerFrame)%len(mascotFrames)]
	}

	var b strings.Builder
	for _, line := range strings.Split(frame, "\n") {
		b.WriteString(mascotStyle.Render(line) + "\n")
	}
	return b.String()
}
