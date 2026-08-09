package prereq

import (
	"testing"

	"github.com/mgmaster24/nvimforge/internal/config"
	"github.com/mgmaster24/nvimforge/internal/runner/runnertest"
)

func TestRun_UniversalOnly(t *testing.T) {
	f := runnertest.New().
		WithPath("git", "/usr/bin/git").
		WithPath("cc", "/usr/bin/cc").
		WithPath("make", "/usr/bin/make").
		WithPath("rg", "/usr/bin/rg").
		WithPath("fd", "/usr/bin/fd").
		WithPath("curl", "/usr/bin/curl").
		WithPath("tar", "/usr/bin/tar")

	report := Run(f, nil, "linux")

	if len(report.Results) != len(UniversalChecks) {
		t.Fatalf("got %d results, want %d (no languages selected)", len(report.Results), len(UniversalChecks))
	}
	for _, res := range report.Results {
		if !res.Found {
			t.Errorf("check %q should be found (fake has it on PATH)", res.Check.Name)
		}
	}
	if report.HasMissingRequired() {
		t.Error("HasMissingRequired() should be false when everything is present")
	}
}

func TestRun_MissingRequiredBlocksButRecommendedDoesNot(t *testing.T) {
	// A fake with NOTHING on PATH: every check misses. downloader/archiver
	// are Required; everything else is Recommended.
	f := runnertest.New()

	report := Run(f, nil, "linux")

	if !report.HasMissingRequired() {
		t.Error("HasMissingRequired() should be true: downloader and archiver are missing")
	}

	missing := report.Missing()
	if len(missing) != len(UniversalChecks) {
		t.Errorf("Missing() = %d results, want %d (all universal checks missing)", len(missing), len(UniversalChecks))
	}

	for _, res := range missing {
		if res.Check.Name == "downloader" || res.Check.Name == "archiver" {
			if res.Check.Severity != SeverityRequired {
				t.Errorf("%q should be SeverityRequired", res.Check.Name)
			}
		} else if res.Check.Severity != SeverityRecommended {
			t.Errorf("%q should be SeverityRecommended", res.Check.Name)
		}
	}
}

func TestRun_IncludesOnlySelectedLanguageChecks(t *testing.T) {
	f := runnertest.New().WithPath("go", "/usr/bin/go")

	report := Run(f, []config.Language{config.LangGo}, "linux")

	wantLen := len(UniversalChecks) + len(LanguageChecks[config.LangGo])
	if len(report.Results) != wantLen {
		t.Fatalf("got %d results, want %d", len(report.Results), wantLen)
	}

	var sawRustChecks, sawGoChecks bool
	for _, res := range report.Results {
		switch res.Check.Language {
		case config.LangRust:
			sawRustChecks = true
		case config.LangGo:
			sawGoChecks = true
		}
	}
	if sawRustChecks {
		t.Error("Rust checks should not run when only Go is selected")
	}
	if !sawGoChecks {
		t.Error("Go checks should run when Go is selected")
	}
}

func TestRun_DetectFirstOfMatchesAlternateBinaryName(t *testing.T) {
	// Debian names the fd binary "fdfind"; the fake only has that on PATH.
	f := runnertest.New().WithPath("fdfind", "/usr/bin/fdfind")

	report := Run(f, nil, "linux")

	for _, res := range report.Results {
		if res.Check.Name == "fd" {
			if !res.Found {
				t.Error(`fd check should be satisfied by "fdfind" via detectFirstOf`)
			}
			if res.VersionInfo != "fdfind" {
				t.Errorf("VersionInfo = %q, want %q", res.VersionInfo, "fdfind")
			}
			return
		}
	}
	t.Fatal("fd check not found in results")
}

func TestRun_PopulatesPackageManagers(t *testing.T) {
	f := runnertest.New().WithPath("brew", "/opt/homebrew/bin/brew")
	report := Run(f, nil, "darwin")
	if len(report.PackageManagers) != 1 || report.PackageManagers[0] != PMBrew {
		t.Errorf("PackageManagers = %v, want [brew]", report.PackageManagers)
	}
}
