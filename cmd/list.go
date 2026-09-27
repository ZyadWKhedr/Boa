package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"compressor/internal/extract"
	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:     "list <archive.zip>",
	Aliases: []string{"ls", "view", "inspect", "l"},
	Short:   "List contents, file sizes, and metadata of a zip archive",
	Long:    `Displays a clean, dense tabular breakdown of entries inside a zip archive without extracting it.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		archivePath := args[0]

		engine := extract.New()
		summary, err := engine.InspectArchive(archivePath)
		if err != nil {
			return err
		}

		if listJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(summary)
		}

		if !flagQuiet {
			ui.PrintBanner()
			ui.PrintInfo(fmt.Sprintf("Inspecting Archive: %s", ui.Bold(archivePath)))
			fmt.Println()
		}

		ui.RenderArchiveList(summary)
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output archive analysis in structured JSON format")
	RootCmd.AddCommand(listCmd)
}
