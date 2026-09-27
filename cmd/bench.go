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
		}{
			{0, "0 (Store)"},
			{1, "1 (Fastest)"},
			{6, "6 (Default)"},
			{9, "9 (Best)"},
		}

		table := ui.NewTable(
			ui.Column{Title: "Level", Align: ui.AlignLeft},
			ui.Column{Title: "Original Size", Align: ui.AlignRight},
			ui.Column{Title: "Packed Size", Align: ui.AlignRight},
			ui.Column{Title: "Saved", Align: ui.AlignRight},
			ui.Column{Title: "Time", Align: ui.AlignRight},
			ui.Column{Title: "Speed", Align: ui.AlignRight},
		)

		engine := compress.New()

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

			savingsPct := fmt.Sprintf("%.1f%%", summary.SpaceSavedPercent)
			speedStr := stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration)

			table.AddRow(
				ui.Bold(l.Name),
				stats.FormatBytes(summary.UncompressedBytes),
				stats.FormatBytes(summary.CompressedBytes),
				ui.Green(savingsPct),
				stats.FormatDuration(summary.Duration),
				ui.Cyan(speedStr),
			)
		}

		fmt.Println()
		table.Render()
		fmt.Println()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(benchCmd)
}
