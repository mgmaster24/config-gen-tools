package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GeneratedMarkerName is written into a deploy directory by genconfig.Write
// after a successful generation. Its presence means the directory holds
// nvimforge's own prior output rather than a user's hand-written config.
const GeneratedMarkerName = ".nvimforge-generated"

// NeedsBackup reports whether deployPath holds pre-existing content that
// nvimforge did not itself generate, and therefore should be preserved
// before being overwritten. A missing path, an empty directory, or a
// directory carrying GeneratedMarkerName all report false.
func NeedsBackup(deployPath string) (bool, error) {
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

	if _, err := os.Stat(filepath.Join(deployPath, GeneratedMarkerName)); err == nil {
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
