package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNeedsBackup(t *testing.T) {
	t.Run("nonexistent path", func(t *testing.T) {
		dir := t.TempDir()
		got, err := NeedsBackup(filepath.Join(dir, "missing"))
		if err != nil {
			t.Fatalf("NeedsBackup: %v", err)
		}
		if got {
			t.Error("want false for nonexistent path")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "empty")
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		got, err := NeedsBackup(target)
		if err != nil {
			t.Fatalf("NeedsBackup: %v", err)
		}
		if got {
			t.Error("want false for empty directory")
		}
	})

	t.Run("directory with nvimforge marker", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "generated")
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(target, "init.lua"), []byte("-- x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		if err := os.WriteFile(filepath.Join(target, GeneratedMarkerName), []byte("{}"), 0o644); err != nil {
			t.Fatalf("WriteFile marker: %v", err)
		}
		got, err := NeedsBackup(target)
		if err != nil {
			t.Fatalf("NeedsBackup: %v", err)
		}
		if got {
			t.Error("want false when GeneratedMarkerName is present")
		}
	})

	t.Run("directory with unrelated content", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "existing-config")
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(target, "init.lua"), []byte("-- hand written"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		got, err := NeedsBackup(target)
		if err != nil {
			t.Fatalf("NeedsBackup: %v", err)
		}
		if !got {
			t.Error("want true for a pre-existing non-generated directory")
		}
	})

	t.Run("plain file", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "somefile")
		if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		got, err := NeedsBackup(target)
		if err != nil {
			t.Fatalf("NeedsBackup: %v", err)
		}
		if !got {
			t.Error("want true for a plain file at deployPath")
		}
	})
}

func TestBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "init.lua"), []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	backupPath, err := Backup(target)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("original path %s should no longer exist, stat err = %v", target, err)
	}
	got, err := os.ReadFile(filepath.Join(backupPath, "init.lua"))
	if err != nil {
		t.Fatalf("ReadFile backup content: %v", err)
	}
	if string(got) != "original" {
		t.Errorf("backup content = %q, want %q", got, "original")
	}
	if filepath.Dir(backupPath) != dir {
		t.Errorf("backupPath = %s, want sibling of %s", backupPath, target)
	}
}
