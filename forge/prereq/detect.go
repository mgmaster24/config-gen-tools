package prereq

import (
	"github.com/mgmaster24/config-gen-tools/forge/runner"
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
// missing — the only condition that should ever make a command
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

// Run detects every check in checks, returning a Report. Callers assemble
// the list themselves — which checks always apply and which are conditional
// is a tool's decision, not this package's. It is pure aside from the
// injected Runner: no printing, no process exit, so it can be unit tested
// and shared between a tool's `doctor` and `install` commands.
func Run(r runner.Runner, checks []Check, goos string) Report {
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
