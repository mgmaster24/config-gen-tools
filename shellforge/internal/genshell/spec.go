// Package genshell generates a shell init script from a set of selected
// integrations. IntegrationSpecs is the single source of truth for what each
// integration emits and, critically, *where* in the file it has to go.
package genshell

import "github.com/mgmaster24/config-gen-tools/shellforge/internal/config"

// Phase is the ordering bucket an integration's snippet belongs to. This is
// the reason shellforge exists: several of these hooks are order-sensitive
// and get silently broken in a hand-edited rc file.
//
// The ordering rules encoded here:
//
//   - PhasePath first, because version managers rewrite PATH and everything
//     after them must see the result.
//   - PhaseTool next: hooks that only need their own binary resolvable.
//   - PhasePrompt after that, so the prompt sees the final PATH.
//   - PhaseLast for hooks that must be the final thing in the file. zoxide is
//     the canonical case — its own docs require it, and when it isn't last it
//     prints a "configuration issue" warning on every shell startup.
type Phase int

const (
	PhasePath Phase = iota
	PhaseTool
	PhasePrompt
	PhaseLast
)

func (p Phase) String() string {
	switch p {
	case PhasePath:
		return "path"
	case PhaseTool:
		return "tool"
	case PhasePrompt:
		return "prompt"
	case PhaseLast:
		return "last"
	default:
		return "unknown"
	}
}

// IntegrationSpec is everything one integration contributes.
type IntegrationSpec struct {
	// Binary is the command that must exist for the snippet to be useful;
	// it's also what the prereq check looks for.
	Binary string
	Phase  Phase
	// Snippets maps a shell to the lines it emits. An integration with no
	// entry for a shell is skipped for that shell rather than emitting
	// something broken.
	Snippets map[config.Shell][]string
	// Guarded wraps the snippet in a "is the binary present?" test. Worth it
	// for anything a user might uninstall without regenerating: a missing
	// binary then costs a no-op instead of an error on every new shell.
	Guarded bool
}

var IntegrationSpecs = map[config.Integration]IntegrationSpec{
	config.IntegrationMise: {
		Binary: "mise",
		Phase:  PhasePath,
		Snippets: map[config.Shell][]string{
			config.ShellZsh:  {`eval "$(mise activate zsh)"`},
			config.ShellBash: {`eval "$(mise activate bash)"`},
		},
		Guarded: true,
	},
	config.IntegrationFnm: {
		Binary: "fnm",
		Phase:  PhasePath,
		Snippets: map[config.Shell][]string{
			config.ShellZsh:  {`eval "$(fnm env --use-on-cd --shell zsh)"`},
			config.ShellBash: {`eval "$(fnm env --use-on-cd --shell bash)"`},
		},
		Guarded: true,
	},
	config.IntegrationFzf: {
		Binary: "fzf",
		Phase:  PhaseTool,
		Snippets: map[config.Shell][]string{
			config.ShellZsh:  {`source <(fzf --zsh)`},
			config.ShellBash: {`eval "$(fzf --bash)"`},
		},
		Guarded: true,
	},
	config.IntegrationDirenv: {
		Binary: "direnv",
		Phase:  PhaseTool,
		Snippets: map[config.Shell][]string{
			config.ShellZsh:  {`eval "$(direnv hook zsh)"`},
			config.ShellBash: {`eval "$(direnv hook bash)"`},
		},
		Guarded: true,
	},
	config.IntegrationEza: {
		Binary: "eza",
		Phase:  PhaseTool,
		Snippets: map[config.Shell][]string{
			config.ShellZsh: {
				`alias ls='eza --group-directories-first'`,
				`alias ll='eza -l --group-directories-first --git'`,
				`alias tree='eza --tree'`,
			},
			config.ShellBash: {
				`alias ls='eza --group-directories-first'`,
				`alias ll='eza -l --group-directories-first --git'`,
				`alias tree='eza --tree'`,
			},
		},
		Guarded: true,
	},
	config.IntegrationBat: {
		Binary: "bat",
		Phase:  PhaseTool,
		Snippets: map[config.Shell][]string{
			config.ShellZsh:  {`export BAT_THEME="ansi"`, `alias cat='bat --paging=never'`},
			config.ShellBash: {`export BAT_THEME="ansi"`, `alias cat='bat --paging=never'`},
		},
		Guarded: true,
	},
	config.IntegrationStarship: {
		Binary: "starship",
		Phase:  PhasePrompt,
		Snippets: map[config.Shell][]string{
			config.ShellZsh:  {`eval "$(starship init zsh)"`},
			config.ShellBash: {`eval "$(starship init bash)"`},
		},
		Guarded: true,
	},
	config.IntegrationZoxide: {
		Binary: "zoxide",
		Phase:  PhaseLast,
		Snippets: map[config.Shell][]string{
			config.ShellZsh:  {`eval "$(zoxide init zsh)"`},
			config.ShellBash: {`eval "$(zoxide init bash)"`},
		},
		Guarded: true,
	},
}
