package cli

import (
	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

var (
	flagVerbose bool
	flagQuiet   bool
	flagNoColor bool
)

func (a *App) initCommands() {
	rootCmd := &cobra.Command{
		Use:   "boa [command]",
		Short: "Boa: High-performance, modern compression toolkit",
		Long: `Boa is an ultrafast, modern compression utility engineered in Go.
Supports standard DEFLATE, raw Store, and modern Zstandard (zstd) algorithms
with built-in Zip-Slip security defenses and interactive dashboard tools.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if flagNoColor {
				ui.NoColor = true
			}
			if flagQuiet {
				ui.SuppressBanner = true
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose execution logging")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress banner and non-essential output")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable ANSI color formatting in output")

	rootCmd.AddCommand(a.newPackCmd())
	rootCmd.AddCommand(a.newUnpackCmd())
	rootCmd.AddCommand(a.newListCmd())
	rootCmd.AddCommand(a.newBenchCmd())
	rootCmd.AddCommand(a.newVersionCmd())
	rootCmd.AddCommand(a.newUninstallCmd())
	rootCmd.AddCommand(a.newLearnCmd())

	a.RootCmd = rootCmd
}

// Execute runs the root command.
func (a *App) Execute() error {
	return a.RootCmd.Execute()
}

// ExecuteArgs runs the root command with specific arguments (useful for tests).
func (a *App) ExecuteArgs(args []string) error {
	a.RootCmd.SetArgs(args)
	return a.RootCmd.Execute()
}
