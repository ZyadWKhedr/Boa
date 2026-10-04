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

func (a *App) newPackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pack <source-folder-or-file> [flags]",
		Aliases: []string{"zip", "compress", "c", "p"},
		Short:   "Compress folders or files into a secure zip archive",
		Long: `Compresses directories or files into a .zip archive with selectable compression method (deflate, store, zstd),
customizable compression levels, exclusion filters (e.g. node_modules, .git), and streaming I/O for maximum performance.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer func() {
				_ = cmd.Flags().Set("output", "")
				_ = cmd.Flags().Set("method", "deflate")
				_ = cmd.Flags().Set("level", "-1")
				_ = cmd.Flags().Set("force", "false")
				_ = cmd.Flags().Set("dry-run", "false")
				_ = cmd.Flags().Set("exclude", ".git*,.DS_Store,node_modules")
			}()

			sources := args
			primarySource := sources[0]

			output, _ := cmd.Flags().GetString("output")
			methodStr, _ := cmd.Flags().GetString("method")
			level, _ := cmd.Flags().GetInt("level")
			excludes, _ := cmd.Flags().GetStringSlice("exclude")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			method, err := types.ParseMethod(methodStr)
			if err != nil {
				return err
			}

			effectiveLevel, err := types.ValidateLevel(method, level)
			if err != nil {
				return err
			}

			destZip := output
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
			}

			opts := types.PackOptions{
				SourcePaths:      sources,
				OutputZipPath:    destZip,
				Method:           method,
				CompressionLevel: effectiveLevel,
				ExcludePatterns:  excludes,
				Overwrite:        force,
				DryRun:           dryRun,
				Verbose:          flagVerbose,
				Quiet:            flagQuiet,
				ProgressCallback: func(file string, bytes int64, count int) {
					if flagVerbose && !flagQuiet {
						fmt.Fprintf(ui.Out, "  %s %s\n", ui.Dim("→ Added:"), file)
					}
				},
			}

			summary, err := a.PackUC.Execute(context.Background(), opts)
			if err != nil {
				return err
			}

			if !flagQuiet {
				ui.RenderCompressionSummary(summary, dryRun)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Destination zip archive path (defaults to <source_name>.zip)")
	cmd.Flags().StringP("method", "m", "deflate", "Compression method: deflate (default), store, zstd")
	cmd.Flags().IntP("level", "l", -1, "Compression level (Deflate: 0-9 [default: 6], Zstandard: 1-11 [default: 3])")
	cmd.Flags().StringSliceP("exclude", "e", []string{".git*", ".DS_Store", "node_modules"}, "Glob patterns to exclude from archive")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing destination archive without prompt")
	cmd.Flags().Bool("dry-run", false, "Preview files to be included without creating archive")

	return cmd
}
