package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgmaster24/nvimforge/internal/github"
	"github.com/mgmaster24/nvimforge/internal/neovim"
)

type fakeReleaseClient struct{ tag string }

func (f fakeReleaseClient) LatestRelease(ctx context.Context, owner, repo string) (github.Release, error) {
	return github.Release{TagName: f.tag}, nil
}

// TestInstallCmd_DryRun_EndToEnd drives the real cobra command wiring
// end-to-end (flags -> resolveInstallConfig -> prereq report -> neovim
// installer -> genconfig) with --dry-run, a fake ReleaseClient, and
// temp-only paths, asserting that nothing touches the network or writes
// to disk, yet the command still reports what it would have done.
func TestInstallCmd_DryRun_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	deployPath := filepath.Join(dir, "nvim-deploy")
	installRoot := filepath.Join(dir, "nvimforge-neovim")
	binDir := filepath.Join(dir, "nvimforge-bin")
	nonexistentConfig := filepath.Join(dir, "does-not-exist.toml")

	origNewInstaller := newNeovimInstaller
	newNeovimInstaller = func() (*neovim.Installer, error) {
		return &neovim.Installer{
			InstallRoot: installRoot,
			BinDir:      binDir,
			Client:      fakeReleaseClient{tag: "v0.99.0"},
			Platform:    neovim.CurrentPlatform(),
		}, nil
	}
	t.Cleanup(func() { newNeovimInstaller = origNewInstaller })

	root := newRootCmd()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{
		"install",
		"--config", nonexistentConfig,
		"--lang", "go",
		"--yes",
		"--dry-run",
		"--deploy-path", deployPath,
		"--no-banner",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("install --dry-run failed: %v\noutput:\n%s", err, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "dry run") {
		t.Errorf("expected dry-run messaging in output, got:\n%s", out)
	}
	if !strings.Contains(out, "v0.99.0") {
		t.Errorf("expected the fake release version to be reported, got:\n%s", out)
	}

	for _, p := range []string{deployPath, installRoot, binDir} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s should not exist after a dry run (stat err = %v)", p, err)
		}
	}
	if _, err := os.Stat(nonexistentConfig); !os.IsNotExist(err) {
		t.Error("no config file should have been written: --yes with --lang skips the save-prompt path")
	}
}
