package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
)

// PromptConfig runs an interactive form asking which languages to set up,
// where to deploy the generated Neovim config, and the small UI
// preferences (backup, banner). Fields in defaults that aren't asked
// about are carried through unchanged.
func PromptConfig(defaults config.Config) (config.Config, error) {
	cfg := defaults

	selected := append([]config.Language{}, cfg.Languages...)
	deployPath := cfg.DeployPath
	backup := cfg.Backup
	showBanner := cfg.ShowBanner

	langOptions := make([]huh.Option[config.Language], 0, len(config.AllLanguages))
	for _, l := range config.AllLanguages {
		langOptions = append(langOptions, huh.NewOption(l.DisplayName(), l))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[config.Language]().
				Title("Which languages should nvimforge set up?").
				Options(langOptions...).
				Value(&selected),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Where should the Neovim config be deployed?").
				Value(&deployPath),
			huh.NewConfirm().
				Title("Back up an existing config at that path before overwriting?").
				Value(&backup),
			huh.NewConfirm().
				Title("Show the nvimforge banner on startup?").
				Value(&showBanner),
		),
	)

	if err := form.Run(); err != nil {
		return config.Config{}, fmt.Errorf("prompt: %w", err)
	}
	if len(selected) == 0 {
		return config.Config{}, fmt.Errorf("at least one language must be selected")
	}

	cfg.Languages = selected
	cfg.DeployPath = deployPath
	cfg.Backup = backup
	cfg.ShowBanner = showBanner
	return cfg, nil
}

// ConfirmSave asks whether to save cfg to path, defaulting to yes.
// ConfirmContinueWithBlocking asks whether to generate a config anyway when
// a prerequisite is missing that will stop mason installing some language's
// tooling. tools names the missing prerequisites (e.g. "dotnet"). It
// defaults to false: silently writing a config that can't finish installing
// is the outcome worth an extra keystroke to avoid.
func ConfirmContinueWithBlocking(tools []string) (bool, error) {
	proceed := false
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Missing %s. Generate the config anyway?", strings.Join(tools, ", "))).
				Description("The affected language's tooling won't install until it's on your PATH.").
				Value(&proceed),
		),
	).Run()
	if err != nil {
		return false, fmt.Errorf("prompt: %w", err)
	}
	return proceed, nil
}

func ConfirmSave(path string) (bool, error) {
	save := true
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Save this configuration to %s?", path)).
				Value(&save),
		),
	).Run()
	if err != nil {
		return false, fmt.Errorf("prompt: %w", err)
	}
	return save, nil
}
