package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"compressor/internal/compress"
	"compressor/internal/extract"
	"compressor/internal/ui"
)

// RunInteractiveDashboard launches the tw93/Mole-inspired interactive terminal dashboard.
func RunInteractiveDashboard() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Clear screen or print clean separator
		fmt.Println()
		ui.PrintBanner()
		fmt.Println()
		fmt.Printf(" %s\n\n", ui.Dim("Update 1.0.0 available, run bo update"))

		fmt.Printf("   %s  %-12s %s\n", ui.Bold(ui.Cyan("1.")), "Pack", ui.Dim("Compress folders into dense archives"))
		fmt.Printf("   %s  %-12s %s\n", ui.Bold(ui.Cyan("2.")), "Unpack", ui.Dim("Safely extract zip archives"))
		fmt.Printf("   %s  %-12s %s\n", ui.Bold(ui.Cyan("3.")), "Inspect", ui.Dim("Explore archive structure & metadata"))
		fmt.Printf("   %s  %-12s %s\n", ui.Bold(ui.Cyan("4.")), "Benchmark", ui.Dim("Compare compression speed & ratios"))
		fmt.Printf(" %s %s  %-12s %s\n\n", ui.Bold(ui.Cyan("➤")), ui.Bold(ui.Cyan("5.")), ui.Bold(ui.Cyan("Status")), ui.Cyan("Runtime health & system info"))

		fmt.Println(ui.Dim(" ⇅  |  Enter  |  H Help  |  V Version  |  Q Quit"))
		fmt.Println()

		fmt.Print(ui.Bold(ui.Cyan(" Select option: ")))
		choice, _ := reader.ReadString('\n')
		choice = strings.ToLower(strings.TrimSpace(choice))

		switch choice {
		case "1", "pack", "p", "zip":
			interactivePack(reader)
		case "2", "unpack", "u", "unzip", "x":
			interactiveUnpack(reader)
		case "3", "inspect", "i", "list", "ls", "l":
			interactiveList(reader)
		case "4", "benchmark", "bench", "b":
			interactiveBench(reader)
		case "5", "status", "s", "v", "version":
			fmt.Println()
			RootCmd.SetArgs([]string{"version"})
			_ = RootCmd.Execute()
		case "h", "help", "?":
			fmt.Println()
			_ = RootCmd.Help()
		case "q", "quit", "exit", "0":
			fmt.Println(ui.Dim("\n Goodbye! 🐍"))
			return
		case "":
			// Default option is 5 (Status) when pressing enter directly
			fmt.Println()
			RootCmd.SetArgs([]string{"version"})
			_ = RootCmd.Execute()
		default:
			ui.PrintWarning("Invalid option. Enter 1-5 or Q to quit.")
		}

		fmt.Println()
		fmt.Print(ui.Dim(" Press Enter to return to main menu..."))
		_, _ = reader.ReadString('\n')
	}
}

func interactivePack(reader *bufio.Reader) {
	fmt.Println()
	ui.PrintSection("Interactive Pack (Folder / File Compression)")

	fmt.Print(ui.Bold(" Enter source folder or file path: "))
	src, _ := reader.ReadString('\n')
	src = strings.TrimSpace(src)
	if src == "" {
		ui.PrintError("Source path cannot be empty.")
		return
	}

	fmt.Print(ui.Bold(" Enter destination archive name (leave blank for auto-name): "))
	dest, _ := reader.ReadString('\n')
	dest = strings.TrimSpace(dest)

	fmt.Print(ui.Bold(" Compression level [0=Store, 1=Fastest, 6=Default, 9=Best] (default 6): "))
	lvlStr, _ := reader.ReadString('\n')
	lvlStr = strings.TrimSpace(lvlStr)
	level := 6
	if lvlStr != "" {
		if val, err := strconv.Atoi(lvlStr); err == nil && val >= 0 && val <= 9 {
			level = val
		}
	}

	args := []string{"pack", src, "-l", strconv.Itoa(level), "-f"}
	if dest != "" {
		args = append(args, "-o", dest)
	}

	RootCmd.SetArgs(args)
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
	}
}

func interactiveUnpack(reader *bufio.Reader) {
	fmt.Println()
	ui.PrintSection("Interactive Unpack (Decompression)")

	fmt.Print(ui.Bold(" Enter zip archive path to extract: "))
	archive, _ := reader.ReadString('\n')
	archive = strings.TrimSpace(archive)
	if archive == "" {
		ui.PrintError("Archive path cannot be empty.")
		return
	}

	fmt.Print(ui.Bold(" Enter destination directory (leave blank for default): "))
	dest, _ := reader.ReadString('\n')
	dest = strings.TrimSpace(dest)

	args := []string{"unpack", archive, "-f"}
	if dest != "" {
		args = append(args, "-o", dest)
	}

	RootCmd.SetArgs(args)
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
	}
}

func interactiveList(reader *bufio.Reader) {
	fmt.Println()
	ui.PrintSection("Archive Inspector")

	fmt.Print(ui.Bold(" Enter zip archive path: "))
	archive, _ := reader.ReadString('\n')
	archive = strings.TrimSpace(archive)
	if archive == "" {
		ui.PrintError("Archive path cannot be empty.")
		return
	}

	engine := extract.New()
	summary, err := engine.InspectArchive(archive)
	if err != nil {
		ui.PrintError(err.Error())
		return
	}

	fmt.Println()
	ui.RenderArchiveList(summary)
}

func interactiveBench(reader *bufio.Reader) {
	fmt.Println()
	ui.PrintSection("Interactive Speed Benchmark")

	fmt.Print(ui.Bold(" Enter directory or file to benchmark: "))
	src, _ := reader.ReadString('\n')
	src = strings.TrimSpace(src)
	if src == "" {
		ui.PrintError("Source path cannot be empty.")
		return
	}

	RootCmd.SetArgs([]string{"bench", src})
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
	}
	_ = compress.New()
}
