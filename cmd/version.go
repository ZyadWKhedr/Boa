package cmd

import (
	"fmt"
	"runtime"

	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version holds the semantic version of Compressor CLI (injected during build or default).
	Version = "v0.1.0"
	// Commit holds the git commit hash.
	Commit = "dev"
	// Date holds the build timestamp.
	Date = "2026-09-27"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v", "status", "info"},
	Short:   "Print the version and runtime environment information",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintBanner()
		fmt.Printf("  %-16s : %s\n", ui.Dim("Version"), ui.Bold(ui.Cyan(Version)))
		fmt.Printf("  %-16s : %s\n", ui.Dim("Commit"), Commit)
		fmt.Printf("  %-16s : %s\n", ui.Dim("Build Date"), Date)
		fmt.Printf("  %-16s : %s\n", ui.Dim("Go Runtime"), runtime.Version())
		fmt.Printf("  %-16s : %s/%s\n", ui.Dim("Platform"), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
