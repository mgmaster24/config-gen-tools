// Package fsutil provides small, dependency-free filesystem helpers shared
// by config loading and generated-config deployment.
package fsutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandHome resolves a leading "~" or "~/..." in path to the current user's
// home directory. Paths without a leading "~" are returned unchanged.
func ExpandHome(path string) (string, error) {
	if path == "" || path == "~" {
		if path == "" {
			return path, nil
		}
		return os.UserHomeDir()
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
