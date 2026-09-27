package cmd

import (
	"fmt"
	"os"

	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	flagVerbose bool
	flagQuiet   bool
	flagNoColor bool
)

// RootCmd is the base command for Boa CLI.
var RootCmd = &cobra.Command{
	Use:            "bo",
	Short:          "🐍 Tight, fast, lossless zip compression engine for terminal power users",
	SilenceUsage:   true,
	SilenceErrors:  true,
	Long: `Boa CLI is a dense, high-performance, cross-platform archive tool.
Designed for macOS, Linux, and Windows with strict Zip-Slip security defenses,
streaming I/O, customizable compression levels, and beautiful terminal metrics.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if flagNoColor {
			ui.NoColor = true
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
		os.Exit(1)
	}
}

func init() {
	RootCmd.Run = func(cmd *cobra.Command, args []string) {
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			RunInteractiveDashboard()
			return
		}
		_ = cmd.Help()
	}

	RootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show detailed operation logs")
	RootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-essential output")
	RootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable ANSI color formatting")

	// Custom tw93/Mole inspired dense help template
	RootCmd.SetHelpTemplate(getMoleHelpTemplate())
}

func getMoleHelpTemplate() string {
	return fmt.Sprintf(`%s
  bo                           Main menu
  bo pack                      Compress folders into zip archives
  bo unpack                    Safely extract zip archives
  bo list                      Inspect contents & compression ratios
  bo bench                     Benchmark compression levels (0-9)
  bo version                   Show version & platform info
  bo --help                    Show help

  bo pack ./folder -o dist.zip -l 9
  bo pack ./folder --dry-run
  bo unpack dist.zip -o ./out --force
  bo list dist.zip --json
  bo bench ./large-data

%s
  -v, --verbose                Show detailed operation logs
  -q, --quiet                  Suppress non-essential output
      --no-color               Disable ANSI color formatting
`,
		ui.Bold(ui.Cyan("COMMANDS")),
		ui.Bold(ui.Cyan("OPTIONS")),
	)
}