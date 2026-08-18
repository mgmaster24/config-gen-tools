package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/forge/runner"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/checks"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/shellforge/internal/genshell"
)

func newGenerateCmd() *cobra.Command {
	var (
		configPath       string
		shellFlag        string
		integrationFlags []string
		deployPathFlag   string
		backupFlag       bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate the shell init script",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()

			cfg, err := resolveGenerateConfig(generateOptions{
				ConfigPath:       configPath,
				ShellFlag:        shellFlag,
				IntegrationFlags: integrationFlags,
				DeployPathFlag:   deployPathFlag,
				BackupChanged:    cmd.Flags().Changed("backup"),
				BackupFlag:       backupFlag,
			})
			if err != nil {
				return err
			}

			report := prereq.Run(runner.OSRunner{}, checks.ForIntegrations(cfg.Integrations), runtime.GOOS)
			prereq.RenderText(out, report)
			fmt.Fprintln(out)

			deployPath, err := cfg.ExpandedDeployPath()
			if err != nil {
				return err
			}

			data := genshell.BuildTemplateData(cfg)
			files, err := genshell.Render(cfg.Shell, data)
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Fprintf(out, "[dry run] would write %d file(s) to %s\n", len(files), deployPath)
				for _, f := range files {
					fmt.Fprintf(out, "\n--- %s ---\n%s", f.RelPath, f.Content)
				}
				fmt.Fprint(out, sourceHint(cfg, deployPath))
				return nil
			}

			backedUpTo, err := genshell.Write(files, deployPath, cfg.Backup)
			if err != nil {
				return err
			}
			if backedUpTo != "" {
				fmt.Fprintf(out, "Backed up existing content to %s\n", backedUpTo)
			}
			fmt.Fprintf(out, "Wrote %d file(s) to %s\n", len(files), deployPath)
			fmt.Fprint(out, sourceHint(cfg, deployPath))
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to shellforge.toml")
	cmd.Flags().StringVar(&shellFlag, "shell", "", "shell to generate for (zsh or bash); overrides the config file")
	cmd.Flags().StringArrayVar(&integrationFlags, "integration", nil, "integration to enable (repeatable); overrides the config file")
	cmd.Flags().StringVar(&deployPathFlag, "deploy-path", "", "where to write the generated script")
	cmd.Flags().BoolVar(&backupFlag, "backup", true, "back up existing content before overwriting")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be written without writing it")
	return cmd
}

type generateOptions struct {
	ConfigPath       string
	ShellFlag        string
	IntegrationFlags []string
	DeployPathFlag   string
	BackupChanged    bool
	BackupFlag       bool
}

// resolveGenerateConfig layers defaults, then the config file, then flags.
// Split out from the cobra command so the precedence rules are unit testable
// without running the CLI.
func resolveGenerateConfig(opts generateOptions) (config.Config, error) {
	path, existed, err := config.Resolve(opts.ConfigPath)
	if err != nil {
		return config.Config{}, err
	}

	cfg := config.Default()
	if existed {
		cfg, err = config.Load(path)
		if err != nil {
			return config.Config{}, err
		}
	}

	if opts.ShellFlag != "" {
		s, err := config.ParseShell(opts.ShellFlag)
		if err != nil {
			return config.Config{}, err
		}
		cfg.Shell = s
	}
	if len(opts.IntegrationFlags) > 0 {
		integrations, err := parseIntegrations(opts.IntegrationFlags)
		if err != nil {
			return config.Config{}, err
		}
		cfg.Integrations = integrations
	}
	if opts.DeployPathFlag != "" {
		cfg.DeployPath = opts.DeployPathFlag
	}
	if opts.BackupChanged {
		cfg.Backup = opts.BackupFlag
	}

	if err := cfg.Validate(); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func parseIntegrations(values []string) ([]config.Integration, error) {
	out := make([]config.Integration, 0, len(values))
	for _, v := range values {
		i, err := config.ParseIntegration(v)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, nil
}
