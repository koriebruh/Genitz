package generator

import (
	"fmt"
	"os/exec"
)

// CheckBinary fails fast with a clear message when name isn't on PATH,
// instead of letting a raw exec.Command error surface mid-InstallStep
// animation (InstallStep failures are fatal, so the sooner the message the
// better — see runGitInit / go mod callers for where this is used).
func CheckBinary(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%q is required but not found on PATH", name)
	}
	return nil
}
