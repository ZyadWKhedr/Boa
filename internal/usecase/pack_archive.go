package usecase

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"compressor/internal/compress"
	"compressor/internal/domain"
	codecImg "compressor/internal/infrastructure/codec/image"
	codecMedia "compressor/internal/infrastructure/codec/media"
	"compressor/internal/safety"
	"compressor/internal/stats"
	"compressor/pkg/types"
	"github.com/klauspost/compress/zstd"
)

// PackArchiveUseCase orchestrates creating zip archives safely.
type PackArchiveUseCase struct {
	engine     *compress.Engine
	policy     *domain.LossyEligibilityPolicy
	imageCodec domain.ImageCodec
	transcoder domain.MediaTranscoder
}

// NewPackArchiveUseCase creates a new PackArchiveUseCase.
func NewPackArchiveUseCase() *PackArchiveUseCase {
	return &PackArchiveUseCase{
		engine:     compress.New(),
		policy:     domain.NewLossyEligibilityPolicy(),
		imageCodec: codecImg.NewStandardImageCodec(),
		transcoder: codecMedia.NewFFmpegTranscoder(),
	}
}

// Execute performs archive packaging with full safety checks from legacy PackOptions.
func (uc *PackArchiveUseCase) Execute(ctx context.Context, opts types.PackOptions) (*types.ArchiveSummary, error) {
	if len(opts.SourcePaths) == 0 {
		return nil, errors.New("no source paths provided for compression")
	}
	if opts.OutputZipPath == "" {
		return nil, errors.New("output zip path must not be empty")
	}

	return uc.engine.Pack(opts)
}

// ExecutePlan performs archive packaging guided by a validated domain.CompressionPlan.
func (uc *PackArchiveUseCase) ExecutePlan(
	ctx context.Context,
	plan *domain.CompressionPlan,
	opts types.PackOptions,
) (*types.ArchiveSummary, error) {
	if plan == nil || len(plan.Files) == 0 {
		return nil, errors.New("cannot execute empty compression plan")
	}

	startTime := time.Now()

	// Ensure destination directory exists
	outDir := filepath.Dir(plan.OutputPath)
	if outDir != "" && outDir != "." {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create destination directory %q: %w", outDir, err)
		}
	}

	tempFile, err := os.CreateTemp(outDir, ".boa_tmp_*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp archive: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		if safety.FileExists(tempPath) {
			_ = os.Remove(tempPath)
		}
	}()

	zipWriter := zip.NewWriter(tempFile)

	// Configure method-specific compressors
	switch opts.Method {
	case types.MethodDeflate:
		if opts.CompressionLevel > 0 {
			zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
				return flate.NewWriter(out, opts.CompressionLevel)
			})
		}
	case types.MethodZstd:
		zipWriter.RegisterCompressor(types.ZipMethodZstd, func(out io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedDefault))
		})
	}

	// Standard ZIP metadata comment for lossy archives (ignored safely by other tools)
	if plan.HasLossyMedia {
		zipWriter.SetComment(fmt.Sprintf("Boa-Lossy: files=%d", plan.LossyFileCount))
	}

	summary := &types.ArchiveSummary{
		ArchivePath:       plan.OutputPath,
		CompressionMethod: opts.Method.DisplayName(),
		CompressionLevel:  opts.CompressionLevel,
	}

	consent := domain.LossyConsent{
		AllowLossyImages: true,
		AllowLossyAudio:  true,
		AllowLossyVideo:  true,
	}

	for _, planned := range plan.Files {
		entry := planned.Entry

		// Defense in depth check 2: verify domain policy directly before any write/transcode
		if err := uc.policy.ValidateAction(entry, planned.Action, consent); err != nil {
			return nil, fmt.Errorf("pre-transcode safety violation on %q: %w", entry.Path, err)
		}

		if err := uc.packPlannedFile(ctx, zipWriter, planned, summary, opts); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize zip archive: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	fi, err := os.Stat(tempPath)
	if err != nil {
		return nil, err
	}
	summary.CompressedBytes = fi.Size()

	if err := os.Rename(tempPath, plan.OutputPath); err != nil {
		// Fallback for cross-device moves
		if err := copyFile(tempPath, plan.OutputPath); err != nil {
			return nil, err
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

func (uc *PackArchiveUseCase) packPlannedFile(
	ctx context.Context,
	zw *zip.Writer,
	planned domain.PlannedFile,
	summary *types.ArchiveSummary,
	opts types.PackOptions,
) error {
	entry := planned.Entry
	relPath := filepath.ToSlash(entry.RelPath)

	header := &zip.FileHeader{
		Name:     relPath,
		Method:   zip.Deflate,
		Modified: entry.ModTime,
	}
	if header.Modified.IsZero() {
		header.Modified = time.Now()
	}

	switch planned.Action.Kind {
	case domain.ActionStore:
		header.Method = zip.Store
	case domain.ActionZstd:
		header.Method = types.ZipMethodZstd
	}

	// Prepare data stream
	var dataReader io.Reader
	fileData, err := os.ReadFile(entry.Path)
	if err != nil {
		return fmt.Errorf("cannot read source file %q: %w", entry.Path, err)
	}
	dataReader = bytes.NewReader(fileData)

	// Apply transcoding if needed
	if planned.Action.Kind == domain.ActionTranscode && planned.Action.TranscodeParams != nil {
		params := planned.Action.TranscodeParams
		if params.Kind == domain.KindImage && uc.imageCodec.CanHandle(entry.Classification.MIMEType) {
			var outBuf bytes.Buffer
			if params.Lossless {
				_ = uc.imageCodec.OptimizeLossless(ctx, bytes.NewReader(fileData), &outBuf, entry.Classification.MIMEType)
			} else {
				_ = uc.imageCodec.CompressLossy(ctx, bytes.NewReader(fileData), &outBuf, entry.Classification.MIMEType, params.Quality)
			}
			if outBuf.Len() > 0 && outBuf.Len() < len(fileData) {
				dataReader = &outBuf
				header.Method = zip.Store // Already compressed lossy image
			}
		}
	}

	w, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry for %q: %w", relPath, err)
	}

	n, err := io.Copy(w, dataReader)
	if err != nil {
		return fmt.Errorf("failed to write data for %q: %w", relPath, err)
	}

	summary.TotalFiles++
	summary.UncompressedBytes += entry.SizeBytes

	if opts.ProgressCallback != nil {
		opts.ProgressCallback(relPath, n, summary.TotalFiles)
	}

	return nil
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

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
