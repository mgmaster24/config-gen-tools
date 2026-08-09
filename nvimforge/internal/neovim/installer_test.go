package neovim

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mgmaster24/nvimforge/internal/github"
)

type fakeReleaseClient struct {
	release github.Release
	err     error
}

func (f fakeReleaseClient) LatestRelease(ctx context.Context, owner, repo string) (github.Release, error) {
	return f.release, f.err
}

func newTestInstaller(t *testing.T, client github.ReleaseClient) *Installer {
	t.Helper()
	dir := t.TempDir()
	return &Installer{
		InstallRoot: filepath.Join(dir, "neovim"),
		BinDir:      filepath.Join(dir, "bin"),
		Client:      client,
		Platform:    Platform{OS: "linux", Arch: "amd64"},
	}
}

func assertNoFilesystemWrites(t *testing.T, inst *Installer) {
	t.Helper()
	if _, err := os.Stat(inst.InstallRoot); !os.IsNotExist(err) {
		t.Errorf("InstallRoot %s should not exist after a dry run (stat err = %v)", inst.InstallRoot, err)
	}
	if _, err := os.Stat(inst.BinDir); !os.IsNotExist(err) {
		t.Errorf("BinDir %s should not exist after a dry run (stat err = %v)", inst.BinDir, err)
	}
}

func TestEnsureInstalled_DryRun_NotInstalled(t *testing.T) {
	client := fakeReleaseClient{release: github.Release{TagName: "v0.10.2"}}
	inst := newTestInstaller(t, client)

	result, err := inst.EnsureInstalled(context.Background(), InstallOptions{DryRun: true})
	if err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	if result.Action != ActionInstalled {
		t.Errorf("Action = %v, want ActionInstalled", result.Action)
	}
	if result.Version.String() != "v0.10.2" {
		t.Errorf("Version = %v, want v0.10.2", result.Version)
	}
	assertNoFilesystemWrites(t, inst)
}

func TestEnsureInstalled_DryRun_AlreadyUpToDate(t *testing.T) {
	client := fakeReleaseClient{release: github.Release{TagName: "v0.10.2"}}
	inst := newTestInstaller(t, client)

	// Simulate a prior real install by writing the pointer file + binary
	// nvimforge's own bookkeeping expects, without going through
	// EnsureInstalled itself.
	versionDir := inst.versionDir(Version{Major: 0, Minor: 10, Patch: 2})
	if err := os.MkdirAll(filepath.Join(versionDir, "bin"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "bin", "nvim"), []byte("fake"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := inst.writeCurrentVersion(Version{Major: 0, Minor: 10, Patch: 2}); err != nil {
		t.Fatalf("writeCurrentVersion: %v", err)
	}

	result, err := inst.EnsureInstalled(context.Background(), InstallOptions{DryRun: true})
	if err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	if result.Action != ActionNoOp {
		t.Errorf("Action = %v, want ActionNoOp", result.Action)
	}
}

func TestEnsureInstalled_DryRun_Outdated(t *testing.T) {
	client := fakeReleaseClient{release: github.Release{TagName: "v0.11.0"}}
	inst := newTestInstaller(t, client)

	versionDir := inst.versionDir(Version{Major: 0, Minor: 10, Patch: 2})
	if err := os.MkdirAll(filepath.Join(versionDir, "bin"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "bin", "nvim"), []byte("fake"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := inst.writeCurrentVersion(Version{Major: 0, Minor: 10, Patch: 2}); err != nil {
		t.Fatalf("writeCurrentVersion: %v", err)
	}

	result, err := inst.EnsureInstalled(context.Background(), InstallOptions{DryRun: true})
	if err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	if result.Action != ActionUpdated {
		t.Errorf("Action = %v, want ActionUpdated", result.Action)
	}
	if result.Version.String() != "v0.11.0" {
		t.Errorf("Version = %v, want v0.11.0", result.Version)
	}
}

func TestEnsureInstalled_ReleaseClientError(t *testing.T) {
	client := fakeReleaseClient{err: context.DeadlineExceeded}
	inst := newTestInstaller(t, client)

	_, err := inst.EnsureInstalled(context.Background(), InstallOptions{DryRun: true})
	if err == nil {
		t.Fatal("expected error to propagate from ReleaseClient")
	}
}

func TestCurrentVersion_NoPointerFile(t *testing.T) {
	inst := newTestInstaller(t, fakeReleaseClient{})
	_, found := inst.CurrentVersion()
	if found {
		t.Error("CurrentVersion should report not-found with no pointer file")
	}
}

func TestCurrentVersion_StalePointerFileWithMissingBinary(t *testing.T) {
	inst := newTestInstaller(t, fakeReleaseClient{})
	// Pointer file exists but the binary it references was never written
	// (e.g. manually deleted) — should report not-found, not error.
	if err := os.MkdirAll(inst.InstallRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := inst.writeCurrentVersion(Version{Major: 0, Minor: 10, Patch: 2}); err != nil {
		t.Fatalf("writeCurrentVersion: %v", err)
	}

	_, found := inst.CurrentVersion()
	if found {
		t.Error("CurrentVersion should report not-found when the pointed-to binary is missing")
	}
}
