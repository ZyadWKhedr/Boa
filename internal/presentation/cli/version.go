package cli

import (
	"fmt"
	"runtime"

	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

func (a *App) newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v", "info"},
		Short:   "Display application version, build runtime & environment details",
		Run: func(cmd *cobra.Command, args []string) {
			if !flagQuiet {
				ui.PrintBanner()
				fmt.Println()
			}

			ui.PrintSection("Boa Runtime Environment")
			fmt.Printf("   %-16s %s\n", ui.Dim("Version"), ui.Bold(ui.Cyan(a.Version)))
			fmt.Printf("   %-16s %s\n", ui.Dim("Go Runtime"), runtime.Version())
			fmt.Printf("   %-16s %s/%s\n", ui.Dim("Platform"), runtime.GOOS, runtime.GOARCH)
			fmt.Printf("   %-16s %d logical threads\n", ui.Dim("Concurrency"), runtime.NumCPU())
			fmt.Printf("   %-16s %s\n", ui.Dim("Repository"), ui.Cyan("https://github.com/ZyadWKhedr/Boa"))
			fmt.Println()
		},
	}

	return cmd
}
