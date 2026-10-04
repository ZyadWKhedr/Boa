package usecase

import (
	"context"
	"fmt"
	"strings"

	"compressor/internal/domain"
)

// BuildCompressionPlanUseCase constructs and verifies a complete compression plan.
type BuildCompressionPlanUseCase struct {
	policy    *domain.LossyEligibilityPolicy
	estimator *EstimateSavingsUseCase
}

// NewBuildCompressionPlanUseCase creates a new BuildCompressionPlanUseCase.
func NewBuildCompressionPlanUseCase(
	policy *domain.LossyEligibilityPolicy,
	estimator *EstimateSavingsUseCase,
) *BuildCompressionPlanUseCase {
	return &BuildCompressionPlanUseCase{
		policy:    policy,
		estimator: estimator,
	}
}

// Execute creates a verified CompressionPlan matching user preferences against safety policies.
func (uc *BuildCompressionPlanUseCase) Execute(
	ctx context.Context,
	scan *domain.ScanResult,
	prefs domain.UserPreferences,
	outputPath string,
) (*domain.CompressionPlan, error) {
	if scan == nil {
		return nil, fmt.Errorf("scan result cannot be nil")
	}

	consent := domain.LossyConsent{
		AllowLossyImages: prefs.Images.EnableLossy,
		AllowLossyAudio:  prefs.Audio.EnableLossy,
		AllowLossyVideo:  prefs.Video.EnableLossy,
	}

	plan := &domain.CompressionPlan{
		SourcePath: scan.RootPath,
		OutputPath: outputPath,
		TotalFiles: scan.TotalFiles,
		TotalBytes: scan.TotalBytes,
		Files:      make([]domain.PlannedFile, 0, len(scan.Entries)),
	}

	for _, entry := range scan.Entries {
		action, reason, isLossy := uc.determineAction(entry, prefs)

		// Defense in Depth 1: Enforce domain policy during planning
		if err := uc.policy.ValidateAction(entry, action, consent); err != nil {
			return nil, fmt.Errorf("policy rejection on file %q: %w", entry.Path, err)
		}

		if isLossy {
			plan.LossyFileCount++
			plan.HasLossyMedia = true
		}

		plan.Files = append(plan.Files, domain.PlannedFile{
			Entry:   entry,
			Action:  action,
			Reason:  reason,
			IsLossy: isLossy,
		})
	}

	if uc.estimator != nil {
		plan.EstimatedSize = uc.estimator.Estimate(scan, prefs)
	}

	return plan, nil
}

func (uc *BuildCompressionPlanUseCase) determineAction(
	entry domain.FileEntry,
	prefs domain.UserPreferences,
) (domain.Action, string, bool) {
	cls := entry.Classification
	kind := cls.MediaKind

	switch kind {
	case domain.KindImage:
		if prefs.Images.EnableLossy {
			q := prefs.Images.Quality
			if q <= 0 {
				q = 75
			}
			return domain.Action{
				Kind: domain.ActionTranscode,
				TranscodeParams: &domain.TranscodeParams{
					Kind:     domain.KindImage,
					Quality:  q,
					Lossless: false,
				},
			}, fmt.Sprintf("Lossy image optimization (%d%% quality)", q), true
		}

		// Lossless mode for images:
		if strings.ToLower(cls.Extension) == ".png" {
			return domain.Action{
				Kind: domain.ActionTranscode,
				TranscodeParams: &domain.TranscodeParams{
					Kind:     domain.KindImage,
					Lossless: true,
				},
			}, "Lossless PNG recompression", false
		}
		// Lossless JPEG: store as-is
		return domain.Action{Kind: domain.ActionStore}, "Store as-is (Lossless)", false

	case domain.KindAudio:
		if prefs.Audio.EnableLossy {
			q := prefs.Audio.Quality
			if q <= 0 {
				q = 64 // Balanced
			}
			return domain.Action{
				Kind: domain.ActionTranscode,
				TranscodeParams: &domain.TranscodeParams{
					Kind:     domain.KindAudio,
					Quality:  q,
					Lossless: false,
				},
			}, fmt.Sprintf("Lossy audio encoding (%d%% bitrate quality)", q), true
		}

		// Lossless audio
		if !cls.IsAlreadyCompressed && strings.ToLower(cls.Extension) == ".wav" {
			return domain.Action{
				Kind: domain.ActionTranscode,
				TranscodeParams: &domain.TranscodeParams{
					Kind:     domain.KindAudio,
					Lossless: true,
				},
			}, "Lossless audio (FLAC)", false
		}
		return domain.Action{Kind: domain.ActionStore}, "Store as-is (Lossless)", false

	case domain.KindVideo:
		if prefs.Video.EnableLossy {
			q := prefs.Video.Quality
			if q <= 0 {
				q = 70
			}
			return domain.Action{
				Kind: domain.ActionTranscode,
				TranscodeParams: &domain.TranscodeParams{
					Kind:     domain.KindVideo,
					Quality:  q,
					Lossless: false,
				},
			}, fmt.Sprintf("Lossy video encoding (CRF %d%% quality)", q), true
		}
		// Lossless video: always store as-is
		return domain.Action{Kind: domain.ActionStore}, "Store as-is (Lossless)", false

	default: // KindOther (documents, code, archives, databases, executables)
		if cls.IsAlreadyCompressed {
			return domain.Action{Kind: domain.ActionStore}, "Store as-is (already compressed)", false
		}
		if prefs.ArchiveMethod == domain.MethodZstd {
			lvl := prefs.DefaultLevel
			if lvl <= 0 {
				lvl = 3
			}
			return domain.Action{
				Kind:      domain.ActionZstd,
				ZstdLevel: lvl,
			}, fmt.Sprintf("Zstandard (Level %d)", lvl), false
		}

		lvl := prefs.DefaultLevel
		if lvl <= 0 {
			lvl = 6
		}
		return domain.Action{
			Kind:         domain.ActionDeflate,
			DeflateLevel: lvl,
		}, fmt.Sprintf("DEFLATE (Level %d)", lvl), false
	}
}
