package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

var (
	flagUninstallForce bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "🗑 Safely remove Boa binaries, symlinks, and aliases from your system",
	Long: `Uninstall removes all installed Boa executable files, aliases (bo, compressor),
and installation references cleanly from ~/.local/bin and /usr/local/bin.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunUninstall(flagUninstallForce)
	},
}

func init() {
	uninstallCmd.Flags().BoolVarP(&flagUninstallForce, "force", "f", false, "Bypass confirmation prompt and uninstall immediately")
	RootCmd.AddCommand(uninstallCmd)
}

// AnimateBoaSnake displays a live slithering boa snake moving smoothly from left to right.
func AnimateBoaSnake(taskName string, steps int, stepDelay time.Duration) {
	trackWidth := 34
	snakeBodies := []string{
		"~=~=~=🐍",
		"=-=-=-🐍",
		"=~=~=~🐍",
		"-=-=-=🐍",
	}

	for i := 0; i <= steps; i++ {
		// Calculate position across the track
		pos := (i * (trackWidth - 7)) / steps
		if pos < 0 {
			pos = 0
		}
		if pos > trackWidth-7 {
			pos = trackWidth - 7
		}

		body := snakeBodies[i%len(snakeBodies)]
		leftPad := strings.Repeat(" ", pos)
		rightPad := strings.Repeat(" ", trackWidth-7-pos)

		// Render track with slithering green boa snake
		coloredSnake := ui.Bold(ui.Green(body))
		track := fmt.Sprintf("[%s%s%s]", leftPad, coloredSnake, rightPad)

		fmt.Printf("\r  %s  %s\033[K", track, ui.Dim(taskName))
		time.Sleep(stepDelay)
	}
	fmt.Printf("\r\033[K")
}

// RunUninstall performs interactive or forced uninstallation of Boa.
func RunUninstall(force bool) {
	if !force {
		fmt.Println()
		ui.PrintBanner()
		fmt.Println()
		ui.PrintSection("Boa Uninstallation")
		fmt.Println(" This will remove the Boa binary and all associated symlinks (bo, compressor)")
		fmt.Println(" from your system directories (~/.local/bin and /usr/local/bin).")
		fmt.Println()
		fmt.Print(ui.Bold(" Are you sure you want to uninstall Boa? [y/N]: "))

		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))

		if ans != "y" && ans != "yes" {
			fmt.Println(ui.Dim("\n Uninstallation cancelled. Boa is still installed.\n"))
			return
		}
	}

	fmt.Println()
	ui.PrintSection("Uninstalling Boa")
	fmt.Println()

	// Candidate paths to search and clean
	homeDir, _ := os.UserHomeDir()
	candidatePaths := []string{
		filepath.Join(homeDir, ".local", "bin", "boa"),
		filepath.Join(homeDir, ".local", "bin", "bo"),
		filepath.Join(homeDir, ".local", "bin", "compressor"),
		"/usr/local/bin/boa",
		"/usr/local/bin/bo",
		"/usr/local/bin/compressor",
	}

	// Also check current executable path
	if execPath, err := os.Executable(); err == nil {
		if realExec, err := filepath.EvalSymlinks(execPath); err == nil {
			candidatePaths = append(candidatePaths, realExec)
		}
		candidatePaths = append(candidatePaths, execPath)
	}

	// Phase 1: Scanning
	AnimateBoaSnake("Scanning installation directories...", 25, 30*time.Millisecond)

	// Phase 2: Removing binaries & symlinks
	var removed []string
	var failed []string

	AnimateBoaSnake("Slithering through system paths & removing binaries...", 35, 35*time.Millisecond)

	seen := make(map[string]bool)
	for _, path := range candidatePaths {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true

		// Check if file exists or symlink exists
		if _, err := os.Lstat(path); err == nil {
			if err := os.Remove(path); err == nil {
				removed = append(removed, path)
			} else {
				failed = append(failed, fmt.Sprintf("%s (%v)", path, err))
			}
		}
	}

	// Phase 3: Finalizing
	AnimateBoaSnake("Finalizing cleanup...", 20, 25*time.Millisecond)

	fmt.Println()
	if len(removed) > 0 {
		ui.PrintSuccess("Boa has been successfully uninstalled from your system.")
		fmt.Println()
		fmt.Println(ui.Bold(" Removed items:"))
		for _, item := range removed {
			fmt.Printf("   %s %s\n", ui.Green("✔"), ui.Dim(ui.PrettyPath(item)))
		}
	} else {
		ui.PrintInfo("No global binaries or symlinks found to remove.")
	}

	if len(failed) > 0 {
		fmt.Println()
		ui.PrintWarning("Some files could not be removed (permission required):")
		for _, item := range failed {
			fmt.Printf("   %s %s\n", ui.Red("✖"), ui.Dim(item))
		}
		fmt.Println(ui.Dim(" You may need to run with sudo: sudo rm <path>"))
	}

	fmt.Println()
	fmt.Println(ui.Cyan(" Thank you for using Boa! We're sad to see you go. 🐍"))
	fmt.Println()
}
