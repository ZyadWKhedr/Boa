package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrLossyDisallowedForNonMedia is returned when attempting lossy operations on non-media files.
	ErrLossyDisallowedForNonMedia = errors.New("lossy compression is strictly forbidden for non-media files (documents, code, executables, etc.)")

	// ErrLossyMissingConsent is returned when lossy processing is requested without explicit user consent.
	ErrLossyMissingConsent = errors.New("lossy compression requires explicit user consent")
)

// LossyConsent holds explicit user consent flags per media category.
type LossyConsent struct {
	AllowLossyImages bool `json:"allow_lossy_images"`
	AllowLossyAudio  bool `json:"allow_lossy_audio"`
	AllowLossyVideo  bool `json:"allow_lossy_video"`
}

// HasAnyLossyConsent returns true if any media category has explicit lossy consent.
func (c LossyConsent) HasAnyLossyConsent() bool {
	return c.AllowLossyImages || c.AllowLossyAudio || c.AllowLossyVideo
}

// LossyEligibilityPolicy encapsulates and strictly enforces domain rules for lossy operations.
type LossyEligibilityPolicy struct{}

// NewLossyEligibilityPolicy creates a new instance of the domain policy.
func NewLossyEligibilityPolicy() *LossyEligibilityPolicy {
	return &LossyEligibilityPolicy{}
}

// IsLossyEligible determines whether a file entry is allowed to undergo lossy compression.
// Rule 1: ONLY Image, Audio, or Video may ever be eligible.
// Rule 2: Other/Unknown files fail closed and are NEVER eligible.
func (p *LossyEligibilityPolicy) IsLossyEligible(entry FileEntry) bool {
	if entry.IsDir {
		return false
	}
	switch entry.Classification.MediaKind {
	case KindImage, KindAudio, KindVideo:
		return true
	default:
		return false
	}
}

// ValidateAction validates whether a decided action is safe and allowed for a file entry under the given consent.
// This provides defense-in-depth verification during planning and immediately prior to transcoding.
func (p *LossyEligibilityPolicy) ValidateAction(entry FileEntry, action Action, consent LossyConsent) error {
	if action.Kind != ActionTranscode || action.TranscodeParams == nil {
		return nil // Lossless actions (Store, Deflate, Zstd) are always allowed for any file.
	}

	params := action.TranscodeParams
	if params.Lossless {
		return nil // Lossless transcoding (e.g. lossless PNG optimization) is safe.
	}

	// 1. Enforce strict media kind check
	if !p.IsLossyEligible(entry) {
		return fmt.Errorf("%w: file %q classified as %q (%s)",
			ErrLossyDisallowedForNonMedia, entry.Path, entry.Classification.MediaKind, entry.Classification.Description)
	}

	// 2. Enforce explicit user consent per media category
	switch entry.Classification.MediaKind {
	case KindImage:
		if !consent.AllowLossyImages {
			return fmt.Errorf("%w: image lossy compression not granted for %q", ErrLossyMissingConsent, entry.Path)
		}
	case KindAudio:
		if !consent.AllowLossyAudio {
			return fmt.Errorf("%w: audio lossy compression not granted for %q", ErrLossyMissingConsent, entry.Path)
		}
	case KindVideo:
		if !consent.AllowLossyVideo {
			return fmt.Errorf("%w: video lossy compression not granted for %q", ErrLossyMissingConsent, entry.Path)
		}
	default:
		return fmt.Errorf("%w: unrecognized kind %q for %q", ErrLossyDisallowedForNonMedia, entry.Classification.MediaKind, entry.Path)
	}

	return nil
}
