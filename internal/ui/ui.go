package ui

import (
	"fmt"
	"io"
	"os"

	"compressor/internal/stats"
	"compressor/pkg/types"
)

var Out io.Writer = os.Stdout

// PrintBanner prints the exact Mole-styled ASCII header banner for Boa.
func PrintBanner() {
	if NoColor {
		fmt.Fprintln(Out, ` ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _`+"`"+` |   
| |_) | (_) | (_| |   https://github.com/zyadwael/boa
|____/ \___/ \__,_|   Tight, fast, lossless compression for your files.`)
		return
	}

	ascii := ` ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _` + "`" + ` |   
| |_) | (_) | (_| |   `
	lastLine := `|____/ \___/ \__,_|   `

	url := Cyan("https://github.com/zyadwael/boa")
	tagline := Green("Tight, fast, lossless compression for your files.")

	fmt.Fprintln(Out, Green(ascii)+url)
	fmt.Fprintln(Out, Green(lastLine)+tagline)
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

// HumanSpaceEquivalent gives a cute human-friendly translation like Mole ("That's like ~19 4K movies worth of space!").
func HumanSpaceEquivalent(savedBytes int64) string {
	if savedBytes <= 0 {
		return ""
	}
	const (
		gb = 1024 * 1024 * 1024
		mb = 1024 * 1024
	)

	if savedBytes >= 4*gb {
		movies := float64(savedBytes) / (4.5 * gb)
		return fmt.Sprintf("That's like ~%.0f 4K movies worth of space!", max(movies, 1))
	} else if savedBytes >= 500*mb {
		albums := float64(savedBytes) / (120 * mb)
		return fmt.Sprintf("That's like ~%.0f lossless music albums worth of space!", max(albums, 1))
	} else if savedBytes >= 10*mb {
		photos := float64(savedBytes) / (3.5 * mb)
		return fmt.Sprintf("That's like ~%.0f high-res RAW photos worth of space!", max(photos, 1))
	}
	docs := float64(savedBytes) / (50 * 1024)
	return fmt.Sprintf("That's like ~%.0f text documents worth of space!", max(docs, 1))
}

// RenderCompressionSummary prints a summary card styled after Mole's aesthetic double-bordered cards.
func RenderCompressionSummary(summary *types.ArchiveSummary, isDryRun bool) {
	if summary == nil {
		return
	}

	div := Dim("========================================================================")
	fmt.Fprintln(Out)
	fmt.Fprintln(Out, div)

	if isDryRun {
		fmt.Fprintf(Out, " %s\n", Bold(Magenta("DRY RUN COMPLETE! (Simulation Mode)")))
		fmt.Fprintf(Out, " 📍 Target Archive: %s\n", Bold(Cyan(summary.ArchivePath)))
		avgStr := "-"
		if summary.AverageFileSize > 0 {
			avgStr = stats.FormatBytes(summary.AverageFileSize)
		}
		fmt.Fprintf(Out, " Potential space: %s | Files analyzed: %d (avg size: %s) | Categories: %d folders\n",
			Green(stats.FormatBytes(summary.UncompressedBytes)),
			summary.TotalFiles,
			avgStr,
			summary.TotalDirs,
		)
	} else {
		fmt.Fprintf(Out, " %s\n", Bold(Green("COMPRESSION COMPLETE!")))
		fmt.Fprintf(Out, " 📍 Saved To: %s\n", Bold(Cyan(summary.ArchivePath)))

		ratioStr := fmt.Sprintf("%.2fx (%.1f%% saved)", 1.0/max(summary.CompressionRatio, 0.01), summary.SpaceSavedPercent)
		speedStr := stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration)
		
		fmt.Fprintf(Out, " Space saved: %s | Compression ratio: %s | Speed: %s\n",
			Bold(Green(stats.FormatBytes(summary.SpaceSavedBytes))),
			Cyan(ratioStr),
			Cyan(speedStr),
		)

		if equiv := HumanSpaceEquivalent(summary.SpaceSavedBytes); equiv != "" {
			fmt.Fprintf(Out, " %s\n", Italic(Green(equiv)))
		}

		avgStr := "-"
		if summary.AverageFileSize > 0 {
			avgStr = stats.FormatBytes(summary.AverageFileSize)
		}
		fmt.Fprintf(Out, " Files packed: %d (avg file size: %s) | Categories: %d folders | Time: %s\n",
			summary.TotalFiles,
			Cyan(avgStr),
			summary.TotalDirs,
			stats.FormatDuration(summary.Duration),
		)
	}

	fmt.Fprintln(Out, div)
}

// RenderExtractionSummary prints an extraction summary card.
func RenderExtractionSummary(summary *types.ArchiveSummary, destDir string, isDryRun bool) {
	if summary == nil {
		return
	}

	div := Dim("========================================================================")
	fmt.Fprintln(Out)
	fmt.Fprintln(Out, div)

	if isDryRun {
		fmt.Fprintf(Out, " %s\n", Bold(Magenta("EXTRACTION PREVIEW (Dry Run)")))
		fmt.Fprintf(Out, " 📍 Target Directory: %s\n", Bold(Cyan(destDir)))
		fmt.Fprintf(Out, " Target items: %d files, %d folders | Total uncompressed: %s\n",
			summary.TotalFiles, summary.TotalDirs, Green(stats.FormatBytes(summary.UncompressedBytes)))
	} else {
		fmt.Fprintf(Out, " %s\n", Bold(Green("EXTRACTION COMPLETE!")))
		fmt.Fprintf(Out, " 📍 Extracted To: %s\n", Bold(Cyan(destDir)))
		speedStr := stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration)
		fmt.Fprintf(Out, " Total unpacked: %s | Speed: %s | Time: %s\n",
			Bold(Green(stats.FormatBytes(summary.UncompressedBytes))),
			Cyan(speedStr),
			stats.FormatDuration(summary.Duration),
		)
		avgStr := "-"
		if summary.AverageFileSize > 0 {
			avgStr = stats.FormatBytes(summary.AverageFileSize)
		}
		fmt.Fprintf(Out, " Extracted %d files (avg file size: %s) into destination.\n",
			summary.TotalFiles,
			Cyan(avgStr),
		)
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
		typeColor := Cyan("FILE")
		if entry.IsDir {
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
	}

	table.Render()

	div := Dim("========================================================================")
	fmt.Fprintln(Out, div)
	fmt.Fprintf(Out, " 📍 Archive: %s\n", Bold(Cyan(summary.ArchivePath)))
	avgStr := "-"
	if summary.AverageFileSize > 0 {
		avgStr = stats.FormatBytes(summary.AverageFileSize)
	}
	ratioStr := fmt.Sprintf("%.2fx (%.1f%% saved)", 1.0/max(summary.CompressionRatio, 0.01), summary.SpaceSavedPercent)
	fmt.Fprintf(Out, " Total: %d files (avg size: %s), %d folders | Raw: %s | Packed: %s | Ratio: %s\n",
		summary.TotalFiles, Cyan(avgStr), summary.TotalDirs,
		stats.FormatBytes(summary.UncompressedBytes),
		stats.FormatBytes(summary.CompressedBytes),
		Green(ratioStr),
	)
	fmt.Fprintln(Out, div)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
