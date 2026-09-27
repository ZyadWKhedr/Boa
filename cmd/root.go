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

// RootCmd is the base command for Compressor CLI.
var RootCmd = &cobra.Command{
	Use:   "compressor",
	Short: "⚡ Fast, secure, and modern zip compression engine for terminal power users",
	Long: `Compressor CLI is a dense, high-performance, cross-platform archive tool.
Designed for macOS, Linux, and Windows with strict Zip-Slip security defenses,
streaming I/O, customizable compression levels, and beautiful terminal metrics.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if flagNoColor {
			ui.NoColor = true
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If invoked without arguments, render dense visual help
		_ = cmd.Help()
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
	RootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose per-file terminal logging")
	RootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-essential progress output")
	RootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable ANSI color output (also respects NO_COLOR env)")

	// Custom tw93/Mole inspired dense help template
	RootCmd.SetHelpTemplate(getMoleHelpTemplate())
}

func getMoleHelpTemplate() string {
	return fmt.Sprintf(`%s
  {{.CommandPath}} [command] [flags]

%s
  {{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
    {{rpad .Name 12}} {{.Short}}{{end}}{{end}}

%s
  {{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

%s
  {{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}

%s
  compressor pack ./my-folder -o project.zip -l 9
  compressor unpack project.zip -o ./extracted --force
  compressor list project.zip
  compressor bench ./sample-data

%s
  Use "{{.CommandPath}} [command] --help" for more information about a command.
`,
		ui.Bold(ui.Cyan("USAGE")),
		ui.Bold(ui.Cyan("COMMANDS")),
		ui.Bold(ui.Cyan("FLAGS")),
		ui.Bold(ui.Cyan("GLOBAL FLAGS")),
		ui.Bold(ui.Cyan("EXAMPLES")),
		ui.Dim("LEARN MORE"),
	)
}