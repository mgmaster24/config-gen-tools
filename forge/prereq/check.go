// Package prereq detects (never installs) system prerequisites: the tools a
// forge tool itself needs, plus whatever the config it generates depends on.
// Detection is pure (Runner in, Report out); rendering the report is a
// separate concern, and assembling the check list is the caller's.
package prereq

import (
	"github.com/mgmaster24/config-gen-tools/forge/runner"
)

// Severity classifies how much a missing Check matters. Only
// SeverityRequired should ever block a command; SeverityRecommended
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
	// Scope is the zero value ("") for a check that always applies, or a
	// tool-defined key the check is specific to — a language for nvimforge, a
	// shell plugin for shellforge. It stays a plain string so this package
	// doesn't depend on any one tool's domain types.
	Scope string
	// ScopeLabel is the human-readable form of Scope (e.g. "C#" for
	// "csharp"). Rendering falls back to Scope when it's empty.
	ScopeLabel string
	// Binary is looked up via runner.Runner.LookPath when Detect is nil.
	Binary string
	// Detect, when set, overrides Binary-based lookup.
	Detect DetectFunc
	Hints  []InstallHint
	// BlocksTooling marks a check whose absence means this scope's tooling
	// cannot be installed at all — as opposed to the usual case, where a
	// missing toolchain only degrades the experience. A tool still generates
	// a config either way (prereqs stay report-only), but `install` confirms
	// first rather than silently writing a config that is guaranteed to fail
	// on first use.
	BlocksTooling bool
}

// scopeLabel returns the display form of a Check's scope, falling back to
// the raw Scope key when no label was supplied.
func (c Check) scopeLabel() string {
	if c.ScopeLabel != "" {
		return c.ScopeLabel
	}
	return c.Scope
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

// DetectFirstOf returns a DetectFunc reporting found at the first binary
// present on PATH, returning that name as versionInfo (useful when a tool
// goes by different names per distro, e.g. "fd" vs "fdfind").
func DetectFirstOf(binaries ...string) DetectFunc {
	return func(r runner.Runner) (bool, string) {
		for _, b := range binaries {
			if _, ok := r.LookPath(b); ok {
				return true, b
			}
		}
		return false, ""
	}
}
