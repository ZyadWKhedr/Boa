package bench

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"compressor/internal/compress"
	"compressor/internal/extract"
	"compressor/internal/safety"
	"compressor/internal/stats"
	"compressor/pkg/types"
)

// Runner executes compression benchmarks.
type Runner struct {
	compressEngine *compress.Engine
	extractEngine  *extract.Engine
}

// NewRunner creates a new benchmark runner.
func NewRunner() *Runner {
	return &Runner{
		compressEngine: compress.New(),
		extractEngine:  extract.New(),
	}
}

// Run executes the benchmark according to the provided options.
func (r *Runner) Run(opts Options) (*Report, error) {
	if opts.SourcePath == "" {
		return nil, fmt.Errorf("source path cannot be empty")
	}

	cleanSrc := filepath.Clean(opts.SourcePath)
	if !safety.FileExists(cleanSrc) {
		return nil, fmt.Errorf("source path %q does not exist", cleanSrc)
	}

	// 1. Inspect input source metadata
	inputMeta, err := r.inspectInput(cleanSrc)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect input source: %w", err)
	}

	// 2. Determine levels to benchmark
	levels := r.determineLevels(opts.Level)

	if opts.Runs <= 0 {
		opts.Runs = 1
	}

	// Create temporary workspace for benchmark archives
	benchDir, err := os.MkdirTemp("", "boa_bench_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create benchmark temp dir: %w", err)
	}

	if !opts.KeepArchives {
		defer os.RemoveAll(benchDir)
	}

	report := &Report{
		Title:          "Boa Compression Benchmark",
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
		Input:          inputMeta,
		RunsPerLevel:   opts.Runs,
		IsSmallDataset: inputMeta.SizeBytes < 10*1024, // < 10 KB
		Results:        make([]LevelResult, 0, len(levels)),
	}

	if opts.KeepArchives {
		report.PreservedDir = benchDir
	}

	// 3. Execute benchmark for each level
	for _, lvl := range levels {
		res, err := r.benchmarkLevel(cleanSrc, lvl, opts, benchDir, inputMeta)
		if err != nil {
			return nil, fmt.Errorf("benchmark failed for level %d: %w", lvl, err)
		}
		report.Results = append(report.Results, res)
	}

	// 4. Generate summary highlights and trade-off insights
	report.Summary = r.generateSummary(report.Results, inputMeta)

	return report, nil
}

func (r *Runner) inspectInput(path string) (InputMetadata, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return InputMetadata{}, err
	}

	meta := InputMetadata{
		Path: path,
	}

	if !fi.IsDir() {
		meta.TotalFiles = 1
		meta.SizeBytes = fi.Size()
		meta.SizeHuman = stats.FormatBytes(meta.SizeBytes)
		return meta, nil
	}

	err = filepath.Walk(path, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			meta.TotalDirs++
		} else {
			meta.TotalFiles++
			meta.SizeBytes += info.Size()
		}
		return nil
	})

	meta.SizeHuman = stats.FormatBytes(meta.SizeBytes)
	return meta, err
}

func (r *Runner) determineLevels(chosenLevel int) []int {
	if chosenLevel >= 0 && chosenLevel <= 9 {
		return []int{chosenLevel}
	}
	// Benchmark all meaningful levels: 0 through 9
	return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
}

