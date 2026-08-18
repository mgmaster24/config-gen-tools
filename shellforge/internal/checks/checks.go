// Package checks holds shellforge's prerequisite data: which binaries the
// selected integrations need on the host. The detection machinery lives in
// forge/prereq; this package only supplies the check list.
package checks

import (
	"fmt"

	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/genshell"
)

// hints maps a binary to how to install it, per package manager. Kept as
// data rather than inline in the checks so ForIntegrations can build checks
// straight from IntegrationSpecs and stay in sync automatically.
var hints = map[string][]prereq.InstallHint{
	"mise": {
		{Manager: prereq.PMBrew, Command: "brew install mise"},
		{Manager: prereq.PMApt, Command: "curl https://mise.run | sh"},
	},
	"fnm": {
		{Manager: prereq.PMBrew, Command: "brew install fnm"},
		{Manager: prereq.PMApt, Command: "curl -fsSL https://fnm.vercel.app/install | bash"},
	},
	"fzf": {
		{Manager: prereq.PMBrew, Command: "brew install fzf"},
		{Manager: prereq.PMApt, Command: "sudo apt install fzf"},
		{Manager: prereq.PMDnf, Command: "sudo dnf install fzf"},
		{Manager: prereq.PMPacman, Command: "sudo pacman -S fzf"},
	},
	"direnv": {
		{Manager: prereq.PMBrew, Command: "brew install direnv"},
		{Manager: prereq.PMApt, Command: "sudo apt install direnv"},
		{Manager: prereq.PMDnf, Command: "sudo dnf install direnv"},
	},
	"eza": {
		{Manager: prereq.PMBrew, Command: "brew install eza"},
		{Manager: prereq.PMApt, Command: "sudo apt install eza"},
		{Manager: prereq.PMPacman, Command: "sudo pacman -S eza"},
	},
	"bat": {
		{Manager: prereq.PMBrew, Command: "brew install bat"},
		{Manager: prereq.PMApt, Command: "sudo apt install bat"},
		{Manager: prereq.PMDnf, Command: "sudo dnf install bat"},
	},
	"starship": {
		{Manager: prereq.PMBrew, Command: "brew install starship"},
		{Manager: prereq.PMApt, Command: "curl -sS https://starship.rs/install.sh | sh"},
	},
	"zoxide": {
		{Manager: prereq.PMBrew, Command: "brew install zoxide"},
		{Manager: prereq.PMApt, Command: "sudo apt install zoxide"},
		{Manager: prereq.PMDnf, Command: "sudo dnf install zoxide"},
		{Manager: prereq.PMPacman, Command: "sudo pacman -S zoxide"},
	},
}

// ForIntegrations returns one check per selected integration, derived from
// IntegrationSpecs so a new integration can never be added without its
// prereq check appearing too.
//
// None of these are BlocksTooling: every emitted snippet is guarded by a
// `command -v` test, so a missing binary degrades to a no-op rather than an
// error on shell startup. That's a deliberate contrast with nvimforge, where
// a missing dotnet genuinely prevents mason from installing anything.
func ForIntegrations(integrations []config.Integration) []prereq.Check {
	out := make([]prereq.Check, 0, len(integrations))
	for _, i := range integrations {
		spec, ok := genshell.IntegrationSpecs[i]
		if !ok {
			continue
		}
		out = append(out, prereq.Check{
			Name:        spec.Binary,
			Description: fmt.Sprintf("enables the %s shell integration", i.DisplayName()),
			Severity:    prereq.SeverityRecommended,
			Scope:       string(i),
			ScopeLabel:  i.DisplayName(),
			Binary:      spec.Binary,
			Hints:       hints[spec.Binary],
		})
	}
	return out
}
