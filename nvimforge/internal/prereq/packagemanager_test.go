package prereq

import (
	"reflect"
	"testing"

	"github.com/mgmaster24/nvimforge/internal/runner/runnertest"
)

func TestDetectPackageManagers(t *testing.T) {
	cases := []struct {
		name  string
		goos  string
		found map[string]string // binaries present on the fake PATH
		want  []PackageManager
	}{
		{
			name:  "darwin with brew",
			goos:  "darwin",
			found: map[string]string{"brew": "/opt/homebrew/bin/brew"},
			want:  []PackageManager{PMBrew},
		},
		{
			name:  "darwin without brew",
			goos:  "darwin",
			found: map[string]string{},
			want:  nil,
		},
		{
			name:  "linux with apt and brew",
			goos:  "linux",
			found: map[string]string{"apt-get": "/usr/bin/apt-get", "brew": "/home/linuxbrew/bin/brew"},
			want:  []PackageManager{PMApt, PMBrew},
		},
		{
			name:  "linux with pacman only",
			goos:  "linux",
			found: map[string]string{"pacman": "/usr/bin/pacman"},
			want:  []PackageManager{PMPacman},
		},
		{
			name:  "windows with scoop",
			goos:  "windows",
			found: map[string]string{"scoop": `C:\scoop\shims\scoop.exe`},
			want:  []PackageManager{PMScoop},
		},
		{
			name:  "unknown goos",
			goos:  "plan9",
			found: map[string]string{"brew": "/opt/homebrew/bin/brew"},
			want:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := runnertest.New()
			for bin, path := range tc.found {
				f.WithPath(bin, path)
			}
			got := DetectPackageManagers(f, tc.goos)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("DetectPackageManagers(%q) = %v, want %v", tc.goos, got, tc.want)
			}
		})
	}
}
