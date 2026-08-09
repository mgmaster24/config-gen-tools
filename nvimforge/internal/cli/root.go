// Package cli wires nvimforge's cobra commands together. It contains
// orchestration only — detection, installation, and generation logic all
// live in their own packages and are never duplicated here.
package cli

import "github.com/spf13/cobra"

// Execute builds and runs the nvimforge root command.
func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "nvimforge",
		Short:         "Install Neovim and generate a language-aware, snacks.nvim-based config",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newInstallCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newVersionCmd())

	return root
}
