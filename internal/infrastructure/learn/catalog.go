package learn

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"compressor/internal/domain"
)

//go:embed techniques/*.json
var techniquesFS embed.FS

// EmbeddedTechniqueCatalog loads technique knowledge from embedded files.
type EmbeddedTechniqueCatalog struct {
	mu         sync.RWMutex
	techniques map[string]domain.TechniqueInfo
	list       []domain.TechniqueInfo
}

// NewEmbeddedTechniqueCatalog initializes and loads the embedded catalog.
func NewEmbeddedTechniqueCatalog() (*EmbeddedTechniqueCatalog, error) {
	entries, err := techniquesFS.ReadDir("techniques")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded techniques dir: %w", err)
	}

	techniques := make(map[string]domain.TechniqueInfo)
	var list []domain.TechniqueInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := techniquesFS.ReadFile("techniques/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded file %s: %w", entry.Name(), err)
		}

		var info domain.TechniqueInfo
		if err := json.Unmarshal(data, &info); err != nil {
			return nil, fmt.Errorf("failed to parse technique file %s: %w", entry.Name(), err)
		}

		if info.ID == "" {
			return nil, fmt.Errorf("technique file %s has empty ID", entry.Name())
		}

		techniques[info.ID] = info
		list = append(list, info)
	}

	return &EmbeddedTechniqueCatalog{
		techniques: techniques,
		list:       list,
	}, nil
}

// ListTechniques returns all loaded techniques.
func (c *EmbeddedTechniqueCatalog) ListTechniques() []domain.TechniqueInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]domain.TechniqueInfo, len(c.list))
	copy(result, c.list)
	return result
}

// GetTechniqueInfo returns a technique by its ID.
func (c *EmbeddedTechniqueCatalog) GetTechniqueInfo(id string) (domain.TechniqueInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info, found := c.techniques[id]
	return info, found
}
