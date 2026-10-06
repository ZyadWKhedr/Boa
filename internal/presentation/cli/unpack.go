package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"compressor/internal/ui"
	"compressor/pkg/types"
	"github.com/spf13/cobra"
)

func (a *App) newUnpackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extract <archive.zip> [flags]",
		Aliases: []string{"unpack", "unzip", "x", "u", "d"},
		Short:   "Extract a zip archive safely with Zip-Slip defense",
		Long: `Extracts files from a .zip archive into the specified destination folder.
Features built-in Zip-Slip security defenses, path sanitization, and progress reporting.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer func() {
				_ = cmd.Flags().Set("output", "")
				_ = cmd.Flags().Set("force", "false")
			}()

			archivePath := args[0]

			output, _ := cmd.Flags().GetString("output")
			force, _ := cmd.Flags().GetBool("force")

			destDir := output
			if destDir == "" {
				cleanArc := filepath.Clean(archivePath)
				absArc, err := filepath.Abs(cleanArc)
				if err != nil {
					absArc = cleanArc
				}
				parentDir := filepath.Dir(absArc)
				baseName := strings.TrimSuffix(filepath.Base(absArc), filepath.Ext(absArc))
				destDir = filepath.Join(parentDir, baseName)
			}

			opts := types.UnpackOptions{
				ArchivePath:      archivePath,
				DestinationDir:   destDir,
				Overwrite:        force,
				Verbose:          flagVerbose,
				Quiet:            flagQuiet,
				ProgressCallback: func(file string, bytes int64, count int) {
					if flagVerbose && !flagQuiet {
						fmt.Fprintf(ui.Out, "  %s %s\n", ui.Dim("← Extracted:"), file)
					}
				},
			}

			summary, err := a.ExtractUC.Execute(context.Background(), opts)
			if err != nil {
				return err
			}

			if !flagQuiet {
				ui.RenderExtractionSummary(summary, destDir, false)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Destination directory for extracted files (defaults to folder named after archive)")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing destination files without prompt")

	return cmd
}
