package generator

import "fmt"

// contributingContent returns a minimal CONTRIBUTING.md pointing at the
// standard fork/branch/PR flow — genitz has no way to know a project's real
// contribution process, so this stays generic rather than guessing at one.
func contributingContent(projectName string) string {
	return fmt.Sprintf(`# Contributing to %s

1. Fork the repository and create a branch off main.
2. Make your change, with tests where it makes sense.
3. Run `+"`go build ./...`, `go vet ./...`, `go test ./...`, and `gofmt -l .`"+`.
4. Open a pull request describing what changed and why.
`, projectName)
}

// securityContent returns a minimal SECURITY.md with a placeholder contact
// — same "leave it for the user to fill in" spirit as license.go's
// copyright holder, since genitz has no real security contact to guess.
func securityContent() string {
	return `# Security Policy

If you discover a security vulnerability, please report it privately
instead of opening a public issue — email [SECURITY CONTACT] or use
GitHub's private vulnerability reporting for this repository.
`
}

// dependabotContent returns a .github/dependabot.yml that keeps Go module
// dependencies (and, if IncludeCI is also on, the GitHub Actions versions
// pinned in ci.yml) up to date on a weekly cadence.
func dependabotContent(includeCI bool) string {
	content := `version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
`
	if includeCI {
		content += `  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
`
	}
	return content
}

// bugReportIssueTemplate and featureRequestIssueTemplate are minimal
// GitHub issue templates — same rationale as CONTRIBUTING.md, generic
// rather than guessing at project-specific triage fields.
const bugReportIssueTemplate = `---
name: Bug report
about: Report something that isn't working as expected
labels: bug
---

**Describe the bug**

**Steps to reproduce**

**Expected behavior**

**Environment** (OS, Go version, etc.)
`

const featureRequestIssueTemplate = `---
name: Feature request
about: Suggest an idea for this project
labels: enhancement
---

**What problem would this solve?**

**Describe the solution you'd like**
`
