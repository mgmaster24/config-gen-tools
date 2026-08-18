package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"bare tilde", "~", home},
		{"tilde slash path", "~/.config/nvim", filepath.Join(home, ".config/nvim")},
		{"absolute path unchanged", "/etc/hosts", "/etc/hosts"},
		{"relative path unchanged", "relative/path", "relative/path"},
		{"embedded tilde not expanded", "/foo/~/bar", "/foo/~/bar"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ExpandHome(tc.in)
			if err != nil {
				t.Fatalf("ExpandHome(%q) error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
