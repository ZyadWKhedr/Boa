package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"compressor/internal/domain"
	"compressor/internal/infrastructure/classifier"
	"compressor/internal/ui"
	"compressor/internal/usecase"
	"compressor/pkg/types"
	"github.com/spf13/cobra"
)

func (a *App) newPackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "compress <source-folder-or-file> [flags]",
		Aliases: []string{"pack", "zip", "c", "p"},
		Short:   "Compress folders or files into a secure zip archive",
		Long: `Compresses directories or files into a .zip archive with selectable compression method (deflate, store, zstd),
customizable compression levels, optional lossy media flags (--lossy images,audio,video), exclusion filters, and streaming I/O.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer func() {
				_ = cmd.Flags().Set("output", "")
				_ = cmd.Flags().Set("method", "deflate")
				_ = cmd.Flags().Set("level", "-1")
				_ = cmd.Flags().Set("force", "false")
				_ = cmd.Flags().Set("dry-run", "false")
				_ = cmd.Flags().Set("lossy", "")
				_ = cmd.Flags().Set("quality", "80")
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
			lossyCategories, _ := cmd.Flags().GetStringSlice("lossy")
			quality, _ := cmd.Flags().GetInt("quality")

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

			var pb *ui.ProgressBar
			if !flagQuiet && !dryRun && !flagVerbose {
				pb = ui.NewProgressBar("Compressing", 0, 0)
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
					if pb != nil {
						pb.Update(file, bytes, count)
					} else if flagVerbose && !flagQuiet {
						fmt.Fprintf(ui.Out, "  %s %s\n", ui.Dim("→ Added:"), file)
					}
				},
			}

			if dryRun {
				summary, err := a.PackUC.Execute(context.Background(), opts)
				if err != nil {
					return err
				}
				if !flagQuiet {
					ui.RenderCompressionSummary(summary, true)
				}
				return nil
			}

			// If lossy flags are explicitly provided, use the Plan-guided workflow
			if len(lossyCategories) > 0 {
				allowImages := false
				allowAudio := false
				allowVideo := false

				for _, cat := range lossyCategories {
					switch strings.ToLower(strings.TrimSpace(cat)) {
					case "image", "images", "img":
						allowImages = true
					case "audio", "sound", "music":
						allowAudio = true
					case "video", "videos", "vid":
						allowVideo = true
					case "all", "media":
						allowImages = true
						allowAudio = true
						allowVideo = true
					}
				}

				scanUC := usecase.NewScanPathUseCase(classifier.NewMagicClassifier())
				scanRes, err := scanUC.Execute(context.Background(), primarySource, excludes)
				if err != nil {
					return err
				}

				policy := domain.NewLossyEligibilityPolicy()
				estimator := usecase.NewEstimateSavingsUseCase()
				planUC := usecase.NewBuildCompressionPlanUseCase(policy, estimator)

				var domainMethod domain.CompressionMethod
				switch method {
				case types.MethodZstd:
					domainMethod = domain.MethodZstd
				case types.MethodStore:
					domainMethod = domain.MethodStore
				default:
					domainMethod = domain.MethodDeflate
				}

				prefs := domain.UserPreferences{
					ArchiveMethod: domainMethod,
					DefaultLevel:  effectiveLevel,
					Images: domain.CategoryPreferences{
						EnableLossy: allowImages,
						Quality:     quality,
					},
					Audio: domain.CategoryPreferences{
						EnableLossy: allowAudio,
						Quality:     quality,
					},
					Video: domain.CategoryPreferences{
						EnableLossy: allowVideo,
						Quality:     quality,
					},
				}

				plan, err := planUC.Execute(context.Background(), scanRes, prefs, destZip)
				if err != nil {
					return err
				}

				if pb != nil {
					pb.Finish()
					pb = ui.NewProgressBar("Compressing", plan.TotalFiles, plan.TotalBytes)
				}

				summary, err := a.PackUC.ExecutePlan(context.Background(), plan, opts)
				if pb != nil {
					pb.Finish()
				}
				if err != nil {
					return err
				}

				if !flagQuiet {
					ui.RenderCompressionSummary(summary, false)
				}
				return nil
			}

			// Standard lossless pack
			summary, err := a.PackUC.Execute(context.Background(), opts)
			if pb != nil {
				pb.Finish()
			}
			if err != nil {
				return err
			}

			if !flagQuiet {
				ui.RenderCompressionSummary(summary, false)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Destination zip archive path (defaults to <source_name>.zip)")
	cmd.Flags().StringP("method", "m", "deflate", "Compression method: deflate (default), store, zstd")
	cmd.Flags().IntP("level", "l", -1, "Compression level (Deflate: 0-9 [default: 6], Zstandard: 1-11 [default: 3])")
	cmd.Flags().StringSlice("lossy", nil, "Explicitly enable lossy compression for media categories (images, audio, video)")
	cmd.Flags().Int("quality", 80, "Lossy quality setting (0-100, default: 80)")
	cmd.Flags().StringSliceP("exclude", "e", []string{".git*", ".DS_Store", "node_modules"}, "Glob patterns to exclude from archive")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing destination archive without prompt")
	cmd.Flags().Bool("dry-run", false, "Preview files to be included without creating archive")

	return cmd
}
