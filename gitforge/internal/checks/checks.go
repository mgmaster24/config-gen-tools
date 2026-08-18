// Package checks holds gitforge's prerequisite data. The detection
// machinery lives in forge/prereq; this package only supplies the list.
package checks

import (
	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/gitforge/internal/config"
)

// gitCheck is the one genuinely required prerequisite: without git there is
// nothing for the generated config to configure.
var gitCheck = prereq.Check{
	Name:        "git",
	Description: "the config gitforge generates is only meaningful with git installed",
	Severity:    prereq.SeverityRequired,
	Binary:      "git",
	Hints: []prereq.InstallHint{
		{Manager: prereq.PMBrew, Command: "brew install git"},
		{Manager: prereq.PMApt, Command: "sudo apt install git"},
		{Manager: prereq.PMDnf, Command: "sudo dnf install git"},
		{Manager: prereq.PMPacman, Command: "sudo pacman -S git"},
		{Manager: prereq.PMWinget, Command: "winget install Git.Git"},
	},
}

// ForConfig returns git plus whichever checks the enabled features and
// identities imply.
//
// The delta check is BlocksTooling: unlike shellforge's guarded snippets,
// `core.pager = delta` is unconditional once written, so a missing delta
// breaks every `git diff` rather than degrading quietly.
func ForConfig(cfg config.Config) []prereq.Check {
	out := []prereq.Check{gitCheck}

	if cfg.HasFeature(config.FeatureDelta) {
		out = append(out, prereq.Check{
			Name:          "delta",
			Description:   "sets core.pager; without it every `git diff` fails",
			Severity:      prereq.SeverityRecommended,
			BlocksTooling: true,
			Scope:         string(config.FeatureDelta),
			ScopeLabel:    "delta pager",
			Binary:        "delta",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install git-delta"},
				{Manager: prereq.PMApt, Command: "sudo apt install git-delta"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install git-delta"},
				{Manager: prereq.PMPacman, Command: "sudo pacman -S git-delta"},
			},
		})
	}

	// Signing needs a backend. ssh-keygen ships with OpenSSH; gpg does not
	// always exist, so only check the one actually selected.
	needsSSH, needsGPG := false, false
	for _, id := range cfg.Identities {
		if id.SigningKey == "" {
			continue
		}
		if id.SSHSign {
			needsSSH = true
		} else {
			needsGPG = true
		}
	}
	if needsSSH {
		out = append(out, prereq.Check{
			Name:        "ssh-keygen",
			Description: "required to sign commits with an SSH key",
			Severity:    prereq.SeverityRecommended,
			Scope:       "signing",
			ScopeLabel:  "SSH signing",
			Binary:      "ssh-keygen",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMApt, Command: "sudo apt install openssh-client"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install openssh-clients"},
			},
		})
	}
	if needsGPG {
		out = append(out, prereq.Check{
			Name:        "gpg",
			Description: "required to sign commits with a GPG key",
			Severity:    prereq.SeverityRecommended,
			Scope:       "signing",
			ScopeLabel:  "GPG signing",
			Binary:      "gpg",
			Hints: []prereq.InstallHint{
				{Manager: prereq.PMBrew, Command: "brew install gnupg"},
				{Manager: prereq.PMApt, Command: "sudo apt install gnupg"},
				{Manager: prereq.PMDnf, Command: "sudo dnf install gnupg2"},
			},
		})
	}

	return out
}
