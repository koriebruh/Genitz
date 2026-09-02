package generator

import (
	"os"
	"path/filepath"
)

const goGitignore = `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Test binary, build with 'go test -c'
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work
go.work.sum

# Env files
.env
`

// runGitInit initializes a git repo at dir, writes a Go-flavored .gitignore
// if one doesn't already exist, stages everything, and makes the first
// commit. Fully local — no network calls, nothing published.
func runGitInit(dir string) error {
	if err := runCaptured(dir, "git", "init"); err != nil {
		return err
	}

	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(goGitignore), 0o644); err != nil {
			return err
		}
	}

	if err := runCaptured(dir, "git", "add", "-A"); err != nil {
		return err
	}
	return runCaptured(dir, "git", "commit", "-m", "Initial commit from genitz")
}

// SuggestedPublishCommand returns the command to actually publish the repo
// to GitHub — deliberately not run automatically. gh repo creation is a
// real external side effect with a visibility decision the user should
// make explicitly, and InstallStep failures are fatal, so a missing/
// unauthenticated gh shouldn't be able to abort an otherwise-successful
// scaffold.
func SuggestedPublishCommand() string {
	return "gh repo create --private --source=. --push"
}
