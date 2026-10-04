package usecase

import (
	"context"

	"compressor/internal/bench"
)

// BenchmarkUseCase orchestrates compression benchmarking.
type BenchmarkUseCase struct {
	runner *bench.Runner
}

// NewBenchmarkUseCase creates a new BenchmarkUseCase.
func NewBenchmarkUseCase() *BenchmarkUseCase {
	return &BenchmarkUseCase{
		runner: bench.NewRunner(),
	}
}

// Execute runs a benchmark with the provided options.
func (uc *BenchmarkUseCase) Execute(ctx context.Context, opts bench.Options) (*bench.Report, error) {
	return uc.runner.Run(opts)
}
