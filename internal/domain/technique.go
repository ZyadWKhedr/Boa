package domain

// TechniqueCategory represents whether a technique is lossless or lossy.
type TechniqueCategory string

const (
	CategoryLossless TechniqueCategory = "lossless"
	CategoryLossy    TechniqueCategory = "lossy"
)

// TechniqueInfo represents educational knowledge explaining a compression technique in friendly terms.
type TechniqueInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    TechniqueCategory `json:"category"`
	MediaKind   string            `json:"media_kind"` // general | image | audio | video
	Summary     string            `json:"summary"`    // 1 sentence
	Analogy     string            `json:"analogy"`    // 1-2 sentences
	WhatYouGain string            `json:"what_you_gain"`
	WhatYouLose string            `json:"what_you_lose"`
	BestFor     string            `json:"best_for"`
	AvoidFor    string            `json:"avoid_for"`
	GoDeeper    string            `json:"go_deeper,omitempty"` // Optional deeper technical paragraph
}
