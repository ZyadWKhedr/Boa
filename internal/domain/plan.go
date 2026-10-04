package domain

// CategorySummary aggregates counts and sizes for a given media kind.
type CategorySummary struct {
	Count         int    `json:"count"`
	TotalBytes    int64  `json:"total_bytes"`
	Formats       []string `json:"formats"`
	AlreadyLossy  int    `json:"already_lossy_count"`
}

// ScanResult aggregates scanned items and category breakdowns.
type ScanResult struct {
	RootPath        string                   `json:"root_path"`
	Entries         []FileEntry              `json:"entries"`
	TotalFiles      int                      `json:"total_files"`
	TotalDirs       int                      `json:"total_dirs"`
	TotalBytes      int64                    `json:"total_bytes"`
	Images          CategorySummary          `json:"images"`
	Audio           CategorySummary          `json:"audio"`
	Video           CategorySummary          `json:"video"`
	Other           CategorySummary          `json:"other"`
}

// CategoryPreferences holds the user's selected mode and quality for a media kind.
type CategoryPreferences struct {
	EnableLossy bool `json:"enable_lossy"`
	Quality     int  `json:"quality"` // 0-100
}

// UserPreferences captures user choices across all media categories.
type UserPreferences struct {
	ArchiveMethod       CompressionMethod `json:"archive_method"` // deflate, store, zstd
	DefaultLevel        int               `json:"default_level"`  // 0-9 for deflate, 1-11 for zstd
	Images              CategoryPreferences `json:"images"`
	Audio               CategoryPreferences `json:"audio"`
	Video               CategoryPreferences `json:"video"`
	StripMetadata       bool              `json:"strip_metadata"`
}

// EstimatedSavings holds estimated space savings with uncertainty range.
type EstimatedSavings struct {
	OriginalBytes  int64   `json:"original_bytes"`
	EstimatedMin   int64   `json:"estimated_min"`
	EstimatedMax   int64   `json:"estimated_max"`
	EstimatedAvg   int64   `json:"estimated_avg"`
	PercentSavedAvg float64 `json:"percent_saved_avg"`
}

// PlannedFile encapsulates the file entry and its decided action.
type PlannedFile struct {
	Entry          FileEntry `json:"entry"`
	Action         Action    `json:"action"`
	Reason         string    `json:"reason"`
	IsLossy        bool      `json:"is_lossy"`
}

// CompressionPlan represents the validated, ready-to-execute packaging strategy.
type CompressionPlan struct {
	SourcePath      string        `json:"source_path"`
	OutputPath      string        `json:"output_path"`
	Files           []PlannedFile `json:"files"`
	TotalFiles      int           `json:"total_files"`
	TotalBytes      int64         `json:"total_bytes"`
	LossyFileCount  int           `json:"lossy_file_count"`
	HasLossyMedia   bool          `json:"has_lossy_media"`
	EstimatedSize   EstimatedSavings `json:"estimated_size"`
}

// CompressionMethod aliases method for domain plan.
type CompressionMethod string

const (
	MethodDeflate CompressionMethod = "deflate"
	MethodStore   CompressionMethod = "store"
	MethodZstd    CompressionMethod = "zstd"
)
