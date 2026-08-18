// Package cli wires gitforge's cobra commands together.
package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/forge/runner"
	"github.com/mgmaster24/config-gen-tools/gitforge/internal/checks"
	"github.com/mgmaster24/config-gen-tools/gitforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/gitforge/internal/gengit"
)

// Version is overwritten at release time via ldflags.
var Version = "dev"

// exitError carries a specific exit code without printing twice.
type exitError struct{ code int }

func (e *exitError) Error() string { return fmt.Sprintf("exit status %d", e.code) }

func Execute() int {
	root := &cobra.Command{
		Use:           "gitforge",
		Short:         "Generate an includable gitconfig with directory-scoped identities",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newGenerateCmd(), newDoctorCmd(), newVersionCmd())

	if err := root.Execute(); err != nil {
		var ee *exitError
		if ok := asExitError(err, &ee); ok {
			return ee.code
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func asExitError(err error, target **exitError) bool {
	if e, ok := err.(*exitError); ok {
		*target = e
		return true
	}
	return false
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the gitforge version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "gitforge %s\n", Version)
			return nil
		},
	}
}

func newDoctorCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Report missing prerequisites (never installs anything)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadOrDefault(configPath)
			if err != nil {
				return err
			}
			report := prereq.Run(runner.OSRunner{}, checks.ForConfig(cfg), runtime.GOOS)
			prereq.RenderText(cmd.OutOrStdout(), report)
			if report.HasMissingRequired() {
				return &exitError{code: 1}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "", "path to gitforge.toml")
	return cmd
}

// loadOrDefault returns the config file if one exists, else the defaults.
// Defaults have no identities, so callers that need a valid config must
// still Validate.
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

func newGenerateCmd() *cobra.Command {
	var (
		configPath     string
		deployPathFlag string
		backupFlag     bool
		dryRun         bool
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate the gitconfig files",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()

			cfg, err := loadOrDefault(configPath)
			if err != nil {
				return err
			}
			if deployPathFlag != "" {
				cfg.DeployPath = deployPathFlag
			}
			if cmd.Flags().Changed("backup") {
				cfg.Backup = backupFlag
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("%w\n\nwrite a %s describing your identities; see the README for an example",
					err, config.FileName)
			}

			report := prereq.Run(runner.OSRunner{}, checks.ForConfig(cfg), runtime.GOOS)
			prereq.RenderText(out, report)
			fmt.Fprintln(out)
			if report.HasMissingRequired() {
				return &exitError{code: 1}
			}

			deployPath, err := cfg.ExpandedDeployPath()
			if err != nil {
				return err
			}

			files, err := gengit.Render(cfg, gengit.BuildTemplateData(cfg, cfg.DeployPath))
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Fprintf(out, "[dry run] would write %d file(s) to %s\n", len(files), deployPath)
				for _, f := range files {
					fmt.Fprintf(out, "\n--- %s ---\n%s", f.RelPath, f.Content)
				}
			} else {
				backedUpTo, err := gengit.Write(files, deployPath, cfg.Backup)
				if err != nil {
					return err
				}
				if backedUpTo != "" {
					fmt.Fprintf(out, "Backed up existing content to %s\n", backedUpTo)
				}
				fmt.Fprintf(out, "Wrote %d file(s) to %s\n", len(files), deployPath)
			}

			fmt.Fprintf(out, "\nAdd this to your ~/.gitconfig:\n\n%s\n",
				gengit.IncludeSnippet(cfg.DeployPath))
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to gitforge.toml")
	cmd.Flags().StringVar(&deployPathFlag, "deploy-path", "", "where to write the generated files")
	cmd.Flags().BoolVar(&backupFlag, "backup", true, "back up existing content before overwriting")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be written without writing it")
	return cmd
}
