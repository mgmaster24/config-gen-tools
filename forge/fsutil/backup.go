package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MarkerName returns the marker filename a tool writes into a deploy
// directory after a successful generation, e.g. ".nvimforge-generated".
// Its presence means the directory holds that tool's own prior output
// rather than a user's hand-written config. Each tool has its own marker so
// one forge tool never mistakes another's output for its own.
func MarkerName(tool string) string {
	return "." + tool + "-generated"
}

// NeedsBackup reports whether deployPath holds pre-existing content the
// calling tool did not itself generate, and therefore should be preserved
// before being overwritten. A missing path, an empty directory, or a
// directory carrying marker all report false.
func NeedsBackup(deployPath, marker string) (bool, error) {
	info, err := os.Stat(deployPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", deployPath, err)
	}
	if !info.IsDir() {
		return true, nil
	}

	entries, err := os.ReadDir(deployPath)
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", deployPath, err)
	}
	if len(entries) == 0 {
		return false, nil
	}

	if _, err := os.Stat(filepath.Join(deployPath, marker)); err == nil {
		return false, nil
	}
	return true, nil
}

// Backup renames deployPath to a timestamped sibling path and returns it.
func Backup(deployPath string) (string, error) {
	backupPath := fmt.Sprintf("%s.%s.bak", deployPath, time.Now().Format("20060102-150405"))
	if err := os.Rename(deployPath, backupPath); err != nil {
		return "", fmt.Errorf("backing up %s: %w", deployPath, err)
	}
	return backupPath, nil
}
