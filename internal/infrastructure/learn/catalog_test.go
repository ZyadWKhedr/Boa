package learn

import (
	"testing"
)

func TestEmbeddedTechniqueCatalog(t *testing.T) {
	cat, err := NewEmbeddedTechniqueCatalog()
	if err != nil {
		t.Fatalf("Failed to create EmbeddedTechniqueCatalog: %v", err)
	}

	techniques := cat.ListTechniques()
	if len(techniques) < 10 {
		t.Errorf("Expected at least 10 techniques, got %d", len(techniques))
	}

	requiredIDs := []string{
		"store",
		"deflate",
		"zstd",
		"opt-png",
		"flac",
		"store-media",
		"jpeg-quality",
		"png-palette",
		"audio-bitrate",
		"video-quality",
		"video-res",
		"strip-meta",
	}

	for _, id := range requiredIDs {
		info, found := cat.GetTechniqueInfo(id)
		if !found {
			t.Errorf("Required technique %q not found in catalog", id)
			continue
		}

		if info.ID != id {
			t.Errorf("Technique ID mismatch: got %q, want %q", info.ID, id)
		}
		if info.Name == "" {
			t.Errorf("Technique %q has empty Name", id)
		}
		if info.Category == "" {
			t.Errorf("Technique %q has empty Category", id)
		}
		if info.Summary == "" {
			t.Errorf("Technique %q has empty Summary", id)
		}
		if info.Analogy == "" {
			t.Errorf("Technique %q has empty Analogy", id)
		}
		if info.WhatYouGain == "" {
			t.Errorf("Technique %q has empty WhatYouGain", id)
		}
		if info.WhatYouLose == "" {
			t.Errorf("Technique %q has empty WhatYouLose", id)
		}
		if info.BestFor == "" {
			t.Errorf("Technique %q has empty BestFor", id)
		}
		if info.AvoidFor == "" {
			t.Errorf("Technique %q has empty AvoidFor", id)
		}
	}
}
