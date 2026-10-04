package cli

import (
	"compressor/internal/usecase"
	"github.com/spf13/cobra"
)

// App encapsulates dependencies and commands for the CLI presentation layer.
type App struct {
	PackUC      *usecase.PackArchiveUseCase
	ExtractUC   *usecase.ExtractArchiveUseCase
	InspectUC   *usecase.InspectArchiveUseCase
	BenchmarkUC *usecase.BenchmarkUseCase
	LearnUC     *usecase.LearnUseCase
	Version     string
	RootCmd     *cobra.Command
}

// NewApp constructs a new CLI App with injected use cases.
func NewApp(
	packUC *usecase.PackArchiveUseCase,
	extractUC *usecase.ExtractArchiveUseCase,
	inspectUC *usecase.InspectArchiveUseCase,
	benchUC *usecase.BenchmarkUseCase,
	learnUC *usecase.LearnUseCase,
	version string,
) *App {
	app := &App{
		PackUC:      packUC,
		ExtractUC:   extractUC,
		InspectUC:   inspectUC,
		BenchmarkUC: benchUC,
		LearnUC:     learnUC,
		Version:     version,
	}

	app.initCommands()
	return app
}
