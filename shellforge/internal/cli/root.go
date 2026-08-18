// Package cli wires shellforge's cobra commands together.
package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/forge/runner"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/checks"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/genshell"
)

// Version is overwritten at release time via ldflags.
var Version = "dev"

func Execute() int {
	root := &cobra.Command{
		Use:           "shellforge",
		Short:         "Generate an ordered shell init script",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newGenerateCmd(), newDoctorCmd(), newVersionCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the shellforge version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "shellforge %s\n", Version)
			return nil
		},
	}
}

func newDoctorCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Report which integration binaries are missing (never installs)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadOrDefault(configPath)
			if err != nil {
				return err
			}
			report := prereq.Run(runner.OSRunner{}, checks.ForIntegrations(cfg.Integrations), runtime.GOOS)
			prereq.RenderText(cmd.OutOrStdout(), report)
			return nil
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "", "path to shellforge.toml")
	return cmd
}

// loadOrDefault returns the config file's contents if one exists, otherwise
// the built-in defaults. Unlike `generate`, the read-only commands never
// prompt.
func loadOrDefault(explicitPath string) (config.Config, error) {
	path, existed, err := config.Resolve(explicitPath)
	if err != nil {
		return config.Config{}, err
	}
	if !existed {
		return config.Default(), nil
	}
	return config.Load(path)
}

// sourceHint prints the one line the user must add to their rc file. It is
// printed on every successful generate, not just the first: the whole design
// depends on that line being present and last.
func sourceHint(cfg config.Config, deployPath string) string {
	return fmt.Sprintf(
		"\nAdd this as the LAST line of your ~/.%src:\n\n  %s\n",
		cfg.Shell, genshell.SourceLine(cfg.Shell, deployPath))
}
