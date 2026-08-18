package neovim

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mgmaster24/config-gen-tools/forge/fsutil"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/github"
)

// currentVersionFile records which version directory under InstallRoot is
// active. A plain pointer file is used instead of a symlink so nothing
// here depends on symlink privileges (notably absent for unprivileged
// users on Windows).
const currentVersionFile = "CURRENT_VERSION"

// Action describes what EnsureInstalled did.
type Action int

const (
	ActionNoOp Action = iota
	ActionInstalled
	ActionUpdated
)

func (a Action) String() string {
	switch a {
	case ActionNoOp:
		return "no-op (already up to date)"
	case ActionInstalled:
		return "installed"
	case ActionUpdated:
		return "updated"
	default:
		return "unknown"
	}
}

// InstallResult is the outcome of EnsureInstalled.
type InstallResult struct {
	Action  Action
	Version Version
}

// InstallOptions configures one EnsureInstalled call.
type InstallOptions struct {
	// DryRun reports what would happen without touching the network or
	// filesystem.
	DryRun bool
}

// Installer manages an nvimforge-private Neovim installation. No step it
// performs ever requires sudo/admin privileges: everything lives under
// InstallRoot and BinDir, both expected to be user-writable.
type Installer struct {
	InstallRoot string // e.g. ~/.local/share/nvimforge/neovim
	BinDir      string // e.g. ~/.local/bin
	Client      github.ReleaseClient
	Platform    Platform
}

func (i *Installer) versionDir(v Version) string {
	return filepath.Join(i.InstallRoot, v.String())
}

func (i *Installer) binaryName() string {
	if i.Platform.OS == "windows" {
		return "nvim.exe"
	}
	return "nvim"
}

func (i *Installer) readCurrentVersion() (Version, bool) {
	data, err := os.ReadFile(filepath.Join(i.InstallRoot, currentVersionFile))
	if err != nil {
		return Version{}, false
	}
	v, err := ParseVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return Version{}, false
	}
	return v, true
}

func (i *Installer) writeCurrentVersion(v Version) error {
	path := filepath.Join(i.InstallRoot, currentVersionFile)
	return fsutil.AtomicWriteFile(path, []byte(v.String()+"\n"), 0o644)
}

// CurrentVersion reports the version of the nvimforge-managed Neovim
// install, if any. It trusts InstallRoot's own pointer file rather than
// resolving `nvim` on PATH, so the result is deterministic and independent
// of shell state or a separately-installed system Neovim. If the pointer
// file exists but the binary it names is missing (e.g. deleted by hand),
// this reports not-found rather than erroring, so EnsureInstalled just
// reinstalls.
func (i *Installer) CurrentVersion() (Version, bool) {
	v, ok := i.readCurrentVersion()
	if !ok {
		return Version{}, false
	}
	binPath := filepath.Join(i.versionDir(v), "bin", i.binaryName())
	if _, err := os.Stat(binPath); err != nil {
		return Version{}, false
	}
	return v, true
}

// EnsureInstalled installs the latest neovim/neovim release if none is
// currently managed by nvimforge, or updates it if a newer release exists.
// Already-up-to-date is a no-op, so repeated calls are safe.
func (i *Installer) EnsureInstalled(ctx context.Context, opts InstallOptions) (InstallResult, error) {
	release, err := i.Client.LatestRelease(ctx, "neovim", "neovim")
	if err != nil {
		return InstallResult{}, fmt.Errorf("fetching latest neovim release: %w", err)
	}
	latest, err := ParseVersion(release.TagName)
	if err != nil {
		return InstallResult{}, fmt.Errorf("parsing latest release tag %q: %w", release.TagName, err)
	}

	current, found := i.CurrentVersion()
	if found && current.Compare(latest) == 0 {
		return InstallResult{Action: ActionNoOp, Version: current}, nil
	}

	action := ActionInstalled
	if found {
		action = ActionUpdated
	}
	if opts.DryRun {
		return InstallResult{Action: action, Version: latest}, nil
	}

	asset, err := SelectAsset(release.Assets, i.Platform)
	if err != nil {
		return InstallResult{}, err
	}

	destDir := i.versionDir(latest)
	if err := i.downloadAndExtract(ctx, release, asset, destDir); err != nil {
		return InstallResult{}, err
	}

	if err := i.writeCurrentVersion(latest); err != nil {
		return InstallResult{}, fmt.Errorf("recording installed version: %w", err)
	}

	if err := i.linkBin(latest); err != nil {
		return InstallResult{}, fmt.Errorf("linking nvim into %s: %w", i.BinDir, err)
	}

	return InstallResult{Action: action, Version: latest}, nil
}
