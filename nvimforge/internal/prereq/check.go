// Package prereq detects (never installs) the system prerequisites the
// generated Neovim config benefits from, plus the two tools nvimforge
// itself needs to fetch and extract a Neovim release. Detection is pure
// (Runner in, Report out); rendering the report is a separate concern.
package prereq

import (
	"github.com/mgmaster24/nvimforge/internal/config"
	"github.com/mgmaster24/nvimforge/internal/runner"
)

// Severity classifies how much a missing Check matters. Only
// SeverityRequired should ever block an nvimforge command; SeverityRecommended
// is report-only, by design (see the project's locked-in "prereqs are
// report-only" decision).
type Severity int

const (
	SeverityRecommended Severity = iota
	SeverityRequired
)

func (s Severity) String() string {
	switch s {
	case SeverityRequired:
		return "required"
	case SeverityRecommended:
		return "recommended"
	default:
		return "unknown"
	}
}

// DetectFunc overrides the default Binary/LookPath detection for a Check,
// e.g. to accept the first of several candidate binary names.
type DetectFunc func(r runner.Runner) (found bool, versionInfo string)

// Check describes one prerequisite to look for on the host system.
type Check struct {
	Name        string
	Description string
	Severity    Severity
	// Language is the zero value ("") for a universal check, or the
	// config.Language it's specific to.
	Language config.Language
	// Binary is looked up via runner.Runner.LookPath when Detect is nil.
	Binary string
	// Detect, when set, overrides Binary-based lookup.
	Detect DetectFunc
	Hints  []InstallHint
	// BlocksTooling marks a check whose absence means mason cannot install
	// this language's tooling at all — as opposed to the usual case, where a
	// missing toolchain only degrades the experience. nvimforge still
	// generates a config either way (prereqs stay report-only), but
	// `install` confirms first rather than silently writing a config that is
	// guaranteed to fail on first launch.
	BlocksTooling bool
}

func (c Check) detect(r runner.Runner) (found bool, versionInfo string) {
	if c.Detect != nil {
		return c.Detect(r)
	}
	_, found = r.LookPath(c.Binary)
	return found, ""
}

// CheckResult is the outcome of running one Check against a Runner.
type CheckResult struct {
	Check       Check
	Found       bool
	VersionInfo string
}

// detectFirstOf returns a DetectFunc that reports found=true at the first
// binary name present on PATH, and returns that binary name as
// versionInfo (useful when a tool goes by different names per distro,
// e.g. "fd" vs "fdfind").
func detectFirstOf(binaries ...string) DetectFunc {
	return func(r runner.Runner) (bool, string) {
		for _, b := range binaries {
			if _, ok := r.LookPath(b); ok {
				return true, b
			}
		}
		return false, ""
	}
}
