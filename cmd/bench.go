package cmd

import (
	"fmt"
	"os"

	"compressor/internal/bench"
	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

var (
	benchLevel   int
	benchRuns    int
	benchJSON    bool
	benchCSV     bool
	benchCompare bool
	benchKeep    bool
	benchDecomp  bool
)

var benchCmd = &cobra.Command{
	Use:     "bench <source-folder-or-file> [flags]",
	Aliases: []string{"benchmark", "test-speed"},
	Short:   "Benchmark compression levels (0-9) and compare speed, ratio, and output sizes",
	Long: `Executes side-by-side compression benchmarks across all Deflate levels (0 Store, 1 Fastest,
up to 9 Maximum) or a specified level. Measures compression duration, space saved %, compression ratio,
and throughput (MB/s).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]

		if benchLevel != -1 && (benchLevel < 0 || benchLevel > 9) {
			return fmt.Errorf("invalid compression level %d: must be between 0 and 9", benchLevel)
		}

		if benchRuns <= 0 {
			benchRuns = 1
		}

		opts := bench.Options{
			SourcePath:      sourcePath,
			Level:           benchLevel,
			Runs:            benchRuns,
			BenchmarkDecomp: benchDecomp,
			KeepArchives:    benchKeep,
			CompareVisual:   benchCompare,
		}

		runner := bench.NewRunner()
		report, err := runner.Run(opts)
		if err != nil {
			return err
		}

		// Handle machine-readable JSON output
		if benchJSON {
			return bench.RenderJSON(report, os.Stdout)
		}

		// Handle CSV output
		if benchCSV {
			return bench.RenderCSV(report, os.Stdout)
		}

		// Default formatted terminal output
		if !flagQuiet {
			ui.PrintBanner()
		}

		bench.RenderText(report, benchCompare, os.Stdout)
		return nil
	},
}

func init() {
	benchCmd.Flags().IntVarP(&benchLevel, "level", "l", -1, "Target specific compression level to benchmark (0-9, default: all levels)")
	benchCmd.Flags().IntVarP(&benchRuns, "runs", "r", 1, "Number of benchmark iterations per level to compute averages")
	benchCmd.Flags().BoolVar(&benchJSON, "json", false, "Output benchmark metrics in structured JSON format")
	benchCmd.Flags().BoolVar(&benchCSV, "csv", false, "Output benchmark metrics in standard CSV format")
	benchCmd.Flags().BoolVar(&benchCompare, "compare", false, "Display visual ASCII comparison bar charts for size and throughput")
	benchCmd.Flags().BoolVar(&benchKeep, "keep", false, "Preserve created benchmark archives on disk (default: auto-cleaned)")
	benchCmd.Flags().BoolVar(&benchDecomp, "decomp", false, "Also benchmark decompression time and throughput")

	RootCmd.AddCommand(benchCmd)
}