func (r *Runner) benchmarkLevel(srcPath string, level int, opts Options, tempDir string, input InputMetadata) (LevelResult, error) {
	var totalCompDuration time.Duration
	var minCompDuration time.Duration = time.Duration(1<<63 - 1)
	var maxCompDuration time.Duration

	var totalDecompDuration time.Duration
	var minDecompDuration time.Duration = time.Duration(1<<63 - 1)
	var maxDecompDuration time.Duration

	var lastSummary *types.ArchiveSummary
	outArchive := filepath.Join(tempDir, fmt.Sprintf("bench_level_%d.zip", level))

	for run := 0; run < opts.Runs; run++ {
		packOpts := types.PackOptions{
			SourcePaths:      []string{srcPath},
			OutputZipPath:    outArchive,
			CompressionLevel: level,
			Overwrite:        true,
			Quiet:            true,
		}

		startComp := time.Now()
		summary, err := r.compressEngine.Pack(packOpts)
		compDuration := time.Since(startComp)
		if err != nil {
			return LevelResult{}, err
		}

		lastSummary = summary
		totalCompDuration += compDuration
		if compDuration < minCompDuration {
			minCompDuration = compDuration
		}
		if compDuration > maxCompDuration {
			maxCompDuration = compDuration
		}

		// Benchmark decompression if requested
		if opts.BenchmarkDecomp {
			decompDir := filepath.Join(tempDir, fmt.Sprintf("decomp_lvl_%d_run_%d", level, run))
			unpackOpts := types.UnpackOptions{
				ArchivePath:    outArchive,
				DestinationDir: decompDir,
				Overwrite:      true,
				Quiet:          true,
			}

			startDecomp := time.Now()
			_, err := r.extractEngine.Unpack(unpackOpts)
			decompDuration := time.Since(startDecomp)
			_ = os.RemoveAll(decompDir)

			if err != nil {
				return LevelResult{}, fmt.Errorf("decompression benchmark failed: %w", err)
			}

			totalDecompDuration += decompDuration
			if decompDuration < minDecompDuration {
				minDecompDuration = decompDuration
			}
			if decompDuration > maxDecompDuration {
				maxDecompDuration = decompDuration
			}
		}

		// If not keeping and not the last run, remove archive to avoid reuse artifacts
		if !opts.KeepArchives && run < opts.Runs-1 {
			_ = os.Remove(outArchive)
		}
	}

	avgCompDuration := totalCompDuration / time.Duration(opts.Runs)
	if avgCompDuration <= 0 {
		avgCompDuration = time.Microsecond
	}

	compressedBytes := lastSummary.CompressedBytes
	savedBytes, savedPct := stats.CalculateSpaceSaved(input.SizeBytes, compressedBytes)
	ratio := stats.CalculateRatio(input.SizeBytes, compressedBytes)

	var ratioMultiplier float64
	if compressedBytes > 0 {
		ratioMultiplier = float64(input.SizeBytes) / float64(compressedBytes)
	} else {
		ratioMultiplier = 1.0
	}

	method := "DEFLATE"
	if level == 0 {
		method = "STORE"
	}

	throughputMBps := CalculateThroughputMBps(input.SizeBytes, avgCompDuration)
	throughputHuman := stats.CalculateSpeed(input.SizeBytes, avgCompDuration)

	result := LevelResult{
		Level:             level,
		Name:              GetLevelName(level),
		Method:            method,
		DurationMs:        float64(avgCompDuration.Microseconds()) / 1000.0,
		DurationAvg:       avgCompDuration,
		DurationMin:       minCompDuration,
		DurationMax:       maxCompDuration,
		CompressedBytes:   compressedBytes,
		CompressedHuman:   stats.FormatBytes(compressedBytes),
		CompressionRatio:  ratio,
		RatioMultiplier:   ratioMultiplier,
		SpaceSavedBytes:   savedBytes,
		SpaceSavedPercent: savedPct,
		ThroughputMBps:    throughputMBps,
		ThroughputStr:     throughputHuman,
	}

	if opts.KeepArchives {
		result.ArchivePath = outArchive
	}

	if opts.BenchmarkDecomp && opts.Runs > 0 {
		avgDecompDuration := totalDecompDuration / time.Duration(opts.Runs)
		if avgDecompDuration <= 0 {
			avgDecompDuration = time.Microsecond
		}
		result.DecompDurationMs = float64(avgDecompDuration.Microseconds()) / 1000.0
		result.DecompDurationAvg = avgDecompDuration
		result.DecompThroughput = CalculateThroughputMBps(input.SizeBytes, avgDecompDuration)
		result.DecompThroughputS = stats.CalculateSpeed(input.SizeBytes, avgDecompDuration)
	}

	return result, nil
}

