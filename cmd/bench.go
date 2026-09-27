package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"compressor/internal/compress"
	"compressor/internal/stats"
	"compressor/internal/ui"
	"compressor/pkg/types"
	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:     "bench <source-folder-or-file>",
	Aliases: []string{"benchmark", "test-speed"},
	Short:   "Benchmark compression levels (0, 1, 6, 9) and compare speed vs ratio",
	Long:    `Runs an in-memory or temporary benchmark on the specified source path comparing compression speeds, ratios, and output sizes.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]

		ui.PrintBanner()
		ui.PrintSection(fmt.Sprintf("Benchmarking Compression Levels on: %s", ui.Bold(sourcePath)))

		tempDir, err := os.MkdirTemp("", "compressor_bench_*")
		if err != nil {
			return fmt.Errorf("failed to create benchmark temp dir: %w", err)
		}
		defer os.RemoveAll(tempDir)

		levels := []struct {
			Level int
			Name  string
			Tag   string
		}{
			{0, "0 Store", "Uncompressed"},
			{1, "1 Fastest", "Best Speed"},
			{6, "6 Balanced", "Recommended ★"},
			{9, "9 Maximum", "Max Density"},
		}

		table := ui.NewTable(
			ui.Column{Title: "Compression Level", Align: ui.AlignLeft},
			ui.Column{Title: "Method", Align: ui.AlignCenter},
			ui.Column{Title: "Output Size", Align: ui.AlignRight},
			ui.Column{Title: "Reduction", Align: ui.AlignRight},
			ui.Column{Title: "Ratio", Align: ui.AlignRight},
			ui.Column{Title: "Time", Align: ui.AlignRight},
			ui.Column{Title: "Throughput", Align: ui.AlignRight},
		)

		engine := compress.New()
		var summaries []*types.ArchiveSummary

		for _, l := range levels {
			benchOut := filepath.Join(tempDir, fmt.Sprintf("bench_lvl_%d.zip", l.Level))
			opts := types.PackOptions{
				SourcePaths:      []string{sourcePath},
				OutputZipPath:    benchOut,
				CompressionLevel: l.Level,
				Overwrite:        true,
				Quiet:            true,
			}

			summary, err := engine.Pack(opts)
			if err != nil {
				return fmt.Errorf("benchmark level %d failed: %w", l.Level, err)
			}
			summaries = append(summaries, summary)

			method := "DEFLATE"
			if l.Level == 0 {
				method = "STORE"
			}

			levelLabel := l.Name
			if l.Level == 6 {
				levelLabel = ui.Bold(ui.Cyan(l.Name + " ★"))
			}

			savingsPct := fmt.Sprintf("%.1f%%", summary.SpaceSavedPercent)
			ratioStr := fmt.Sprintf("%.2fx", 1.0/max(summary.CompressionRatio, 0.01))
			speedStr := stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration)

			table.AddRow(
				levelLabel,
				ui.Dim(method),
				stats.FormatBytes(summary.CompressedBytes),
				ui.Green(savingsPct),
				ui.Cyan(ratioStr),
				stats.FormatDuration(summary.Duration),
				speedStr,
			)
		}

		fmt.Println()
		table.Render()
		fmt.Println()

		if len(summaries) >= 4 {
			s0 := summaries[0]
			s6 := summaries[2]
			s9 := summaries[3]
			extraSavings := s6.CompressedBytes - s9.CompressedBytes
			if extraSavings < 0 {
				extraSavings = 0
			}
			fmt.Printf(" %s %s\n", ui.Bold(ui.Cyan("✦ Recommendation:")), 
				fmt.Sprintf("Level 6 gives %s reduction at %s with optimal CPU efficiency.",
					ui.Bold(ui.Green(fmt.Sprintf("%.1f%%", s6.SpaceSavedPercent))),
					stats.CalculateSpeed(s0.UncompressedBytes, s6.Duration)))
			if extraSavings > 0 {
				fmt.Printf("   %s\n", ui.Dim(fmt.Sprintf("Level 9 saves %s more space (%s) but took %s vs %s.",
					stats.FormatBytes(extraSavings),
					fmt.Sprintf("%.1f%% extra", s9.SpaceSavedPercent-s6.SpaceSavedPercent),
					stats.FormatDuration(s9.Duration),
					stats.FormatDuration(s6.Duration))))
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(benchCmd)
}
