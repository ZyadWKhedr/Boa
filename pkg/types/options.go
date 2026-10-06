package types

import (
	"io"
	"time"
)

// PackOptions defines the configuration for compressing files and directories into a zip archive.
type PackOptions struct {
	// SourcePaths are the files or directories to include in the archive.
	SourcePaths []string

	// OutputZipPath is the destination path for the created .zip archive.
	OutputZipPath string

	// Method specifies the compression algorithm: MethodDeflate (default), MethodStore, or MethodZstd.
	Method CompressionMethod

	// CompressionLevel specifies the method-specific compression level.
	// For DEFLATE: 1 (BestSpeed) to 9 (BestCompression), default 6.
	// For Zstandard: 1 (Fastest) to 4 or standard levels (default: 3).
	// For Store: Not applicable (0).
	CompressionLevel int

	// ExcludePatterns are glob patterns for files/directories to skip (e.g. "*.git*", ".DS_Store", "node_modules").
	ExcludePatterns []string

	// Overwrite controls whether an existing archive file should be overwritten.
	Overwrite bool

	// DryRun previews the operation without writing any files.
	DryRun bool

	// Verbose outputs detailed per-file processing information.
	Verbose bool

	// Quiet suppresses all non-essential output.
	Quiet bool

	// BaseDir allows specifying a root directory to calculate relative paths inside the archive.
	BaseDir string

	// TotalFiles is an optional total file count for accurate progress bar estimation.
	TotalFiles int

	// TotalBytes is an optional total byte size for progress percentage estimation.
	TotalBytes int64

	// ProgressCallback reports progress during compression.
	ProgressCallback func(currentFile string, bytesProcessed int64, filesProcessed int)
}

// UnpackOptions defines the configuration for decompressing a zip archive.
type UnpackOptions struct {
	// ArchivePath is the path to the zip archive to extract.
	ArchivePath string

	// DestinationDir is the output directory where extracted files will be written.
	DestinationDir string

	// ExcludePatterns are glob patterns to exclude from extraction.
	ExcludePatterns []string

	// Overwrite controls whether existing files in the destination should be overwritten.
	Overwrite bool

	// DryRun previews the extraction list without writing to disk.
	DryRun bool

	// Verbose outputs detailed per-file extraction details.
	Verbose bool

	// Quiet suppresses all non-essential output.
	Quiet bool

	// SafeMode enforces strict Zip Slip path traversal and symlink validation.
	SafeMode bool

	// TotalFiles is an optional total file count for progress bars.
	TotalFiles int

	// TotalBytes is an optional total byte size for progress bars.
	TotalBytes int64

	// ProgressCallback reports progress during extraction.
	ProgressCallback func(currentFile string, bytesExtracted int64, filesExtracted int)
}

// FileEntry represents a file or directory entry inside a zip archive or filesystem.
type FileEntry struct {
	Name             string      `json:"name"`
	Path             string      `json:"path"`
	OriginalSize     int64       `json:"original_size"`
	CompressedSize   int64       `json:"compressed_size"`
	CompressionRatio float64     `json:"compression_ratio"`
	IsDir            bool        `json:"is_dir"`
	Mode             uint32      `json:"mode"`
	ModTime          time.Time   `json:"mod_time"`
	CRC32            uint32      `json:"crc32"`
	Method           uint16      `json:"method"`
	Comment          string      `json:"comment,omitempty"`
}

// ArchiveSummary holds aggregated metrics for an archive or compression/decompression job.
type ArchiveSummary struct {
	TotalFiles        int           `json:"total_files"`
	TotalDirs         int           `json:"total_dirs"`
	UncompressedBytes int64         `json:"uncompressed_bytes"`
	CompressedBytes   int64         `json:"compressed_bytes"`
	AverageFileSize   int64         `json:"average_file_size"`
	CompressionRatio  float64       `json:"compression_ratio"`
	SpaceSavedBytes   int64         `json:"space_saved_bytes"`
	SpaceSavedPercent float64       `json:"space_saved_percent"`
	CompressionMethod string        `json:"compression_method,omitempty"`
	CompressionLevel  int           `json:"compression_level,omitempty"`
	Duration          time.Duration `json:"duration"`
	ArchivePath       string        `json:"archive_path,omitempty"`
	Comment           string        `json:"comment,omitempty"`
	Entries           []FileEntry   `json:"entries,omitempty"`
}

// ProgressWriter wraps an io.Writer to count bytes written.
type ProgressWriter struct {
	Writer     io.Writer
	TotalBytes int64
	OnWrite    func(n int)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.TotalBytes += int64(n)
	if pw.OnWrite != nil {
		pw.OnWrite(n)
	}
	return n, err
}
