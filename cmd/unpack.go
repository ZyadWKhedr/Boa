package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"compressor/internal/extract"
	"compressor/internal/ui"
	"compressor/pkg/types"
	"github.com/spf13/cobra"
)

var (
	unpackOutput   string
	unpackExcludes []string
	unpackForce    bool
	unpackDryRun   bool
)

var unpackCmd = &cobra.Command{
	Use:     "unpack <archive.zip> [destination-folder] [flags]",
	Aliases: []string{"unzip", "extract", "x", "u"},
	Short:   "Safely decompress a zip archive into a directory",
	Long: `Extracts files from a .zip archive into the specified destination folder.
Enforces strict Zip Slip path traversal security, symlink boundary checks, and streaming I/O.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		archivePath := args[0]
		destDir := unpackOutput

		// If destination directory not provided as flag, check secondary positional argument
		if destDir == "" && len(args) > 1 {
			destDir = args[1]
		}

		// If still empty, default to folder named after the zip archive or current directory
		if destDir == "" {
			base := filepath.Base(archivePath)
			destDir = strings.TrimSuffix(base, filepath.Ext(base))
			if destDir == "" || destDir == "." {
				destDir = "./extracted"
			}
		}

		if !flagQuiet {
			ui.PrintBanner()
			if unpackDryRun {
				ui.PrintInfo(fmt.Sprintf("Simulating extraction for %q -> %q", archivePath, destDir))
			} else {
				ui.PrintInfo(fmt.Sprintf("Unpacking %q -> %q (Safe Mode active)...", archivePath, destDir))
			}
		}

		engine := extract.New()
		opts := types.UnpackOptions{
			ArchivePath:     archivePath,
			DestinationDir:  destDir,
			ExcludePatterns: unpackExcludes,
			Overwrite:       unpackForce,
			DryRun:          unpackDryRun,
			Verbose:         flagVerbose,
			Quiet:           flagQuiet,
			SafeMode:        true,
			ProgressCallback: func(file string, bytes int64, count int) {
				if flagVerbose && !flagQuiet {
					fmt.Fprintf(ui.Out, "  %s %s\n", ui.Dim("← Extracted:"), file)
				}
			},
		}

		summary, err := engine.Unpack(opts)
		if err != nil {
			return err
		}

		if !flagQuiet {
			ui.RenderExtractionSummary(summary, destDir, unpackDryRun)
		}

		return nil
	},
}

func init() {
	unpackCmd.Flags().StringVarP(&unpackOutput, "output", "o", "", "Destination directory for extracted files")
	unpackCmd.Flags().StringSliceVarP(&unpackExcludes, "exclude", "e", nil, "Glob patterns to exclude from extraction")
	unpackCmd.Flags().BoolVarP(&unpackForce, "force", "f", false, "Overwrite existing files in destination directory")
	unpackCmd.Flags().BoolVar(&unpackDryRun, "dry-run", false, "Preview files to be extracted without writing to disk")

	RootCmd.AddCommand(unpackCmd)
}
