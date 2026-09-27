package compress

import (
	"archive/zip"
	"compress/flate"
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

// Engine performs compression operations.
type Engine struct{}

// New creates a new compression engine.
func New() *Engine {
	return &Engine{}
}

// Pack compresses source directories or files into a zip archive based on the provided options.
func (e *Engine) Pack(opts types.PackOptions) (*types.ArchiveSummary, error) {
	if len(opts.SourcePaths) == 0 {
		return nil, errors.New("no source paths provided for compression")
	}
	if opts.OutputZipPath == "" {
		return nil, errors.New("output zip path must not be empty")
	}

	// Check if destination exists when overwrite is false
	if !opts.Overwrite && !opts.DryRun && safety.FileExists(opts.OutputZipPath) {
		return nil, fmt.Errorf("output file %q already exists (use --force to overwrite)", opts.OutputZipPath)
	}

	startTime := time.Now()

	methodName := "Deflate"
	if opts.CompressionLevel == 0 {
		methodName = "Store"
	}

	summary := &types.ArchiveSummary{
		ArchivePath:       opts.OutputZipPath,
		CompressionMethod: methodName,
		CompressionLevel:  opts.CompressionLevel,
	}

	if opts.DryRun {
		// Dry run: collect items and compute uncompressed size without writing archive
		for _, src := range opts.SourcePaths {
			if err := e.simulatePack(src, opts, summary); err != nil {
				return nil, err
			}
		}
		summary.Duration = time.Since(startTime)
		return summary, nil
	}

	// Ensure destination directory exists
	outDir := filepath.Dir(opts.OutputZipPath)
	if outDir != "" && outDir != "." {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create destination directory %q: %w", outDir, err)
		}
	}

	// Use temporary file for atomic write safety
	tempFile, err := os.CreateTemp(outDir, ".compressor_tmp_*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary archive file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		if safety.FileExists(tempPath) {
			_ = os.Remove(tempPath)
		}
	}()

	zipWriter := zip.NewWriter(tempFile)

	// Configure compression level if specified
	if opts.CompressionLevel >= 0 && opts.CompressionLevel <= 9 {
		if opts.CompressionLevel == 0 {
			// Store (no compression)
			zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
				return flate.NewWriter(out, flate.NoCompression)
			})
		} else {
			zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
				return flate.NewWriter(out, opts.CompressionLevel)
			})
		}
	}

	// Pack each source path
	for _, src := range opts.SourcePaths {
		if err := e.packPath(zipWriter, src, opts, summary); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize zip archive: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary archive file: %w", err)
	}

	// Get final compressed file size
	fi, err := os.Stat(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat compressed archive: %w", err)
	}
	summary.CompressedBytes = fi.Size()

	// Atomic rename to final output
	if err := os.Rename(tempPath, opts.OutputZipPath); err != nil {
		// Fallback for cross-device renames: copy then remove
		if copyErr := copyFile(tempPath, opts.OutputZipPath); copyErr != nil {
			return nil, fmt.Errorf("failed to move archive to final path %q: %w", opts.OutputZipPath, copyErr)
		}
	}

	summary.Duration = time.Since(startTime)
	if summary.TotalFiles > 0 {
		summary.AverageFileSize = summary.UncompressedBytes / int64(summary.TotalFiles)
	}
	summary.CompressionRatio = stats.CalculateRatio(summary.UncompressedBytes, summary.CompressedBytes)
	summary.SpaceSavedBytes, summary.SpaceSavedPercent = stats.CalculateSpaceSaved(summary.UncompressedBytes, summary.CompressedBytes)

	return summary, nil
}

func (e *Engine) packPath(zw *zip.Writer, srcPath string, opts types.PackOptions, summary *types.ArchiveSummary) error {
	fi, err := os.Lstat(srcPath)
	if err != nil {
		return fmt.Errorf("cannot read path %q: %w", srcPath, err)
	}

	baseDir := opts.BaseDir
	if baseDir == "" {
		if fi.IsDir() {
			baseDir = filepath.Dir(srcPath)
		} else {
			baseDir = filepath.Dir(srcPath)
		}
	}

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Calculate relative path for zip entry
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		// Check exclusions
		if safety.MatchExclude(relPath, opts.ExcludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Zip spec requires forward slashes
		zipEntryName := filepath.ToSlash(relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %q: %w", path, err)
		}

		header.Name = zipEntryName
		header.Modified = info.ModTime()

		if info.IsDir() {
			if !strings.HasSuffix(header.Name, "/") {
				header.Name += "/"
			}
			header.Method = zip.Store
			if _, err := zw.CreateHeader(header); err != nil {
				return fmt.Errorf("failed to write directory header for %q: %w", path, err)
			}
			summary.TotalDirs++
			return nil
		}

		// Handle files
		if opts.CompressionLevel == 0 {
			header.Method = zip.Store
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create entry header for %q: %w", path, err)
		}

		// Stream file content
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file %q: %w", path, err)
		}
		defer file.Close()

		written, err := io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("failed to stream content for %q: %w", path, err)
		}

		summary.TotalFiles++
		summary.UncompressedBytes += written

		if opts.ProgressCallback != nil {
			opts.ProgressCallback(path, summary.UncompressedBytes, summary.TotalFiles)
		}

		return nil
	})
}

func (e *Engine) simulatePack(srcPath string, opts types.PackOptions, summary *types.ArchiveSummary) error {
	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir = filepath.Dir(srcPath)
	}

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		if safety.MatchExclude(relPath, opts.ExcludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			summary.TotalDirs++
		} else {
			summary.TotalFiles++
			summary.UncompressedBytes += info.Size()
		}
		return nil
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
