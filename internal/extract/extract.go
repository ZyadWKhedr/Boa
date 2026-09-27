package extract

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"compressor/internal/safety"
	"compressor/internal/stats"
	"compressor/pkg/types"
)

// Engine performs decompression and archive inspection operations.
type Engine struct{}

// New creates a new extraction engine.
func New() *Engine {
	return &Engine{}
}

// Unpack extracts a zip archive to the target destination directory according to options.
func (e *Engine) Unpack(opts types.UnpackOptions) (*types.ArchiveSummary, error) {
	if opts.ArchivePath == "" {
		return nil, errors.New("archive path must not be empty")
	}
	if opts.DestinationDir == "" {
		return nil, errors.New("destination directory must not be empty")
	}

	fi, statErr := os.Stat(opts.ArchivePath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("archive %q does not exist", opts.ArchivePath)
		}
		return nil, fmt.Errorf("cannot access %q: %w", opts.ArchivePath, statErr)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("'%s' is an unpacked folder, not a compressed zip archive to extract", filepath.Base(opts.ArchivePath))
	}

	reader, err := zip.OpenReader(opts.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("'%s' is not a valid zip archive or is already unpacked", filepath.Base(opts.ArchivePath))
	}
	defer reader.Close()

	startTime := time.Now()
	summary := &types.ArchiveSummary{
		ArchivePath: opts.ArchivePath,
	}

	if !opts.DryRun {
		if err := os.MkdirAll(opts.DestinationDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create target root directory %q: %w", opts.DestinationDir, err)
		}
	}

	for _, file := range reader.File {
		// Clean and normalize name
		cleanName := filepath.Clean(file.Name)

		// Exclude check
		if safety.MatchExclude(cleanName, opts.ExcludePatterns) {
			continue
		}

		// Security: Validate path against Zip-Slip traversal attacks
		targetFilePath, err := safety.ValidateDestinationPath(opts.DestinationDir, file.Name)
		if err != nil {
			return nil, fmt.Errorf("security validation failed on entry %q: %w", file.Name, err)
		}

		if file.FileInfo().IsDir() {
			summary.TotalDirs++
			if !opts.DryRun {
				if err := os.MkdirAll(targetFilePath, 0o755); err != nil {
					return nil, fmt.Errorf("failed to create directory %q: %w", targetFilePath, err)
				}
			}
			continue
		}

		summary.TotalFiles++
		summary.UncompressedBytes += int64(file.UncompressedSize64)
		summary.CompressedBytes += int64(file.CompressedSize64)

		if opts.DryRun {
			continue
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(targetFilePath)
		if err := os.MkdirAll(parentDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create parent directory for %q: %w", targetFilePath, err)
		}

		// Check if file exists when overwrite is false
		if !opts.Overwrite && safety.FileExists(targetFilePath) {
			return nil, fmt.Errorf("target file %q already exists (use --force to overwrite)", targetFilePath)
		}

		// Extract file content safely
		if err := e.extractFile(file, targetFilePath); err != nil {
			return nil, fmt.Errorf("failed to extract file %q: %w", file.Name, err)
		}

		if opts.ProgressCallback != nil {
			opts.ProgressCallback(file.Name, summary.UncompressedBytes, summary.TotalFiles)
		}
	}

	summary.Duration = time.Since(startTime)
	summary.CompressionRatio = stats.CalculateRatio(summary.UncompressedBytes, summary.CompressedBytes)
	summary.SpaceSavedBytes, summary.SpaceSavedPercent = stats.CalculateSpaceSaved(summary.UncompressedBytes, summary.CompressedBytes)

	return summary, nil
}

func (e *Engine) extractFile(zf *zip.File, destPath string) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Handle symlinks
	if zf.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := io.ReadAll(rc)
		if err != nil {
			return err
		}
		targetStr := strings.TrimSpace(string(linkTarget))
		// Validate symlink destination
		destDir := filepath.Dir(destPath)
		if err := safety.ValidateSymlinkTarget(destDir, destPath, targetStr); err != nil {
			return err
		}
		_ = os.Remove(destPath)
		return os.Symlink(targetStr, destPath)
	}

	// Create output file with safe permissions
	mode := zf.Mode().Perm()
	if mode == 0 {
		mode = 0o644
	}

	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return err
	}

	// Restore modification timestamp
	_ = os.Chtimes(destPath, time.Now(), zf.Modified)

	return nil
}

// InspectArchive reads an archive's header entries and calculates statistics without extracting.
func (e *Engine) InspectArchive(archivePath string) (*types.ArchiveSummary, error) {
	fi, statErr := os.Stat(archivePath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("archive %q does not exist", archivePath)
		}
		return nil, fmt.Errorf("cannot access %q: %w", archivePath, statErr)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("'%s' is an unpacked folder and does not contain packed files to inspect (use 'bo pack' to compress it into a .zip archive)", filepath.Base(archivePath))
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("'%s' is not a valid zip archive or is already unpacked", filepath.Base(archivePath))
	}
	defer reader.Close()

	summary := &types.ArchiveSummary{
		ArchivePath: archivePath,
		Entries:     make([]types.FileEntry, 0, len(reader.File)),
	}

	for _, f := range reader.File {
		entry := types.FileEntry{
			Name:           filepath.Base(f.Name),
			Path:           f.Name,
			OriginalSize:   int64(f.UncompressedSize64),
			CompressedSize: int64(f.CompressedSize64),
			IsDir:          f.FileInfo().IsDir(),
			Mode:           uint32(f.Mode()),
			ModTime:        f.Modified,
			CRC32:          f.CRC32,
			Method:         f.Method,
			Comment:        f.Comment,
		}

		if !entry.IsDir && entry.OriginalSize > 0 {
			entry.CompressionRatio = float64(entry.CompressedSize) / float64(entry.OriginalSize)
		}

		if entry.IsDir {
			summary.TotalDirs++
		} else {
			summary.TotalFiles++
			summary.UncompressedBytes += entry.OriginalSize
			summary.CompressedBytes += entry.CompressedSize
		}

		summary.Entries = append(summary.Entries, entry)
	}

	if summary.TotalFiles > 0 {
		summary.AverageFileSize = summary.UncompressedBytes / int64(summary.TotalFiles)
	}
	summary.CompressionRatio = stats.CalculateRatio(summary.UncompressedBytes, summary.CompressedBytes)
	summary.SpaceSavedBytes, summary.SpaceSavedPercent = stats.CalculateSpaceSaved(summary.UncompressedBytes, summary.CompressedBytes)

	return summary, nil
}
