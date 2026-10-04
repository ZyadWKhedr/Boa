package usecase

import (
	"math"

	"compressor/internal/domain"
)

// EstimateSavingsUseCase calculates estimated size ranges and savings percentages quickly.
type EstimateSavingsUseCase struct{}

// NewEstimateSavingsUseCase creates a new EstimateSavingsUseCase.
func NewEstimateSavingsUseCase() *EstimateSavingsUseCase {
	return &EstimateSavingsUseCase{}
}

// Estimate computes the estimated size bounds (Min, Max, Avg) for a given ScanResult and UserPreferences.
func (uc *EstimateSavingsUseCase) Estimate(scan *domain.ScanResult, prefs domain.UserPreferences) domain.EstimatedSavings {
	if scan == nil || scan.TotalBytes == 0 {
		return domain.EstimatedSavings{}
	}

	var totalMin float64
	var totalMax float64
	var totalAvg float64

	// 1. Images
	if scan.Images.TotalBytes > 0 {
		minR, maxR, avgR := uc.estimateRatio(domain.KindImage, prefs.Images)
		totalMin += float64(scan.Images.TotalBytes) * minR
		totalMax += float64(scan.Images.TotalBytes) * maxR
		totalAvg += float64(scan.Images.TotalBytes) * avgR
	}

	// 2. Audio
	if scan.Audio.TotalBytes > 0 {
		minR, maxR, avgR := uc.estimateRatio(domain.KindAudio, prefs.Audio)
		totalMin += float64(scan.Audio.TotalBytes) * minR
		totalMax += float64(scan.Audio.TotalBytes) * maxR
		totalAvg += float64(scan.Audio.TotalBytes) * avgR
	}

	// 3. Video
	if scan.Video.TotalBytes > 0 {
		minR, maxR, avgR := uc.estimateRatio(domain.KindVideo, prefs.Video)
		totalMin += float64(scan.Video.TotalBytes) * minR
		totalMax += float64(scan.Video.TotalBytes) * maxR
		totalAvg += float64(scan.Video.TotalBytes) * avgR
	}

	// 4. Other (Documents, Code, etc.)
	if scan.Other.TotalBytes > 0 {
		// Lossless Deflate / zstd saves roughly 60-75% on text/code, 0% on already compressed
		ratio := 0.35 // ~65% reduction
		if prefs.ArchiveMethod == domain.MethodStore {
			ratio = 1.0
		}
		totalMin += float64(scan.Other.TotalBytes) * (ratio * 0.85)
		totalMax += float64(scan.Other.TotalBytes) * (ratio * 1.15)
		totalAvg += float64(scan.Other.TotalBytes) * ratio
	}

	estMin := int64(math.Round(totalMin))
	estMax := int64(math.Round(totalMax))
	estAvg := int64(math.Round(totalAvg))

	if estMin < 1 {
		estMin = 1
	}
	if estMax < estMin {
		estMax = estMin
	}
	if estAvg < estMin {
		estAvg = estMin
	}
	if estAvg > estMax {
		estAvg = estMax
	}

	savedPercent := 0.0
	if scan.TotalBytes > 0 {
		savedBytes := scan.TotalBytes - estAvg
		savedPercent = (float64(savedBytes) / float64(scan.TotalBytes)) * 100.0
		if savedPercent < 0 {
			savedPercent = 0
		}
	}

	return domain.EstimatedSavings{
		OriginalBytes:   scan.TotalBytes,
		EstimatedMin:    estMin,
		EstimatedMax:    estMax,
		EstimatedAvg:    estAvg,
		PercentSavedAvg: savedPercent,
	}
}

func (uc *EstimateSavingsUseCase) estimateRatio(kind domain.MediaKind, prefs domain.CategoryPreferences) (minR, maxR, avgR float64) {
	if !prefs.EnableLossy {
		// Lossless mode
		switch kind {
		case domain.KindImage:
			return 0.85, 1.0, 0.95 // Small lossless optimization savings
		case domain.KindAudio:
			return 0.60, 1.0, 0.85 // FLAC for WAV, untouched for MP3
		case domain.KindVideo:
			return 1.0, 1.0, 1.0   // Stored as-is
		default:
			return 0.30, 0.50, 0.40
		}
	}

	// Lossy mode: scaled by quality (0-100)
	q := float64(prefs.Quality)
	if q <= 0 {
		q = 75
	}
	if q > 100 {
		q = 100
	}

	// Quality curve: q=100 -> ratio ~0.8; q=50 -> ratio ~0.35; q=20 -> ratio ~0.15
	baseRatio := 0.10 + (q/100.0)*0.70

	return math.Max(0.05, baseRatio*0.80), math.Min(1.0, baseRatio*1.20), baseRatio
}
