package cmd

import (
	"os"

	"compressor/internal/infrastructure/learn"
	"compressor/internal/presentation/cli"
	"compressor/internal/presentation/tui"
	"compressor/internal/ui"
	"compressor/internal/usecase"
	"github.com/spf13/cobra"
)

var (
	// Version holds the semantic version of Compressor CLI.
	Version = "v0.3.1"
	// Commit holds the git commit hash.
	Commit = "dev"
	// Date holds the build timestamp.
	Date = "2026-10-06"

	// RootCmd is the base command for Boa CLI.
	RootCmd *cobra.Command

	// Global application references
	cliApp    *cli.App
	dashboard *tui.Dashboard
)

func init() {
	packUC := usecase.NewPackArchiveUseCase()
	extractUC := usecase.NewExtractArchiveUseCase()
	inspectUC := usecase.NewInspectArchiveUseCase()
	benchUC := usecase.NewBenchmarkUseCase()

	catalog, err := learn.NewEmbeddedTechniqueCatalog()
	if err != nil {
		panic("failed to initialize technique catalog: " + err.Error())
	}
	learnUC := usecase.NewLearnUseCase(catalog)

	cliApp = cli.NewApp(packUC, extractUC, inspectUC, benchUC, learnUC, Version)
	dashboard = tui.NewDashboard(packUC, extractUC, inspectUC, benchUC, learnUC, Version)

	RootCmd = cliApp.RootCmd
	RootCmd.Use = "bo"
	RootCmd.Run = func(cmd *cobra.Command, args []string) {
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			dashboard.Run()
			return
		}
		_ = cmd.Help()
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
		os.Exit(1)
	}
}

// RunInteractiveDashboard launches the interactive terminal dashboard.
func RunInteractiveDashboard() {
	dashboard.Run()
}

// PickPath bridges to tui.PickPath.
func PickPath(title string, mode tui.FilePickerMode) (string, error) {
	return tui.PickPath(title, mode)
}

// AnimateBoaSnake bridges to cli.AnimateBoaSnake.
var AnimateBoaSnake = cli.AnimateBoaSnake

// RunUninstall bridges to cli.RunUninstall.
var RunUninstall = cli.RunUninstall