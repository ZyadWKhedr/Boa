package bench

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"unicode/utf8"

	"compressor/internal/stats"
	"compressor/internal/ui"
)

// RenderText formats and prints the human-readable benchmark report to the provided writer.
func RenderText(report *Report, showCompare bool, out io.Writer) {
	if report == nil {
		return
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, " %s\n\n", ui.Bold(ui.Cyan(report.Title)))

	// 1. Input Section
	fmt.Fprintf(out, " %s\n", ui.Bold("Input"))
	fmt.Fprintf(out, " %s\n", ui.Dim(strings.Repeat("─", 44)))
	fmt.Fprintf(out, "   %-14s %s\n", ui.Dim("Path:"), ui.PrettyPath(report.Input.Path))
	fmt.Fprintf(out, "   %-14s %s\n", ui.Dim("Files:"), formatNumber(report.Input.TotalFiles))
	if report.Input.TotalDirs > 0 {
		fmt.Fprintf(out, "   %-14s %s\n", ui.Dim("Directories:"), formatNumber(report.Input.TotalDirs))
	}
	fmt.Fprintf(out, "   %-14s %s\n", ui.Dim("Input size:"), ui.Bold(report.Input.SizeHuman))
	if report.RunsPerLevel > 1 {
		fmt.Fprintf(out, "   %-14s %d iterations per level (averages shown)\n", ui.Dim("Runs:"), report.RunsPerLevel)
	}
	if report.PreservedDir != "" {
		fmt.Fprintf(out, "   %-14s %s\n", ui.Dim("Preserved at:"), ui.Cyan(report.PreservedDir))
	}
	fmt.Fprintln(out)

	if report.IsSmallDataset {
		fmt.Fprintf(out, " %s %s\n\n", ui.Dim("✦ Note:"), ui.Dim("Dataset is small (< 10 KB); throughput metrics reflect CPU/timer granularity."))
	}

	// 2. Results Section
	fmt.Fprintf(out, " %s\n", ui.Bold("Results"))
	fmt.Fprintf(out, " %s\n", ui.Dim(strings.Repeat("─", 62)))

	var cols []ui.Column
	hasDecomp := len(report.Results) > 0 && report.Results[0].DecompDurationAvg > 0

	cols = append(cols,
		ui.Column{Title: "Level", Align: ui.AlignLeft},
		ui.Column{Title: "Time", Align: ui.AlignRight},
		ui.Column{Title: "Size", Align: ui.AlignRight},
		ui.Column{Title: "Saved", Align: ui.AlignRight},
		ui.Column{Title: "Ratio", Align: ui.AlignRight},
		ui.Column{Title: "Speed", Align: ui.AlignRight},
	)

	if hasDecomp {
		cols = append(cols,
			ui.Column{Title: "Decomp Time", Align: ui.AlignRight},
			ui.Column{Title: "Decomp Speed", Align: ui.AlignRight},
		)
	}

	table := ui.NewTable(cols...)
	table.SetOutput(out)

	for _, res := range report.Results {
		levelStr := fmt.Sprintf("%d", res.Level)
		if res.Level == report.Summary.BestBalanceLevel {
			levelStr = ui.Bold(ui.Cyan(fmt.Sprintf("%d ★", res.Level)))
		} else if res.Level == 0 {
			levelStr = fmt.Sprintf("%d (store)", res.Level)
		}

		timeStr := stats.FormatDuration(res.DurationAvg)
		sizeStr := res.CompressedHuman
		savedStr := fmt.Sprintf("%.1f%%", res.SpaceSavedPercent)
		if res.SpaceSavedPercent > 0 {
			savedStr = ui.Green(savedStr)
		} else {
			savedStr = ui.Dim(savedStr)
		}

		ratioStr := fmt.Sprintf("%.2fx", res.RatioMultiplier)
		speedStr := ui.Cyan(res.ThroughputStr)

		row := []string{
			levelStr,
			timeStr,
			sizeStr,
			savedStr,
			ratioStr,
			speedStr,
		}

		if hasDecomp {
			decompTimeStr := stats.FormatDuration(res.DecompDurationAvg)
			decompSpeedStr := res.DecompThroughputS
			row = append(row, decompTimeStr, decompSpeedStr)
		}

		table.AddRow(row...)
	}

	table.Render()
	fmt.Fprintln(out)

	// 3. Visual Comparison Section (if requested or compare mode)
	if showCompare && len(report.Results) > 1 {
		renderVisualComparison(report, out)
	}

	// 4. Summary & Trade-off Insights
	if len(report.Results) > 1 {
		fmt.Fprintf(out, " %s\n", ui.Bold("Summary"))
		fmt.Fprintf(out, " %s\n", ui.Dim(strings.Repeat("─", 44)))
		fmt.Fprintf(out, "   %-16s Level %-2d  • %s\n", ui.Dim("Fastest:"), report.Summary.FastestLevel, report.Summary.FastestSpeed)
		fmt.Fprintf(out, "   %-16s Level %-2d  • %s\n", ui.Dim("Smallest:"), report.Summary.SmallestLevel, report.Summary.SmallestSize)
		fmt.Fprintf(out, "   %-16s Level %-2d  • %s\n", ui.Dim("Best balance:"), report.Summary.BestBalanceLevel, report.Summary.BestBalanceDetail)
		fmt.Fprintln(out)

		if len(report.Summary.TradeOffInsights) > 0 {
			for _, insight := range report.Summary.TradeOffInsights {
				fmt.Fprintf(out, "   %s %s\n", ui.Dim("•"), insight)
			}
			fmt.Fprintln(out)
		}
	}
}

