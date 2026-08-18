package prereq

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func fakeReport() Report {
	return Report{
		OS:              "linux",
		PackageManagers: []PackageManager{PMApt},
		Results: []CheckResult{
			{Check: Check{Name: "git", Description: "vcs", Severity: SeverityRecommended}, Found: true},
			{
				Check: Check{
					Name: "ripgrep", Description: "grep tool", Severity: SeverityRecommended,
					Hints: []InstallHint{
						{PMApt, "sudo apt install ripgrep"},
						{PMBrew, "brew install ripgrep"},
					},
				},
				Found: false,
			},
			{
				Check: Check{
					Name: "downloader", Description: "fetches releases", Severity: SeverityRequired,
					Hints: []InstallHint{{PMApt, "sudo apt install curl"}},
				},
				Found: false,
			},
		},
	}
}

func TestRenderText_AllPassing(t *testing.T) {
	report := Report{
		OS: "darwin",
		Results: []CheckResult{
			{Check: Check{Name: "git", Description: "vcs"}, Found: true},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report)
	out := buf.String()

	if !strings.Contains(out, "All checks passed.") {
		t.Errorf("expected pass message, got:\n%s", out)
	}
}

func TestRenderText_MissingItemsShowFilteredHintsAndBlockMessage(t *testing.T) {
	var buf bytes.Buffer
	RenderText(&buf, fakeReport())
	out := buf.String()

	if !strings.Contains(out, "ripgrep") {
		t.Error("expected ripgrep to be listed as missing")
	}
	if !strings.Contains(out, "sudo apt install ripgrep") {
		t.Error("expected the apt hint to be shown (apt is the detected package manager)")
	}
	if strings.Contains(out, "brew install ripgrep") {
		t.Error("brew hint should be filtered out: brew wasn't detected on this machine")
	}
	if !strings.Contains(out, "Cannot continue") {
		t.Error("expected a blocking message since downloader (Required) is missing")
	}
}

func TestFilterHints_FallsBackToAllWhenNoneDetected(t *testing.T) {
	hints := []InstallHint{{PMBrew, "brew install x"}, {PMApt, "apt install x"}}
	got := filterHints(hints, nil)
	if len(got) != 2 {
		t.Errorf("expected fallback to all hints when no package managers detected, got %v", got)
	}
}

func TestRenderJSON_RoundTrips(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderJSON(&buf, fakeReport()); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	var decoded jsonReport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.OS != "linux" {
		t.Errorf("OS = %q, want %q", decoded.OS, "linux")
	}
	if !decoded.HasMissingRequired {
		t.Error("HasMissingRequired should be true (downloader missing)")
	}
	if len(decoded.Results) != 3 {
		t.Fatalf("got %d results, want 3", len(decoded.Results))
	}
}
