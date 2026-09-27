package bench

import (
	"fmt"
	"math"
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
	if opts.OnProgress != nil {
		opts.OnProgress("Scanning files", 0, 0, filepath.Base(cleanSrc))
	}
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
	for i, lvl := range levels {
		if opts.OnProgress != nil {
			lvlDesc := fmt.Sprintf("Level %d", lvl)
			if lvl == 0 {
				lvlDesc = "Level 0 (Store)"
			}
			opts.OnProgress("Testing compression", i+1, len(levels), lvlDesc)
		}
		res, err := r.benchmarkLevel(cleanSrc, lvl, opts, benchDir, inputMeta)
		if err != nil {
			return nil, fmt.Errorf("benchmark failed for level %d: %w", lvl, err)
		}
		report.Results = append(report.Results, res)
	}

	if opts.OnProgress != nil {
		opts.OnProgress("Finalizing report", len(levels), len(levels), "done")
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
	maxSpaceSavedPct := -1.0
	maxCompressedThroughput := -1.0

	for i, res := range results {
		if res.ThroughputMBps > highestThroughput {
			highestThroughput = res.ThroughputMBps
			fastestIdx = i
		}
		if res.Level > 0 && res.ThroughputMBps > maxCompressedThroughput {
			maxCompressedThroughput = res.ThroughputMBps
		}
		if res.CompressedBytes < smallestSize {
			smallestSize = res.CompressedBytes
			smallestIdx = i
		}
		if res.SpaceSavedPercent > maxSpaceSavedPct {
			maxSpaceSavedPct = res.SpaceSavedPercent
		}
	}
	if maxCompressedThroughput <= 0 {
		maxCompressedThroughput = highestThroughput
	}

	fastest := results[fastestIdx]
	smallest := results[smallestIdx]

	// Dynamically calculate the real Best Balance level based on measured compression ratio vs speed efficiency
	bestBalanceIdx := 0
	var bestScore float64 = -1.0

	for i, res := range results {
		// If compression is achieved (> 0.01%), level 0 (Store) is skipped for best compression balance
		if res.Level == 0 && len(results) > 1 && maxSpaceSavedPct > 0.01 {
			continue
		}

		normComp := 0.0
		if maxSpaceSavedPct > 0 {
			normComp = math.Max(0, res.SpaceSavedPercent/maxSpaceSavedPct)
		} else if res.CompressedBytes > 0 && smallestSize > 0 {
			normComp = float64(smallestSize) / float64(res.CompressedBytes)
		}

		normSpeed := 0.0
		if maxCompressedThroughput > 0 {
			normSpeed = math.Max(0, res.ThroughputMBps/maxCompressedThroughput)
		}

		// Weighted geometric mean giving 65% weight to compression efficiency and 35% to throughput
		score := math.Pow(math.Max(normComp, 0.001), 0.65) * math.Pow(math.Max(normSpeed, 0.001), 0.35)
		if score > bestScore {
			bestScore = score
			bestBalanceIdx = i
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

	// Generate real dynamic insights comparing best balance, fastest, and smallest
	if len(results) >= 2 {
		if bestBalance.Level != fastest.Level && fastest.ThroughputMBps > 0 {
			speedDiff := ((fastest.ThroughputMBps - bestBalance.ThroughputMBps) / bestBalance.ThroughputMBps) * 100.0
			compGain := bestBalance.SpaceSavedPercent - fastest.SpaceSavedPercent
			if compGain > 0 {
				summary.TradeOffInsights = append(summary.TradeOffInsights,
					fmt.Sprintf("Level %d yields +%.1f%% extra compression over Level %d with %s speed.",
						bestBalance.Level, compGain, fastest.Level, bestBalance.ThroughputStr),
				)
			} else {
				summary.TradeOffInsights = append(summary.TradeOffInsights,
					fmt.Sprintf("Level %d runs %.1f%% faster while maintaining %s archive size.",
						fastest.Level, speedDiff, fastest.CompressedHuman),
				)
			}
		}

		if bestBalance.Level != smallest.Level {
			sizeDiff := bestBalance.CompressedBytes - smallest.CompressedBytes
			if sizeDiff > 0 {
				pctExtra := (float64(sizeDiff) / float64(smallest.CompressedBytes)) * 100.0
				summary.TradeOffInsights = append(summary.TradeOffInsights,
					fmt.Sprintf("Level %d saves an additional %s (%.1f%% smaller) over Level %d with longer processing time.",
						smallest.Level, stats.FormatBytes(sizeDiff), pctExtra, bestBalance.Level),
				)
			}
		}
	}

	return summary
}
