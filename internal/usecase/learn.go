package usecase

import (
	"fmt"
	"sort"

	"compressor/internal/domain"
)

// LearnUseCase handles educational compression technique inquiries.
type LearnUseCase struct {
	catalog domain.TechniqueCatalog
}

// NewLearnUseCase creates a new LearnUseCase.
func NewLearnUseCase(catalog domain.TechniqueCatalog) *LearnUseCase {
	return &LearnUseCase{catalog: catalog}
}

// ListTechniques returns all available techniques sorted by category and name.
func (uc *LearnUseCase) ListTechniques() []domain.TechniqueInfo {
	list := uc.catalog.ListTechniques()
	sort.Slice(list, func(i, j int) bool {
		if list[i].Category != list[j].Category {
			return list[i].Category < list[j].Category
		}
		return list[i].Name < list[j].Name
	})
	return list
}

// GetTechniqueInfo retrieves details for a specific technique ID.
func (uc *LearnUseCase) GetTechniqueInfo(id string) (domain.TechniqueInfo, error) {
	info, found := uc.catalog.GetTechniqueInfo(id)
	if !found {
		return domain.TechniqueInfo{}, fmt.Errorf("technique %q not found in catalog", id)
	}
	return info, nil
}
