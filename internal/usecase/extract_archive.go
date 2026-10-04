package usecase

import (
	"context"

	"compressor/internal/extract"
	"compressor/pkg/types"
)

// ExtractArchiveUseCase orchestrates extracting zip archives safely.
type ExtractArchiveUseCase struct {
	engine *extract.Engine
}

// NewExtractArchiveUseCase creates a new ExtractArchiveUseCase.
func NewExtractArchiveUseCase() *ExtractArchiveUseCase {
	return &ExtractArchiveUseCase{
		engine: extract.New(),
	}
}

// Execute extracts an archive to a destination directory.
func (uc *ExtractArchiveUseCase) Execute(ctx context.Context, opts types.UnpackOptions) (*types.ArchiveSummary, error) {
	return uc.engine.Unpack(opts)
}
