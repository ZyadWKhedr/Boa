package bench

import (
	"time"
)

// Options holds configuration for running a compression benchmark.
type Options struct {
	SourcePath      string `json:"source_path"`
	Level           int    `json:"level"`            // -1 means benchmark all levels (0-9)
	Runs            int    `json:"runs"`             // Number of iterations per level (default: 1)
	BenchmarkDecomp bool   `json:"benchmark_decomp"` // Measure decompression throughput
	KeepArchives    bool   `json:"keep_archives"`    // Preserve created benchmark archives
	OutputDir       string `json:"output_dir"`       // Directory to store archives if kept
	CompareVisual   bool   `json:"compare_visual"`   // Show ASCII bar chart visualizer
}

// InputMetadata contains metadata about the benchmark target.
type InputMetadata struct {
	Path       string `json:"path"`
	TotalFiles int    `json:"total_files"`
	TotalDirs  int    `json:"total_dirs"`
	SizeBytes  int64  `json:"size_bytes"`
	SizeHuman  string `json:"size_human"`
}

// LevelResult stores the measured performance and compression metrics for a single Deflate level.
type LevelResult struct {
	Level             int           `json:"level"`
	Name              string        `json:"name"`
	Method            string        `json:"method"`
	DurationMs        float64       `json:"duration_ms"`
	DurationAvg       time.Duration `json:"duration_avg"`
	DurationMin       time.Duration `json:"duration_min"`
	DurationMax       time.Duration `json:"duration_max"`
	CompressedBytes   int64         `json:"compressed_size"`
	CompressedHuman   string        `json:"compressed_size_human"`
	CompressionRatio  float64       `json:"compression_ratio"`    // compressed / original
	RatioMultiplier   float64       `json:"ratio_multiplier"`     // original / compressed (e.g. 2.79)
	SpaceSavedBytes   int64         `json:"space_saved_bytes"`
	SpaceSavedPercent float64       `json:"space_saved_percent"`
	ThroughputMBps    float64       `json:"throughput_mbps"`
	ThroughputStr     string        `json:"throughput_human"`

	// Decompression metrics (optional)
	DecompDurationMs  float64       `json:"decomp_duration_ms,omitempty"`
	DecompDurationAvg time.Duration `json:"decomp_duration_avg,omitempty"`
	DecompThroughput  float64       `json:"decomp_throughput_mbps,omitempty"`
	DecompThroughputS string        `json:"decomp_throughput_human,omitempty"`

	ArchivePath string `json:"archive_path,omitempty"`
}

// SummaryHighlights summarizes the trade-offs across tested levels.
type SummaryHighlights struct {
	FastestLevel      int      `json:"fastest_level"`
	FastestSpeed      string   `json:"fastest_speed"`
	SmallestLevel     int      `json:"smallest_level"`
	SmallestSize      string   `json:"smallest_size"`
	BestBalanceLevel  int      `json:"best_balance_level"`
	BestBalanceDetail string   `json:"best_balance_detail"`
	TradeOffInsights  []string `json:"tradeoff_insights"`
}

// Report encapsulates the complete benchmark execution results.
type Report struct {
	Title          string            `json:"title"`
	Timestamp      string            `json:"timestamp"`
	Input          InputMetadata     `json:"input"`
	RunsPerLevel   int               `json:"runs_per_level"`
	IsSmallDataset bool              `json:"is_small_dataset"`
	Results        []LevelResult     `json:"results"`
	Summary        SummaryHighlights `json:"summary"`
	PreservedDir   string            `json:"preserved_dir,omitempty"`
}
