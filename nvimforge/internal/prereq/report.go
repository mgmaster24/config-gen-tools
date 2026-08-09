package prereq

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
)

// RenderText writes a human-readable report to w.
func RenderText(w io.Writer, report Report) {
	fmt.Fprintf(w, "Prerequisite check (%s):\n\n", report.OS)

	anyMissing := false
	for _, res := range report.Results {
		status := "ok"
		if !res.Found {
			status = "missing"
			anyMissing = true
		}
		fmt.Fprintf(w, "  [%-7s] %-14s %s\n", status, res.Check.Name, res.Check.Description)
	}

	if !anyMissing {
		fmt.Fprintln(w, "\nAll checks passed.")
		return
	}

	fmt.Fprintln(w, "\nMissing:")
	for _, res := range report.Missing() {
		fmt.Fprintf(w, "\n  %s (%s)\n", res.Check.Name, res.Check.Severity)
		for _, h := range filterHints(res.Check.Hints, report.PackageManagers) {
			fmt.Fprintf(w, "    %s: %s\n", h.Manager, h.Command)
		}
	}

	if report.HasMissingRequired() {
		fmt.Fprintln(w, "\nnvimforge cannot continue until the required items above are installed.")
	}
}

// filterHints narrows hints down to the package managers actually detected
// on the machine, falling back to the full hint list when none were
// detected (e.g. a minimal container with no package manager on PATH).
func filterHints(hints []InstallHint, available []PackageManager) []InstallHint {
	if len(available) == 0 {
		return hints
	}
	var filtered []InstallHint
	for _, h := range hints {
		if slices.Contains(available, h.Manager) {
			filtered = append(filtered, h)
		}
	}
	if len(filtered) == 0 {
		return hints
	}
	return filtered
}

type jsonReport struct {
	OS                 string            `json:"os"`
	PackageManagers    []PackageManager  `json:"package_managers"`
	Results            []jsonCheckResult `json:"results"`
	HasMissingRequired bool              `json:"has_missing_required"`
}

type jsonCheckResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Language    string `json:"language,omitempty"`
	Found       bool   `json:"found"`
	VersionInfo string `json:"version_info,omitempty"`
}

// RenderJSON writes a machine-readable report to w.
func RenderJSON(w io.Writer, report Report) error {
	jr := jsonReport{
		OS:                 report.OS,
		PackageManagers:    report.PackageManagers,
		HasMissingRequired: report.HasMissingRequired(),
	}
	for _, res := range report.Results {
		jr.Results = append(jr.Results, jsonCheckResult{
			Name:        res.Check.Name,
			Description: res.Check.Description,
			Severity:    res.Check.Severity.String(),
			Language:    string(res.Check.Language),
			Found:       res.Found,
			VersionInfo: res.VersionInfo,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jr)
}
