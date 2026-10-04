package usecase

import (
	"context"
	"errors"

	"compressor/internal/compress"
	"compressor/internal/domain"
	"compressor/pkg/types"
)

// PackArchiveUseCase orchestrates creating zip archives safely.
type PackArchiveUseCase struct {
	engine *compress.Engine
	policy *domain.LossyEligibilityPolicy
}

// NewPackArchiveUseCase creates a new PackArchiveUseCase.
func NewPackArchiveUseCase() *PackArchiveUseCase {
	return &PackArchiveUseCase{
		engine: compress.New(),
		policy: domain.NewLossyEligibilityPolicy(),
	}
}

// Execute performs archive packaging with full safety checks.
func (uc *PackArchiveUseCase) Execute(ctx context.Context, opts types.PackOptions) (*types.ArchiveSummary, error) {
	if len(opts.SourcePaths) == 0 {
		return nil, errors.New("no source paths provided for compression")
	}
	if opts.OutputZipPath == "" {
		return nil, errors.New("output zip path must not be empty")
	}

	// Defense in depth: policy check
	// (Step 1 keeps lossless pack; further steps route media transcoding through here)
	return uc.engine.Pack(opts)
}
