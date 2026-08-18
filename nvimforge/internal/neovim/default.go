package neovim

import (
	"fmt"

	"github.com/mgmaster24/config-gen-tools/forge/fsutil"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/github"
)

// Default install locations, following the XDG-ish convention of keeping
// nvimforge-managed state under the user's home directory rather than a
// system path that would require elevated privileges to write to.
const (
	defaultInstallRootRel = "~/.local/share/nvimforge/neovim"
	defaultBinDirRel      = "~/.local/bin"
)

// NewInstaller returns an Installer configured with nvimforge's default
// install locations and the current platform.
func NewInstaller(client github.ReleaseClient) (*Installer, error) {
	installRoot, err := fsutil.ExpandHome(defaultInstallRootRel)
	if err != nil {
		return nil, fmt.Errorf("resolving install root: %w", err)
	}
	binDir, err := fsutil.ExpandHome(defaultBinDirRel)
	if err != nil {
		return nil, fmt.Errorf("resolving bin dir: %w", err)
	}
	return &Installer{
		InstallRoot: installRoot,
		BinDir:      binDir,
		Client:      client,
		Platform:    CurrentPlatform(),
	}, nil
}
