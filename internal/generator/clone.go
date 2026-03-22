package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CloneRemoteTemplate downloads a remote git repository and cleans up its .git history
// so it acts as a fresh project template immediately usable by Genitz.
func CloneRemoteTemplate(repoURL, projectName string) error {
	targetPath, err := filepath.Abs(projectName)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	if err := ensureFreshProjectDir(targetPath); err != nil {
		return err
	}

	fmt.Printf("🌍 Fetching remote template from %s...\n", repoURL)
	cmd := exec.Command("git", "clone", "--depth=1", repoURL, targetPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(targetPath) // cleanup partial clone
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Remove the .git folder so the project is detached from the template source
	gitDir := filepath.Join(targetPath, ".git")
	_ = os.RemoveAll(gitDir)

	fmt.Printf("⚙️  Adapting Go module name to %s...\n", projectName)
	
	// We run `go mod edit -module=...` just in case there's an existing go.mod from the template
	// Alternatively, if there is no go.mod, we init one.
	modPath := filepath.Join(targetPath, "go.mod")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		initCmd := exec.Command("go", "mod", "init", projectName)
		initCmd.Dir = targetPath
		initCmd.Run()
	} else {
		editCmd := exec.Command("go", "mod", "edit", "-module="+projectName)
		editCmd.Dir = targetPath
		editCmd.Run()
	}

	fmt.Println("✨ Finalizing project...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = targetPath
	tidyCmd.Run()

	fmt.Printf("\n✅ Successfully initialized %q from remote template!\n", projectName)
	fmt.Printf("➜ cd %s\n", projectName)
	return nil
}
