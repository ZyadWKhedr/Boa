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

// SuppressBanner prevents duplicate banner printing in interactive sessions.
var SuppressBanner = false

// PrettyPath converts long absolute home paths into compact ~/ paths for clean display.
func PrettyPath(p string) string {
	if p == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

// PrintBanner prints the exact Mole-styled ASCII header banner for Boa.
func PrintBanner() {
	if SuppressBanner {
		return
	}

	if NoColor {
		fmt.Fprintln(Out, ` ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _`+"`"+` |   
| |_) | (_) | (_| |   https://github.com/ZyadWKhedr/Boa
|____/ \___/ \__,_|   Tight, fast, lossless compression for your files.`)
		return
	}

	ascii := ` ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _` + "`" + ` |   
| |_) | (_) | (_| |   `
	lastLine := `|____/ \___/ \__,_|   `

	url := Cyan("https://github.com/ZyadWKhedr/Boa")
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
	fmt.Fprintf(Out, "%s\n", Bold(Cyan("▶ "+title)))
}

// RenderCompressionSummary prints a clean, beautifully formatted, easy-to-read summary card.
func RenderCompressionSummary(summary *types.ArchiveSummary, isDryRun bool) {
	if summary == nil {
		return
	}

	fmt.Fprintln(Out)
	if isDryRun {
		fmt.Fprintf(Out, " %s\n\n", Bold(Magenta("✦ Dry Run Preview (No files written)")))
	} else {
		fmt.Fprintf(Out, " %s\n\n", Bold(Green("✔ Compression Complete")))
	}

	prettyDest := PrettyPath(summary.ArchivePath)

	printAlignedRow("Archive", Bold(Cyan(prettyDest)))
	printAlignedRow("Original Size", stats.FormatBytes(summary.UncompressedBytes))

	if !isDryRun {
		printAlignedRow("Compressed", stats.FormatBytes(summary.CompressedBytes))
		
		savingsPercent := fmt.Sprintf("%.1f%% reduction", summary.SpaceSavedPercent)
		savingsStr := fmt.Sprintf("%s (%s)", stats.FormatBytes(summary.SpaceSavedBytes), Green(savingsPercent))
		printAlignedRow("Space Saved", savingsStr)

		ratioStr := fmt.Sprintf("%.2fx", 1.0/max(summary.CompressionRatio, 0.01))
		printAlignedRow("Ratio", Cyan(ratioStr))

		speedStr := stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration)
		printAlignedRow("Speed", speedStr)
	}

	avgStr := "-"
	if summary.AverageFileSize > 0 {
		avgStr = stats.FormatBytes(summary.AverageFileSize)
	}
	itemsStr := fmt.Sprintf("%d files (avg %s), %d folders", summary.TotalFiles, avgStr, summary.TotalDirs)
	printAlignedRow("Packed Items", itemsStr)
	printAlignedRow("Duration", stats.FormatDuration(summary.Duration))

	fmt.Fprintln(Out)
}

// RenderExtractionSummary prints a clean extraction summary card.
func RenderExtractionSummary(summary *types.ArchiveSummary, destDir string, isDryRun bool) {
	if summary == nil {
		return
	}

	fmt.Fprintln(Out)
	if isDryRun {
		fmt.Fprintf(Out, " %s\n\n", Bold(Magenta("✦ Extraction Preview (Dry Run)")))
	} else {
		fmt.Fprintf(Out, " %s\n\n", Bold(Green("✔ Extraction Complete")))
	}

	prettyDest := PrettyPath(destDir)
	prettySrc := PrettyPath(summary.ArchivePath)

	printAlignedRow("Source Archive", prettySrc)
	printAlignedRow("Extracted To", Bold(Cyan(prettyDest)))
	printAlignedRow("Total Size", stats.FormatBytes(summary.UncompressedBytes))

	avgStr := "-"
	if summary.AverageFileSize > 0 {
		avgStr = stats.FormatBytes(summary.AverageFileSize)
	}
	itemsStr := fmt.Sprintf("%d files (avg %s), %d folders", summary.TotalFiles, avgStr, summary.TotalDirs)
	printAlignedRow("Items", itemsStr)

	if !isDryRun {
		speedStr := stats.CalculateSpeed(summary.UncompressedBytes, summary.Duration)
		printAlignedRow("Speed", speedStr)
	}
	printAlignedRow("Duration", stats.FormatDuration(summary.Duration))

	fmt.Fprintln(Out)
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

	fmt.Fprintln(Out)
	prettyArchive := PrettyPath(summary.ArchivePath)
	printAlignedRow("Archive", Bold(Cyan(prettyArchive)))

	avgStr := "-"
	if summary.AverageFileSize > 0 {
		avgStr = stats.FormatBytes(summary.AverageFileSize)
	}
	printAlignedRow("Total Entries", fmt.Sprintf("%d files (avg %s), %d folders", summary.TotalFiles, avgStr, summary.TotalDirs))
	printAlignedRow("Original Size", stats.FormatBytes(summary.UncompressedBytes))
	printAlignedRow("Archive Size", stats.FormatBytes(summary.CompressedBytes))

	ratioStr := fmt.Sprintf("%.2fx (%.1f%% saved)", 1.0/max(summary.CompressionRatio, 0.01), summary.SpaceSavedPercent)
	printAlignedRow("Total Savings", Green(ratioStr))
	fmt.Fprintln(Out)
}

func printAlignedRow(label, value string) {
	fmt.Fprintf(Out, "   %-16s %s\n", Dim(label), value)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
