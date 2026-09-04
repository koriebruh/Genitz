package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const undoRootEnvVar = "GENITZ_UNDO_OVERRIDE"

// undoDirFor returns the snapshot directory for targetDir — keyed by a
// hash of its absolute path so genitz can track "the last snapshot for
// this project" without a database, and so two different projects never
// collide even if run concurrently.
func undoDirFor(targetDir string) (string, error) {
	abs, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("resolve project path: %w", err)
	}
	sum := sha256.Sum256([]byte(abs))
	key := hex.EncodeToString(sum[:])

	if root := os.Getenv(undoRootEnvVar); root != "" {
		return filepath.Join(root, key), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine user config directory: %w", err)
	}
	return filepath.Join(configDir, "genitz", "undo", key), nil
}

// SnapshotForUndo copies go.mod (required) and go.sum (if present) from
// targetDir into that project's undo directory, overwriting any previous
// snapshot — deliberately LIFO/single-level, not a full history tree, to
// keep `genitz undo` predictable ("undoes the last add/remove") instead of
// needing its own navigation UI. Called right before BuildAddSteps/
// BuildRemoveSteps run; a failure here is returned (unlike RecordHistory's
// silent best-effort) since it means `genitz undo` silently can't help —
// worth surfacing, though callers should still let the actual add/remove
// proceed rather than treat it as fatal.
func SnapshotForUndo(targetDir string) error {
	dir, err := undoDirFor(targetDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create undo snapshot directory: %w", err)
	}

	if err := copyFile(filepath.Join(targetDir, "go.mod"), filepath.Join(dir, "go.mod")); err != nil {
		return fmt.Errorf("snapshot go.mod: %w", err)
	}

	sumSrc := filepath.Join(targetDir, "go.sum")
	if _, err := os.Stat(sumSrc); err == nil {
		if err := copyFile(sumSrc, filepath.Join(dir, "go.sum")); err != nil {
			return fmt.Errorf("snapshot go.sum: %w", err)
		}
	} else {
		// No go.sum yet (e.g. a project with no dependencies at all) — not
		// an error, just nothing to snapshot; remove any stale one from a
		// previous snapshot so Undo doesn't restore a go.sum that
		// shouldn't exist anymore.
		os.Remove(filepath.Join(dir, "go.sum"))
	}
	return nil
}

// Undo restores the most recent snapshot for targetDir over its current
// go.mod/go.sum. Scope is deliberately narrow: it only reverts what
// SnapshotForUndo captured (go.mod/go.sum), not any other file — genitz
// doesn't track a full diff, so it doesn't pretend to offer one.
func Undo(targetDir string) error {
	dir, err := undoDirFor(targetDir)
	if err != nil {
		return err
	}

	modSnapshot := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(modSnapshot); errors.Is(err, os.ErrNotExist) {
		return errors.New("nothing to undo for this project (no add/remove has been recorded here yet)")
	} else if err != nil {
		return fmt.Errorf("check undo snapshot: %w", err)
	}

	if err := copyFile(modSnapshot, filepath.Join(targetDir, "go.mod")); err != nil {
		return fmt.Errorf("restore go.mod: %w", err)
	}

	sumSnapshot := filepath.Join(dir, "go.sum")
	if _, err := os.Stat(sumSnapshot); err == nil {
		if err := copyFile(sumSnapshot, filepath.Join(targetDir, "go.sum")); err != nil {
			return fmt.Errorf("restore go.sum: %w", err)
		}
	} else {
		// The snapshot had no go.sum (project had zero deps at snapshot
		// time) — remove the current one too so the restore is exact.
		os.Remove(filepath.Join(targetDir, "go.sum"))
	}

	return os.RemoveAll(dir)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
