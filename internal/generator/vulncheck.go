package generator

import (
	"os/exec"
	"strings"
)

// VulnCheckAdvisory runs `govulncheck ./...` in dir if the binary is on
// PATH and returns a short advisory string to print after a successful
// install — "" when govulncheck isn't installed (nothing to report, not a
// failure) or when it found nothing. This is advisory-only, never fatal:
// InstallStep failures abort the whole run, and a known vulnerability in a
// freshly-picked dependency shouldn't block scaffolding the project.
func VulnCheckAdvisory(dir string) string {
	if err := CheckBinary("govulncheck"); err != nil {
		return ""
	}

	cmd := exec.Command("govulncheck", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return ""
	}
	// A non-zero exit from govulncheck means it found something (or the
	// module isn't buildable yet) — surface the summary either way rather
	// than silently discarding a real finding.
	text := strings.TrimSpace(string(out))
	if text == "" {
		return ""
	}
	return "⚠️  govulncheck found potential issues:\n" + text
}
