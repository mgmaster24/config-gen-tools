// Package neovim installs and updates an nvimforge-private copy of Neovim,
// independent of any system-installed nvim that might also be on PATH.
package neovim

import (
	"fmt"
	"runtime"

	"github.com/mgmaster24/nvimforge/internal/github"
)

// Platform identifies the OS/architecture combination Neovim is being
// installed for.
type Platform struct {
	OS   string // runtime.GOOS values: "darwin", "linux", "windows"
	Arch string // runtime.GOARCH values: "amd64", "arm64"
}

// CurrentPlatform returns the Platform this nvimforge binary is running on.
func CurrentPlatform() Platform {
	return Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

// assetCandidates lists, per Platform, exact release asset names to try in
// order. Neovim's asset naming has changed across releases (e.g.
// nvim-macos.tar.gz vs nvim-macos-x86_64.tar.gz vs nvim-macos-arm64.tar.gz),
// so this is an ordered, tolerant list rather than one hardcoded name —
// adding a new naming scheme later is a one-line change here.
var assetCandidates = map[Platform][]string{
	{OS: "darwin", Arch: "arm64"}:  {"nvim-macos-arm64.tar.gz", "nvim-macos.tar.gz"},
	{OS: "darwin", Arch: "amd64"}:  {"nvim-macos-x86_64.tar.gz", "nvim-macos.tar.gz"},
	{OS: "linux", Arch: "amd64"}:   {"nvim-linux-x86_64.tar.gz", "nvim-linux64.tar.gz"},
	{OS: "linux", Arch: "arm64"}:   {"nvim-linux-arm64.tar.gz"},
	{OS: "windows", Arch: "amd64"}: {"nvim-win64.zip", "nvim-windows-x86_64.zip"},
}

// SelectAsset picks the release asset matching p from assets, trying each
// of p's candidate names in priority order.
func SelectAsset(assets []github.Asset, p Platform) (github.Asset, error) {
	candidates, ok := assetCandidates[p]
	if !ok {
		return github.Asset{}, fmt.Errorf("nvimforge has no known Neovim release asset pattern for platform %s/%s", p.OS, p.Arch)
	}

	for _, candidate := range candidates {
		for _, a := range assets {
			if a.Name == candidate {
				return a, nil
			}
		}
	}

	names := make([]string, len(assets))
	for i, a := range assets {
		names[i] = a.Name
	}
	return github.Asset{}, fmt.Errorf(
		"no release asset matched platform %s/%s (tried %v; release has: %v) — neovim's release naming may have changed",
		p.OS, p.Arch, candidates, names,
	)
}
