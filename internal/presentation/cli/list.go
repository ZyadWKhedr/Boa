package cli

import (
	"context"

	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

func (a *App) newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <archive.zip>",
		Aliases: []string{"ls", "l", "inspect", "info"},
		Short:   "List files, sizes, and directory contents of a zip archive",
		Long:    `Inspects and lists all files contained within a .zip archive, showing uncompressed sizes, compressed sizes, and compression ratios.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			archivePath := args[0]

			if !flagQuiet {
				ui.PrintBanner()
			}

			summary, err := a.InspectUC.Execute(context.Background(), archivePath)
			if err != nil {
				return err
			}

			ui.RenderArchiveList(summary, flagVerbose)
			return nil
		},
	}

	return cmd
}