// CalculateThroughputMBps returns throughput in Megabytes per second (MB/s).
func CalculateThroughputMBps(bytes int64, d time.Duration) float64 {
	if d <= 0 || bytes <= 0 {
		return 0.0
	}
	seconds := d.Seconds()
	mb := float64(bytes) / (1024.0 * 1024.0)
	return mb / seconds
}

// GetLevelName returns a human-friendly name tag for a Deflate level.
func GetLevelName(level int) string {
	switch level {
	case 0:
		return "0 (Store)"
	case 1:
		return "1 (Fastest)"
	case 6:
		return "6 (Default)"
	case 9:
		return "9 (Maximum)"
	default:
		return fmt.Sprintf("%d", level)
	}
}

func (r *Runner) generateSummary(results []LevelResult, input InputMetadata) SummaryHighlights {
	if len(results) == 0 {
		return SummaryHighlights{}
	}

	fastestIdx := 0
	smallestIdx := 0
	highestThroughput := -1.0
	smallestSize := int64(1<<63 - 1)

	for i, res := range results {
		// Level 0 (store) has virtually no compression CPU overhead, so for fastest compressed level we prefer level 1 or highest throughput
		if res.ThroughputMBps > highestThroughput {
			highestThroughput = res.ThroughputMBps
			fastestIdx = i
		}
		if res.CompressedBytes < smallestSize {
			smallestSize = res.CompressedBytes
			smallestIdx = i
		}
	}

	fastest := results[fastestIdx]
	smallest := results[smallestIdx]

	// Best balance is typically level 6, or the highest ratio with < 2x level 1 time
	bestBalanceIdx := 0
	for i, res := range results {
		if res.Level == 6 {
			bestBalanceIdx = i
			break
		}
	}
	bestBalance := results[bestBalanceIdx]

	summary := SummaryHighlights{
		FastestLevel:      fastest.Level,
		FastestSpeed:      fastest.ThroughputStr,
		SmallestLevel:     smallest.Level,
		SmallestSize:      smallest.CompressedHuman,
		BestBalanceLevel:  bestBalance.Level,
		BestBalanceDetail: fmt.Sprintf("%s / %s", bestBalance.ThroughputStr, bestBalance.CompressedHuman),
		TradeOffInsights:  make([]string, 0),
	}

	// Generate trade-off comparisons if all levels were run
	if len(results) >= 7 {
		var l1, l6, l9 *LevelResult
		for i := range results {
			switch results[i].Level {
			case 1:
				l1 = &results[i]
			case 6:
				l6 = &results[i]
			case 9:
				l9 = &results[i]
			}
		}

		if l1 != nil && l6 != nil {
			compGain := l6.SpaceSavedPercent - l1.SpaceSavedPercent
			var speedDiff float64
			if l1.ThroughputMBps > 0 {
				speedDiff = ((l6.ThroughputMBps - l1.ThroughputMBps) / l1.ThroughputMBps) * 100.0
			}

			summary.TradeOffInsights = append(summary.TradeOffInsights,
				fmt.Sprintf("Level 6 compared with Level 1: %+.1f%% compression, %+.1f%% throughput", compGain, speedDiff),
			)
		}

		if l6 != nil && l9 != nil {
			diffPercent := l9.SpaceSavedPercent - l6.SpaceSavedPercent
			if diffPercent < 1.0 {
				summary.TradeOffInsights = append(summary.TradeOffInsights,
					fmt.Sprintf("Levels 6–9 produced less than 1.0%% difference in archive size (%.1f%% extra space saved at Level 9).", diffPercent),
				)
			} else {
				summary.TradeOffInsights = append(summary.TradeOffInsights,
					fmt.Sprintf("Level 9 saved an additional %.1f%% space over Level 6 with longer duration.", diffPercent),
				)
			}
		}
	}

	return summary
}
