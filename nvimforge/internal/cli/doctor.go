package cli

import (
	"runtime"

	"github.com/spf13/cobra"

	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/forge/runner"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/checks"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
)

func newDoctorCmd() *cobra.Command {
	var (
		configPath string
		langFlags  []string
		asJSON     bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Report missing system prerequisites (report-only, never installs anything)",
		RunE: func(cmd *cobra.Command, args []string) error {
			langs, err := resolveDoctorLanguages(configPath, langFlags)
			if err != nil {
				return err
			}

			report := prereq.Run(runner.OSRunner{}, checks.ForLanguages(langs), runtime.GOOS)

			if asJSON {
				if err := prereq.RenderJSON(cmd.OutOrStdout(), report); err != nil {
					return err
				}
			} else {
				prereq.RenderText(cmd.OutOrStdout(), report)
			}

			if report.HasMissingRequired() {
				return &ExitError{Code: 1}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to nvimforge.toml (default: ./nvimforge.toml, then ~/.config/nvimforge/config.toml)")
	cmd.Flags().StringArrayVar(&langFlags, "lang", nil, "language to check (repeatable); overrides the config file")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output machine-readable JSON")

	return cmd
}

// resolveDoctorLanguages determines which languages' prereqs to check:
// explicit --lang flags win; otherwise the resolved config file's
// languages, if one exists; otherwise nil (universal checks only).
func resolveDoctorLanguages(configPath string, langFlags []string) ([]config.Language, error) {
	if len(langFlags) > 0 {
		return parseLanguages(langFlags)
	}

	resolvedPath, existed, err := config.Resolve(configPath)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	cfg, err := config.Load(resolvedPath)
	if err != nil {
		return nil, err
	}
	return cfg.Languages, nil
}

// parseLanguages validates and converts raw --lang flag values.
func parseLanguages(raw []string) ([]config.Language, error) {
	langs := make([]config.Language, 0, len(raw))
	for _, s := range raw {
		l, err := config.ParseLanguage(s)
		if err != nil {
			return nil, err
		}
		langs = append(langs, l)
	}
	return langs, nil
}
