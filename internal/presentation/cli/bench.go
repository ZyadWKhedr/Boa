package cli

import (
	"context"
	"fmt"
	"os"

	"compressor/internal/bench"
	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

func (a *App) newBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bench <target-path> [flags]",
		Aliases: []string{"benchmark", "b", "test"},
		Short:   "Measure compression speed, duration & throughput across levels",
		Long: `Benchmarks zip compression performance on a target file or folder across all compression levels (0-9).
Supports JSON/CSV export, multi-run averaging, side-by-side level comparison, and visual bar charts.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer func() {
				_ = cmd.Flags().Set("level", "-1")
				_ = cmd.Flags().Set("runs", "1")
				_ = cmd.Flags().Set("compare", "false")
				_ = cmd.Flags().Set("format", "table")
				_ = cmd.Flags().Set("json", "false")
				_ = cmd.Flags().Set("csv", "false")
				_ = cmd.Flags().Set("keep-archives", "false")
			}()

			src := args[0]

			level, _ := cmd.Flags().GetInt("level")
			runs, _ := cmd.Flags().GetInt("runs")
			compare, _ := cmd.Flags().GetBool("compare")
			outputFormat, _ := cmd.Flags().GetString("format")
			jsonOut, _ := cmd.Flags().GetBool("json")
			csvOut, _ := cmd.Flags().GetBool("csv")
			keepArchives, _ := cmd.Flags().GetBool("keep-archives")

			if jsonOut {
				outputFormat = "json"
			} else if csvOut {
				outputFormat = "csv"
			}

			if outputFormat == "table" && !flagQuiet {
				ui.PrintBanner()
				fmt.Println()
				ui.PrintInfo(fmt.Sprintf("Benchmarking target: %s", ui.Bold(ui.PrettyPath(src))))
				if compare {
					fmt.Println(ui.Dim(" Mode: Comparative trade-off evaluation"))
				}
				fmt.Println()
			}

			opts := bench.Options{
				SourcePath:    src,
				Level:         level,
				Runs:          runs,
				CompareVisual: compare,
				KeepArchives:  keepArchives,
				OnProgress: func(stage string, current, total int, item string) {
					if outputFormat == "table" && !flagQuiet && !flagVerbose {
						fmt.Fprintf(ui.Out, "\r %s [%d/%d] %s: %-25s",
							ui.Cyan("⏳"), current, total, stage, ui.Dim(item))
					}
				},
			}

			report, err := a.BenchmarkUC.Execute(context.Background(), opts)
			if err != nil {
				if outputFormat == "table" {
					fmt.Fprintln(ui.Out)
				}
				return err
			}

			if outputFormat == "table" && !flagQuiet {
				fmt.Fprint(ui.Out, "\r\033[K")
			}

			switch outputFormat {
			case "json":
				return bench.RenderJSON(report, os.Stdout)
			case "csv":
				return bench.RenderCSV(report, os.Stdout)
			default:
				bench.RenderText(report, compare, os.Stdout)
				return nil
			}
		},
	}

	cmd.Flags().IntP("level", "l", -1, "Benchmark specific compression level (-1 for full sweep 0-9)")
	cmd.Flags().IntP("runs", "r", 1, "Number of benchmark iterations per level for statistical averaging")
	cmd.Flags().BoolP("compare", "c", false, "Generate comparative side-by-side trade-off matrix and bar charts")
	cmd.Flags().StringP("format", "f", "table", "Output report format: table, json, csv")
	cmd.Flags().BoolVar(&benchKeepArchives, "keep-archives", false, "Preserve temporary benchmark archive files for inspection")
	cmd.Flags().Bool("json", false, "Shortcut for --format json")
	cmd.Flags().Bool("csv", false, "Shortcut for --format csv")

	return cmd
}

var benchKeepArchives bool
