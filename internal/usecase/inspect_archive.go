package usecase

import (
	"context"

	"compressor/internal/extract"
	"compressor/pkg/types"
)

// InspectArchiveUseCase inspects archive contents and metadata safely.
type InspectArchiveUseCase struct {
	engine *extract.Engine
}

// NewInspectArchiveUseCase creates a new InspectArchiveUseCase.
func NewInspectArchiveUseCase() *InspectArchiveUseCase {
	return &InspectArchiveUseCase{
		engine: extract.New(),
	}
}

// Execute returns deep inspection details for a given archive file path.
func (uc *InspectArchiveUseCase) Execute(ctx context.Context, archivePath string) (*types.ArchiveSummary, error) {
	return uc.engine.InspectArchive(archivePath)
}
