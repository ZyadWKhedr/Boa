package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"compressor/internal/compress"
	"compressor/internal/ui"
	"compressor/pkg/types"
	"github.com/spf13/cobra"
)

var (
	packOutput    string
	packLevel     int
	packExcludes  []string
	packForce     bool
	packDryRun    bool
)

var packCmd = &cobra.Command{
	Use:     "pack <source-folder-or-file> [flags]",
	Aliases: []string{"zip", "compress", "c", "p"},
	Short:   "Compress folders or files into a secure zip archive",
	Long: `Compresses directories or files into a .zip archive with customizable compression level,
exclusion filters (e.g. node_modules, .git), and streaming I/O for maximum performance.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sources := args
		primarySource := sources[0]

		// Derive default output zip name in the same parent directory as the primary source
		destZip := packOutput
		if destZip == "" {
			cleanSrc := filepath.Clean(primarySource)
			absSrc, err := filepath.Abs(cleanSrc)
			if err != nil {
				absSrc = cleanSrc
			}
			parentDir := filepath.Dir(absSrc)
			baseName := filepath.Base(absSrc)
			if baseName == "." || baseName == "/" || baseName == "\\" {
				destZip = filepath.Join(parentDir, "archive.zip")
			} else {
				nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
				destZip = filepath.Join(parentDir, nameWithoutExt+".zip")
			}
		}

		if !strings.HasSuffix(strings.ToLower(destZip), ".zip") {
			destZip += ".zip"
		}

		if !flagQuiet {
			ui.PrintBanner()
			if packDryRun {
				ui.PrintInfo(fmt.Sprintf("Simulating compression for %q -> %q (Level %d)", primarySource, destZip, packLevel))
			} else {
				ui.PrintInfo(fmt.Sprintf("Packing %q -> %q (Compression Level %d)...", primarySource, destZip, packLevel))
			}
		}

		engine := compress.New()
		opts := types.PackOptions{
			SourcePaths:      sources,
			OutputZipPath:    destZip,
			CompressionLevel: packLevel,
			ExcludePatterns:  packExcludes,
			Overwrite:        packForce,
			DryRun:           packDryRun,
			Verbose:          flagVerbose,
			Quiet:            flagQuiet,
			ProgressCallback: func(file string, bytes int64, count int) {
				if flagVerbose && !flagQuiet {
					fmt.Fprintf(ui.Out, "  %s %s\n", ui.Dim("→ Added:"), file)
				}
			},
		}

		summary, err := engine.Pack(opts)
		if err != nil {
			return err
		}

		if !flagQuiet {
			ui.RenderCompressionSummary(summary, packDryRun)
		}

		return nil
	},
}

func init() {
	packCmd.Flags().StringVarP(&packOutput, "output", "o", "", "Destination zip archive path (defaults to <source_name>.zip)")
	packCmd.Flags().IntVarP(&packLevel, "level", "l", 6, "Compression level: 0 (Store), 1 (Fastest), 6 (Default), 9 (Best)")
	packCmd.Flags().StringSliceVarP(&packExcludes, "exclude", "e", []string{".git*", ".DS_Store", "node_modules"}, "Glob patterns to exclude from archive")
	packCmd.Flags().BoolVarP(&packForce, "force", "f", false, "Overwrite existing destination archive without prompt")
	packCmd.Flags().BoolVar(&packDryRun, "dry-run", false, "Preview files to be included without creating archive")

	RootCmd.AddCommand(packCmd)
}
