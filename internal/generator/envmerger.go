package generator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvVar represents a key=value pair from a .env file.
type EnvVar struct {
	Key   string
	Value string
}

// MergeEnvFile reads an existing .env file (if any) and appends new key=value
// pairs that are not already present. Existing values are never overwritten.
func MergeEnvFile(envFilePath string, newVars []EnvVar) error {
	existing := map[string]bool{}

	// Read existing keys
	if f, err := os.Open(envFilePath); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if idx := strings.Index(line, "="); idx > 0 {
				existing[strings.TrimSpace(line[:idx])] = true
			}
		}
	}

	// Open file in append mode (create if missing)
	f, err := os.OpenFile(envFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open .env file: %w", err)
	}
	defer f.Close()

	wrote := false
	for _, v := range newVars {
		if existing[v.Key] {
			continue // do not overwrite
		}
		if !wrote {
			// Blank separator line before new block
			fmt.Fprintln(f)
			wrote = true
		}
		fmt.Fprintf(f, "%s=%s\n", v.Key, v.Value)
	}
	return nil
}

// FindEnvFile walks up from baseDir looking for a .env file (max 3 levels).
func FindEnvFile(baseDir string) string {
	dir := baseDir
	for i := 0; i < 3; i++ {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
