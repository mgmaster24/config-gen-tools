package config

import (
	"fmt"
	"slices"
	"strings"
)

// Shell is the shell dialect to generate for. The set is closed: anything
// not in AllShells is invalid.
type Shell string

const (
	ShellZsh  Shell = "zsh"
	ShellBash Shell = "bash"
)

var AllShells = []Shell{ShellZsh, ShellBash}

func (s Shell) Valid() bool { return slices.Contains(AllShells, s) }

// InitFileName is the name of the file shellforge generates for this shell.
func (s Shell) InitFileName() string {
	return "init." + string(s)
}

// ParseShell normalizes and validates a user-supplied shell name.
func ParseShell(v string) (Shell, error) {
	s := Shell(strings.ToLower(strings.TrimSpace(v)))
	if !s.Valid() {
		return "", fmt.Errorf("unknown shell %q (valid: zsh, bash)", v)
	}
	return s, nil
}

// Integration identifies one tool whose shell hook shellforge can emit.
type Integration string

const (
	IntegrationMise     Integration = "mise"
	IntegrationFnm      Integration = "fnm"
	IntegrationFzf      Integration = "fzf"
	IntegrationDirenv   Integration = "direnv"
	IntegrationEza      Integration = "eza"
	IntegrationBat      Integration = "bat"
	IntegrationStarship Integration = "starship"
	IntegrationZoxide   Integration = "zoxide"
)

// AllIntegrations lists every integration, in the order they're presented
// interactively. This is *not* the order they're emitted in — that's decided
// by each integration's Phase (see internal/genshell).
var AllIntegrations = []Integration{
	IntegrationMise,
	IntegrationFnm,
	IntegrationFzf,
	IntegrationDirenv,
	IntegrationEza,
	IntegrationBat,
	IntegrationStarship,
	IntegrationZoxide,
}

// DefaultIntegrations is what a fresh run enables: the widely-used,
// low-surprise ones. Version managers (mise, fnm) stay opt-in because they
// take over PATH resolution.
var DefaultIntegrations = []Integration{
	IntegrationFzf,
	IntegrationEza,
	IntegrationBat,
	IntegrationStarship,
	IntegrationZoxide,
}

var integrationDisplayNames = map[Integration]string{
	IntegrationMise:     "mise",
	IntegrationFnm:      "fnm",
	IntegrationFzf:      "fzf",
	IntegrationDirenv:   "direnv",
	IntegrationEza:      "eza",
	IntegrationBat:      "bat",
	IntegrationStarship: "starship",
	IntegrationZoxide:   "zoxide",
}

func (i Integration) Valid() bool { return slices.Contains(AllIntegrations, i) }

func (i Integration) DisplayName() string {
	if n, ok := integrationDisplayNames[i]; ok {
		return n
	}
	return string(i)
}

// ParseIntegration normalizes and validates a user-supplied integration name.
func ParseIntegration(v string) (Integration, error) {
	i := Integration(strings.ToLower(strings.TrimSpace(v)))
	if !i.Valid() {
		names := make([]string, len(AllIntegrations))
		for n, known := range AllIntegrations {
			names[n] = string(known)
		}
		return "", fmt.Errorf("unknown integration %q (valid: %s)", v, strings.Join(names, ", "))
	}
	return i, nil
}
