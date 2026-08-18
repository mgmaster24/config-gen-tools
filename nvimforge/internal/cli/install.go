package cli

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/mgmaster24/config-gen-tools/forge/prereq"
	"github.com/mgmaster24/config-gen-tools/forge/runner"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/checks"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/genconfig"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/github"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/neovim"
	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/ui"
)

func newInstallCmd() *cobra.Command {
	var (
		configPath        string
		langFlags         []string
		deployPathFlag    string
		backupFlag        bool
		noBanner          bool
		yes               bool
		skipNeovimInstall bool
		dryRun            bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Verify prerequisites, install/update Neovim, and generate a Neovim config",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			opts := installConfigOptions{
				ConfigPath:     configPath,
				LangFlags:      langFlags,
				DeployPathFlag: deployPathFlag,
				BackupFlag:     backupFlag,
				BackupChanged:  cmd.Flags().Changed("backup"),
				NoBanner:       noBanner,
				Yes:            yes,
				Prompt:         ui.PromptConfig,
			}

			cfg, saveNeeded, savePath, err := resolveInstallConfig(opts)
			if err != nil {
				return err
			}

			if saveNeeded {
				confirmed, err := ui.ConfirmSave(savePath)
				if err != nil {
					return err
				}
				if confirmed {
					if err := config.Save(savePath, cfg); err != nil {
						return err
					}
					fmt.Fprintf(out, "Saved configuration to %s\n\n", savePath)
				}
			}

			ui.PrintBanner(out, cfg.ShowBanner)

			report := prereq.Run(runner.OSRunner{}, checks.ForLanguages(cfg.Languages), runtime.GOOS)
			prereq.RenderText(out, report)
			fmt.Fprintln(out)
			if report.HasMissingRequired() {
				return &ExitError{Code: 1}
			}

			// A missing BlocksTooling prereq doesn't stop generation — prereqs
			// stay report-only — but it does mean the config we're about to
			// write can't finish installing itself, so confirm first. Under
			// --yes the warning RenderText already printed has to stand on its
			// own.
			if report.HasMissingBlocking() && !yes {
				names := make([]string, 0, len(report.MissingBlocking()))
				for _, res := range report.MissingBlocking() {
					names = append(names, res.Check.Name)
				}
				proceed, err := ui.ConfirmContinueWithBlocking(names)
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(out, "Aborted; nothing was written.")
					return nil
				}
			}

			if !skipNeovimInstall {
				if err := ensureNeovim(cmd.Context(), out, dryRun); err != nil {
					return err
				}
			}

			return generateConfig(out, cfg, dryRun)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to nvimforge.toml (default: ./nvimforge.toml, then ~/.config/nvimforge/config.toml)")
	cmd.Flags().StringArrayVar(&langFlags, "lang", nil, "language to install (repeatable); overrides the config file")
	cmd.Flags().StringVar(&deployPathFlag, "deploy-path", "", "where to write the generated Neovim config; overrides the config file")
	cmd.Flags().BoolVar(&backupFlag, "backup", true, "back up an existing config before overwriting")
	cmd.Flags().BoolVar(&noBanner, "no-banner", false, "don't show the nvimforge startup banner")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "run non-interactively; requires --lang if no config file exists")
	cmd.Flags().BoolVar(&skipNeovimInstall, "skip-neovim-install", false, "only (re)generate the Neovim config; don't touch the Neovim binary")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would happen without downloading, installing, or writing anything")

	return cmd
}

// newNeovimInstaller constructs the Installer used by `install`. It's a
// package-level seam (rather than a hardcoded call in ensureNeovim) so
// tests can substitute a fake ReleaseClient and temp directories without
// touching the network or the real filesystem.
var newNeovimInstaller = func() (*neovim.Installer, error) {
	return neovim.NewInstaller(github.NewHTTPReleaseClient())
}

func ensureNeovim(ctx context.Context, out io.Writer, dryRun bool) error {
	installer, err := newNeovimInstaller()
	if err != nil {
		return fmt.Errorf("setting up neovim installer: %w", err)
	}
	result, err := installer.EnsureInstalled(ctx, neovim.InstallOptions{DryRun: dryRun})
	if err != nil {
		return fmt.Errorf("installing neovim: %w", err)
	}
	prefix := ""
	if dryRun {
		prefix = "[dry run] would be "
	}
	fmt.Fprintf(out, "Neovim %s: %s%s\n\n", result.Version, prefix, result.Action)
	return nil
}

func generateConfig(out io.Writer, cfg config.Config, dryRun bool) error {
	deployPath, err := cfg.ExpandedDeployPath()
	if err != nil {
		return err
	}

	data := genconfig.BuildTemplateData(cfg)
	files, err := genconfig.Render(data)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(out, "[dry run] would write %d files to %s\n", len(files), deployPath)
		return nil
	}

	backedUpTo, err := genconfig.Write(files, deployPath, cfg.Backup)
	if err != nil {
		return err
	}
	if backedUpTo != "" {
		fmt.Fprintf(out, "Backed up existing config to %s\n", backedUpTo)
	}
	fmt.Fprintf(out, "Wrote %d files to %s\n\n", len(files), deployPath)
	fmt.Fprintln(out, "Run `nvim` and wait for lazy.nvim and mason to finish installing plugins on first launch.")
	return nil
}

// installConfigOptions carries everything resolveInstallConfig needs to
// apply the precedence rule flags > TOML file > interactive prompt >
// defaults. Prompt is injected (rather than calling ui.PromptConfig
// directly) so tests can exercise every branch without driving a real
// interactive terminal form.
type installConfigOptions struct {
	ConfigPath     string
	LangFlags      []string
	DeployPathFlag string
	BackupFlag     bool
	BackupChanged  bool
	NoBanner       bool
	Yes            bool
	Prompt         func(config.Config) (config.Config, error)
}

// resolveInstallConfig returns the final Config to install with, and, if it
// came from a fresh interactive prompt (saveNeeded), the path it should be
// offered for saving to.
func resolveInstallConfig(opts installConfigOptions) (cfg config.Config, saveNeeded bool, savePath string, err error) {
	resolvedPath, existed, err := config.Resolve(opts.ConfigPath)
	if err != nil {
		return config.Config{}, false, "", err
	}

	switch {
	case existed:
		cfg, err = config.Load(resolvedPath)
		if err != nil {
			return config.Config{}, false, "", err
		}
	case len(opts.LangFlags) > 0 || opts.Yes:
		// With no config file and no --lang, --yes falls back to
		// config.DefaultLanguages rather than erroring: the defaults are the
		// answer a user would most likely have given at the prompt.
		cfg = config.Default()
	default:
		cfg, err = opts.Prompt(config.Default())
		if err != nil {
			return config.Config{}, false, "", err
		}
		saveNeeded = true
		savePath = resolvedPath
	}

	if len(opts.LangFlags) > 0 {
		langs, err := parseLanguages(opts.LangFlags)
		if err != nil {
			return config.Config{}, false, "", err
		}
		cfg.Languages = langs
	}
	if opts.DeployPathFlag != "" {
		cfg.DeployPath = opts.DeployPathFlag
	}
	if opts.BackupChanged {
		cfg.Backup = opts.BackupFlag
	}
	if opts.NoBanner {
		cfg.ShowBanner = false
	}

	if err := cfg.Validate(); err != nil {
		return config.Config{}, false, "", err
	}

	return cfg, saveNeeded, savePath, nil
}
