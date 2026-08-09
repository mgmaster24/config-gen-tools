package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/mgmaster24/nvimforge/internal/config"
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
