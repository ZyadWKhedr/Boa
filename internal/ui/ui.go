package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"compressor/internal/stats"
	"compressor/pkg/types"
)

var Out io.Writer = os.Stdout

// PrintBanner prints the application header banner.
func PrintBanner() {
	if NoColor {
		fmt.Fprintln(Out, "--- Compressor CLI ---")
		return
	}
	fmt.Fprintln(Out, Bold(Cyan("⚡ Compressor"))+" "+Dim("• Fast, Secure, Cross-Platform Archive Engine"))
}

// PrintSuccess prints a formatted success message.
func PrintSuccess(msg string) {
	fmt.Fprintf(Out, "%s %s\n", BadgeOK("SUCCESS"), msg)
}

// PrintInfo prints a formatted informational message.
func PrintInfo(msg string) {
	fmt.Fprintf(Out, "%s %s\n", BadgeInfo("INFO"), msg)
}

// PrintWarning prints a formatted warning message.
func PrintWarning(msg string) {
	fmt.Fprintf(Out, "%s %s\n", BadgeWarn("WARN"), msg)
}

// PrintError prints a formatted error message.
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", BadgeErr("ERROR"), Red(msg))
}

// PrintSection prints a section header with stable formatting.
func PrintSection(title string) {
	fmt.Fprintf(Out, "\n%s\n", Bold(Cyan("▶ "+title)))
}

// RenderCompressionSummary prints a dense single-screen summary card following tw93/Mole style.
func RenderCompressionSummary(summary *types.ArchiveSummary, isDryRun bool) {
	if summary == nil {
		return
	}

	fmt.Fprintln(Out)
	if isDryRun {
		fmt.Fprintf(Out, "%s\n", BadgeDryRun("Compression Simulation Summary"))
	} else {
		fmt.Fprintf(Out, "%s\n", Bold(Green("✔ Compression Complete")))
	}

	div := Dim(strings.Repeat("─", 52))
	fmt.Fprintln(Out, div)

	printKV("Destination Archive", summary.ArchivePath)
	printKV("Files Included", fmt.Sprintf("%d files, %d folders", summary.TotalFiles, summary.TotalDirs))
	printKV("Original Total Size", stats.FormatBytes(summary.UncompressedBytes))

	if !isDryRun {
		printKV("Compressed Size", stats.FormatBytes(summary.CompressedBytes))
		savingsStr := fmt.Sprintf("%s (%.1f%% saved, ratio: %.2fx)",
			stats.FormatBytes(summary.SpaceSavedBytes),
			summary.SpaceSavedPercent,
			1.0/max(summary.CompressionRatio, 0.01))
		printKV("Space Reduction", Green(savingsStr))
	}

	printKV("Execution Duration", stats.FormatDuration(summary.Duration))
	if summary.Duration > 0 && summary.UncompressedBytes > 0 {
		printKV("Throughput Speed", stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration))
	}

	fmt.Fprintln(Out, div)
}

// RenderExtractionSummary prints a dense extraction summary card.
func RenderExtractionSummary(summary *types.ArchiveSummary, destDir string, isDryRun bool) {
	if summary == nil {
		return
	}

	fmt.Fprintln(Out)
	if isDryRun {
		fmt.Fprintf(Out, "%s\n", BadgeDryRun("Extraction Simulation Summary"))
	} else {
		fmt.Fprintf(Out, "%s\n", Bold(Green("✔ Extraction Complete")))
	}

	div := Dim(strings.Repeat("─", 52))
	fmt.Fprintln(Out, div)

	printKV("Source Archive", summary.ArchivePath)
	printKV("Target Directory", destDir)
	printKV("Extracted Items", fmt.Sprintf("%d files, %d folders", summary.TotalFiles, summary.TotalDirs))
	printKV("Total Unpacked Size", stats.FormatBytes(summary.UncompressedBytes))
	printKV("Execution Duration", stats.FormatDuration(summary.Duration))
	if summary.Duration > 0 && summary.UncompressedBytes > 0 {
		printKV("Throughput Speed", stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration))
	}

	fmt.Fprintln(Out, div)
}

// RenderArchiveList renders a dense file listing table.
func RenderArchiveList(summary *types.ArchiveSummary) {
	if summary == nil || len(summary.Entries) == 0 {
		PrintInfo("Archive is empty.")
		return
	}

	table := NewTable(
		Column{Title: "Type", Align: AlignCenter},
		Column{Title: "Permissions", Align: AlignLeft},
		Column{Title: "Original", Align: AlignRight},
		Column{Title: "Packed", Align: AlignRight},
		Column{Title: "Ratio", Align: AlignRight},
		Column{Title: "Modified Date", Align: AlignLeft},
		Column{Title: "Path", Align: AlignLeft},
	)

	for _, entry := range summary.Entries {
		entryType := "FILE"
		typeColor := Cyan("FILE")
		if entry.IsDir {
			entryType = "DIR"
			typeColor = Blue("DIR ")
		}

		permStr := os.FileMode(entry.Mode).String()
		origSize := stats.FormatBytes(entry.OriginalSize)
		compSize := stats.FormatBytes(entry.CompressedSize)
		
		var ratioStr string
		if entry.IsDir {
			origSize = "-"
			compSize = "-"
			ratioStr = "-"
		} else {
			pct := (1.0 - entry.CompressionRatio) * 100.0
			if pct < 0 {
				pct = 0
			}
			ratioStr = fmt.Sprintf("%.0f%%", pct)
		}

		modDate := entry.ModTime.Format("2006-01-02 15:04")
		table.AddRow(
			typeColor,
			Dim(permStr),
			origSize,
			compSize,
			Green(ratioStr),
			Dim(modDate),
			entry.Path,
		)
		_ = entryType
	}

	table.Render()

	// Dense summary footer
	div := Dim(strings.Repeat("─", 52))
	fmt.Fprintln(Out, div)
	printKV("Total Entries", fmt.Sprintf("%d files, %d directories", summary.TotalFiles, summary.TotalDirs))
	printKV("Raw Size", stats.FormatBytes(summary.UncompressedBytes))
	printKV("Archive Size", stats.FormatBytes(summary.CompressedBytes))
	printKV("Average Space Saved", Green(fmt.Sprintf("%.1f%% (ratio: %.2fx)", summary.SpaceSavedPercent, 1.0/max(summary.CompressionRatio, 0.01))))
	fmt.Fprintln(Out, div)
}

func printKV(key, value string) {
	fmt.Fprintf(Out, "  %-22s : %s\n", Dim(key), value)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
