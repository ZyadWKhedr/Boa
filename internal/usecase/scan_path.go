package usecase

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"compressor/internal/domain"
	"compressor/internal/safety"
)

// ScanPathUseCase scans target paths and classifies discovered files.
type ScanPathUseCase struct {
	classifier domain.FileClassifier
}

// NewScanPathUseCase creates a new ScanPathUseCase.
func NewScanPathUseCase(classifier domain.FileClassifier) *ScanPathUseCase {
	return &ScanPathUseCase{classifier: classifier}
}

// Execute walks the target path, discovers entries, classifies each item, and builds category summaries.
func (uc *ScanPathUseCase) Execute(ctx context.Context, targetPath string, excludes []string) (*domain.ScanResult, error) {
	cleanPath := filepath.Clean(targetPath)
	fi, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access target path %q: %w", targetPath, err)
	}

	result := &domain.ScanResult{
		RootPath: cleanPath,
		Entries:  make([]domain.FileEntry, 0),
	}

	formatSetImg := make(map[string]bool)
	formatSetAud := make(map[string]bool)
	formatSetVid := make(map[string]bool)
	formatSetOth := make(map[string]bool)

	if !fi.IsDir() {
		// Single file target
		cls, err := uc.classifier.Classify(cleanPath)
		if err != nil {
			return nil, err
		}

		entry := domain.FileEntry{
			Path:           cleanPath,
			RelPath:        filepath.Base(cleanPath),
			IsDir:          false,
			SizeBytes:      fi.Size(),
			ModTime:        fi.ModTime(),
			Mode:           uint32(fi.Mode()),
			Classification: cls,
		}

		result.Entries = append(result.Entries, entry)
		result.TotalFiles = 1
		result.TotalBytes = fi.Size()

		uc.aggregateSummary(entry, result, formatSetImg, formatSetAud, formatSetVid, formatSetOth)
		return result, nil
	}

	baseDir := filepath.Dir(cleanPath)

	err = filepath.WalkDir(cleanPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			relPath = path
		}

		if safety.MatchExclude(relPath, excludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if d.IsDir() {
			result.TotalDirs++
			return nil
		}

		cls, err := uc.classifier.Classify(path)
		if err != nil {
			cls = domain.FileClassification{
				MediaKind:   domain.KindOther,
				Description: "Unclassified file",
			}
		}

		entry := domain.FileEntry{
			Path:           path,
			RelPath:        relPath,
			IsDir:          false,
			SizeBytes:      info.Size(),
			ModTime:        info.ModTime(),
			Mode:           uint32(info.Mode()),
			Classification: cls,
		}

		result.Entries = append(result.Entries, entry)
		result.TotalFiles++
		result.TotalBytes += info.Size()

		uc.aggregateSummary(entry, result, formatSetImg, formatSetAud, formatSetVid, formatSetOth)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scanning failed: %w", err)
	}

	result.Images.Formats = keysToSortedSlice(formatSetImg)
	result.Audio.Formats = keysToSortedSlice(formatSetAud)
	result.Video.Formats = keysToSortedSlice(formatSetVid)
	result.Other.Formats = keysToSortedSlice(formatSetOth)

	return result, nil
}

func (uc *ScanPathUseCase) aggregateSummary(
	entry domain.FileEntry,
	res *domain.ScanResult,
	setImg, setAud, setVid, setOth map[string]bool,
) {
	cls := entry.Classification
	ext := strings.TrimPrefix(strings.ToLower(cls.Extension), ".")
	if ext == "" {
		ext = "unknown"
	}

	switch cls.MediaKind {
	case domain.KindImage:
		res.Images.Count++
		res.Images.TotalBytes += entry.SizeBytes
		setImg[ext] = true
		if cls.IsAlreadyLossy {
			res.Images.AlreadyLossy++
		}
	case domain.KindAudio:
		res.Audio.Count++
		res.Audio.TotalBytes += entry.SizeBytes
		setAud[ext] = true
		if cls.IsAlreadyLossy {
			res.Audio.AlreadyLossy++
		}
	case domain.KindVideo:
		res.Video.Count++
		res.Video.TotalBytes += entry.SizeBytes
		setVid[ext] = true
		if cls.IsAlreadyLossy {
			res.Video.AlreadyLossy++
		}
	default:
		res.Other.Count++
		res.Other.TotalBytes += entry.SizeBytes
		setOth[ext] = true
	}
}

func keysToSortedSlice(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
