package neovim

import (
	"testing"

	"github.com/mgmaster24/nvimforge/internal/github"
)

func TestSelectAsset_ExactMatch(t *testing.T) {
	assets := []github.Asset{
		{Name: "nvim-macos-arm64.tar.gz"},
		{Name: "nvim-linux-x86_64.tar.gz"},
		{Name: "nvim-win64.zip"},
	}

	got, err := SelectAsset(assets, Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("SelectAsset: %v", err)
	}
	if got.Name != "nvim-macos-arm64.tar.gz" {
		t.Errorf("got %q, want nvim-macos-arm64.tar.gz", got.Name)
	}
}

func TestSelectAsset_FallsBackToOlderNamingScheme(t *testing.T) {
	// Simulates an older Neovim release that only published the
	// unsuffixed "nvim-macos.tar.gz" (pre-Apple-Silicon-split).
	assets := []github.Asset{
		{Name: "nvim-macos.tar.gz"},
		{Name: "nvim-linux64.tar.gz"},
	}

	got, err := SelectAsset(assets, Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("SelectAsset: %v", err)
	}
	if got.Name != "nvim-macos.tar.gz" {
		t.Errorf("got %q, want nvim-macos.tar.gz", got.Name)
	}
}

func TestSelectAsset_NoMatch(t *testing.T) {
	assets := []github.Asset{{Name: "nvim-linux64.tar.gz"}}
	_, err := SelectAsset(assets, Platform{OS: "darwin", Arch: "arm64"})
	if err == nil {
		t.Fatal("expected error when no asset matches the platform")
	}
}

func TestSelectAsset_UnknownPlatform(t *testing.T) {
	assets := []github.Asset{{Name: "nvim-linux64.tar.gz"}}
	_, err := SelectAsset(assets, Platform{OS: "plan9", Arch: "amd64"})
	if err == nil {
		t.Fatal("expected error for a platform with no known asset pattern")
	}
}

func TestCurrentPlatform_MatchesRuntime(t *testing.T) {
	p := CurrentPlatform()
	if p.OS == "" || p.Arch == "" {
		t.Errorf("CurrentPlatform() = %+v, want non-empty OS/Arch", p)
	}
}
