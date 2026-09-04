package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryEntry is one recorded genitz operation.
type HistoryEntry struct {
	Time    time.Time `json:"time"`
	Command string    `json:"command"`
	Dir     string    `json:"dir"`
	Detail  string    `json:"detail"`
}

const historyPathEnvVar = "GENITZ_HISTORY_OVERRIDE"

func historyPath() string {
	if p := os.Getenv(historyPathEnvVar); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "genitz", "history.jsonl")
}

// RecordHistory appends one entry to the local history log — best-effort,
// any failure here (no config dir, disk full, ...) is silently swallowed
// since a logging failure must never break the actual operation it's
// recording.
func RecordHistory(command, dir, detail string) {
	path := historyPath()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	b, err := json.Marshal(HistoryEntry{Time: time.Now(), Command: command, Dir: dir, Detail: detail})
	if err != nil {
		return
	}
	f.Write(append(b, '\n'))
}

// ReadHistory returns recorded entries oldest-first, or nil (not an error)
// if nothing's been recorded yet.
func ReadHistory() ([]HistoryEntry, error) {
	path := historyPath()
	if path == "" {
		return nil, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []HistoryEntry
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		if line == "" {
			continue
		}
		var e HistoryEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue // a corrupted line shouldn't take down the whole history
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// PrintHistory renders ReadHistory's result as a table.
func PrintHistory(entries []HistoryEntry) {
	if len(entries) == 0 {
		fmt.Println("No history recorded yet.")
		return
	}
	for _, e := range entries {
		fmt.Printf("%s  %-8s %-30s %s\n", e.Time.Format("2006-01-02 15:04"), e.Command, e.Dir, e.Detail)
	}
}