func renderVisualComparison(report *Report, out io.Writer) {
	fmt.Fprintf(out, " %s\n", ui.Bold("Visual Comparison"))
	fmt.Fprintf(out, " %s\n\n", ui.Dim(strings.Repeat("─", 44)))

	// Size bar chart
	fmt.Fprintf(out, " %s\n", ui.Dim("Compression Size (smaller is better)"))
	var maxSize int64 = 1
	for _, res := range report.Results {
		if res.CompressedBytes > maxSize {
			maxSize = res.CompressedBytes
		}
	}
	const barWidth = 28
	for _, res := range report.Results {
		ratio := float64(res.CompressedBytes) / float64(maxSize)
		filled := int(math.Round(ratio * float64(barWidth)))
		if filled < 1 && res.CompressedBytes > 0 {
			filled = 1
		}
		bar := strings.Repeat("█", filled)
		padding := strings.Repeat(" ", barWidth-filled)
		fmt.Fprintf(out, "   L%-2d  %s%s %s\n", res.Level, ui.Cyan(bar), padding, res.CompressedHuman)
	}
	fmt.Fprintln(out)

	// Speed bar chart
	fmt.Fprintf(out, " %s\n", ui.Dim("Compression Speed (higher is faster)"))
	var maxSpeed float64 = 0.001
	for _, res := range report.Results {
		if res.ThroughputMBps > maxSpeed {
			maxSpeed = res.ThroughputMBps
		}
	}
	for _, res := range report.Results {
		ratio := res.ThroughputMBps / maxSpeed
		filled := int(math.Round(ratio * float64(barWidth)))
		if filled < 1 && res.ThroughputMBps > 0 {
			filled = 1
		}
		bar := strings.Repeat("█", filled)
		padding := strings.Repeat(" ", barWidth-filled)
		fmt.Fprintf(out, "   L%-2d  %s%s %s\n", res.Level, ui.Green(bar), padding, res.ThroughputStr)
	}
	fmt.Fprintln(out)
}

// RenderJSON serializes the benchmark report into structured JSON.
func RenderJSON(report *Report, out io.Writer) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// RenderCSV exports the benchmark level results into standard CSV format.
func RenderCSV(report *Report, out io.Writer) error {
	writer := csv.NewWriter(out)
	defer writer.Flush()

	hasDecomp := len(report.Results) > 0 && report.Results[0].DecompDurationAvg > 0

	header := []string{
		"level",
		"duration_ms",
		"compressed_size_bytes",
		"ratio_multiplier",
		"space_saved_percent",
		"throughput_mbps",
	}

	if hasDecomp {
		header = append(header, "decomp_duration_ms", "decomp_throughput_mbps")
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	for _, res := range report.Results {
		row := []string{
			fmt.Sprintf("%d", res.Level),
			fmt.Sprintf("%.2f", res.DurationMs),
			fmt.Sprintf("%d", res.CompressedBytes),
			fmt.Sprintf("%.2f", res.RatioMultiplier),
			fmt.Sprintf("%.2f", res.SpaceSavedPercent),
			fmt.Sprintf("%.2f", res.ThroughputMBps),
		}

		if hasDecomp {
			row = append(row,
				fmt.Sprintf("%.2f", res.DecompDurationMs),
				fmt.Sprintf("%.2f", res.DecompThroughput),
			)
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func formatNumber(n int) string {
	in := fmt.Sprintf("%d", n)
	numOfDigits := utf8.RuneCountInString(in)
	if numOfDigits <= 3 {
		return in
	}

	var sb strings.Builder
	for i, r := range in {
		sb.WriteRune(r)
		remaining := numOfDigits - 1 - i
		if remaining > 0 && remaining%3 == 0 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}
