package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"compressor/internal/infrastructure/updater"
	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

func (a *App) newUpdateCmd() *cobra.Command {
	var checkOnly bool
	var forceYes bool

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade", "self-update"},
		Short:   "Check for and automatically install the latest Boa release",
		Long: `Update queries GitHub Releases for the newest available version of Boa.
If a newer release is found, it downloads the pre-built binary for your current
OS and architecture, verifies and installs it, updating 'bo' and 'compressor' aliases.`,
		Run: func(cmd *cobra.Command, args []string) {
			RunUpdate(cmd.Context(), a.Version, checkOnly, forceYes)
		},
	}

	cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check for updates without downloading or installing")
	cmd.Flags().BoolVarP(&forceYes, "yes", "y", false, "Automatically accept prompts and install update")

	return cmd
}

// RunUpdate performs update checking and binary installation.
func RunUpdate(ctx context.Context, currentVersion string, checkOnly, forceYes bool) {
	if !flagQuiet {
		ui.PrintBanner()
		fmt.Println()
	}

	AnimateBoaSnake("Checking GitHub for updates...", 25, 25*time.Millisecond)

	rel, isNewer, err := updater.CheckForUpdate(ctx, currentVersion)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to check for updates: %v", err))
		fmt.Println(ui.Dim(" Tip: Verify your internet connection or check https://github.com/ZyadWKhedr/Boa/releases"))
		return
	}

	if !isNewer {
		ui.PrintSuccess(fmt.Sprintf("Boa is up to date! (Version %s)", ui.Bold(ui.Cyan(currentVersion))))
		fmt.Printf("   %-16s %s\n", ui.Dim("Latest Release"), ui.Cyan(rel.TagName))
		if !rel.PublishedAt.IsZero() {
			fmt.Printf("   %-16s %s\n", ui.Dim("Published Date"), rel.PublishedAt.Format("2006-01-02"))
		}
		fmt.Println()
		return
	}

	ui.PrintSection(fmt.Sprintf("New Version Available: %s -> %s", ui.Dim(currentVersion), ui.Bold(ui.Green(rel.TagName))))
	fmt.Printf("   %-16s %s\n", ui.Dim("Current Version"), currentVersion)
	fmt.Printf("   %-16s %s\n", ui.Dim("Latest Version"), ui.Bold(ui.Green(rel.TagName)))
	if !rel.PublishedAt.IsZero() {
		fmt.Printf("   %-16s %s\n", ui.Dim("Release Date"), rel.PublishedAt.Format("2006-01-02"))
	}
	if rel.HTMLURL != "" {
		fmt.Printf("   %-16s %s\n", ui.Dim("Release URL"), ui.Cyan(rel.HTMLURL))
	}
	fmt.Println()

	if checkOnly {
		ui.PrintInfo(fmt.Sprintf("Run '%s' to download and install this update.", ui.Bold("boa update")))
		return
	}

	if !forceYes {
		fmt.Print(ui.Bold(" Do you want to download and install this update now? [Y/n]: "))
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans == "n" || ans == "no" {
			fmt.Println(ui.Dim("\n Update cancelled. You remain on version " + currentVersion + ".\n"))
			return
		}
	}

	fmt.Println()
	pb := ui.NewProgressBar("Downloading Boa "+rel.TagName, 0, 0)
	installedPath, err := updater.DownloadAndInstall(ctx, rel, func(downloaded, total int64) {
		pb.Update(rel.TagName, downloaded, 0)
	})
	pb.Finish()

	if err != nil {
		ui.PrintError(fmt.Sprintf("Update failed: %v", err))
		fmt.Println(ui.Dim(" You can also update using: go install github.com/ZyadWKhedr/Boa/cmd/boa@latest"))
		return
	}

	AnimateBoaSnake("Configuring permissions & updating aliases...", 25, 25*time.Millisecond)

	fmt.Println()
	ui.PrintSuccess(fmt.Sprintf("Boa has been successfully updated to %s! 🐍", ui.Bold(ui.Green(rel.TagName))))
	fmt.Printf("   %-16s %s\n", ui.Dim("Installed At"), ui.Bold(ui.PrettyPath(installedPath)))
	fmt.Printf("   %-16s %s, %s\n", ui.Dim("Aliases"), ui.Cyan("bo"), ui.Cyan("compressor"))
	fmt.Println()
}
