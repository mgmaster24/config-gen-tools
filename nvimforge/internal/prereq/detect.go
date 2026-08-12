package prereq

import (
	"github.com/mgmaster24/nvimforge/internal/config"
	"github.com/mgmaster24/nvimforge/internal/runner"
)

// Report is the outcome of running every relevant Check.
type Report struct {
	OS              string
	PackageManagers []PackageManager
	Results         []CheckResult
}

// Missing returns every CheckResult that wasn't found.
func (r Report) Missing() []CheckResult {
	var missing []CheckResult
	for _, res := range r.Results {
		if !res.Found {
			missing = append(missing, res)
		}
	}
	return missing
}

// HasMissingRequired reports whether any SeverityRequired check is
// missing — the only condition that should ever make an nvimforge command
// fail based on this report.
func (r Report) HasMissingRequired() bool {
	for _, res := range r.Results {
		if !res.Found && res.Check.Severity == SeverityRequired {
			return true
		}
	}
	return false
}

// MissingBlocking returns every missing check marked BlocksTooling — the
// languages whose mason tooling cannot install on this machine.
func (r Report) MissingBlocking() []CheckResult {
	var blocking []CheckResult
	for _, res := range r.Results {
		if !res.Found && res.Check.BlocksTooling {
			blocking = append(blocking, res)
		}
	}
	return blocking
}

// HasMissingBlocking reports whether any BlocksTooling check is missing.
func (r Report) HasMissingBlocking() bool {
	return len(r.MissingBlocking()) > 0
}

// Run performs every UniversalCheck plus the LanguageChecks for langs,
// returning a Report. It is pure aside from the injected Runner: no
// printing, no process exit, so it can be unit tested and reused by both
// `nvimforge doctor` and `nvimforge install`.
func Run(r runner.Runner, langs []config.Language, goos string) Report {
	checks := make([]Check, 0, len(UniversalChecks))
	checks = append(checks, UniversalChecks...)
	for _, lang := range langs {
		checks = append(checks, LanguageChecks[lang]...)
	}

	results := make([]CheckResult, 0, len(checks))
	for _, c := range checks {
		found, versionInfo := c.detect(r)
		results = append(results, CheckResult{Check: c, Found: found, VersionInfo: versionInfo})
	}

	return Report{
		OS:              goos,
		PackageManagers: DetectPackageManagers(r, goos),
		Results:         results,
	}
}
