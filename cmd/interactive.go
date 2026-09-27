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
		ui.PrintBanner()
		fmt.Println()
		fmt.Printf("  %s\n", ui.Bold("What would you like to do?"))
		fmt.Println()
		fmt.Printf("  %s  %s\n", ui.Bold(ui.Cyan("1")), "📦 Compress a folder or file ("+ui.Dim("pack / zip")+")")
		fmt.Printf("  %s  %s\n", ui.Bold(ui.Cyan("2")), "📂 Extract a zip archive ("+ui.Dim("unpack / unzip")+")")
		fmt.Printf("  %s  %s\n", ui.Bold(ui.Cyan("3")), "🔍 Inspect archive contents ("+ui.Dim("list / ls")+")")
		fmt.Printf("  %s  %s\n", ui.Bold(ui.Cyan("4")), "⚡ Benchmark compression performance ("+ui.Dim("bench")+")")
		fmt.Printf("  %s  %s\n", ui.Bold(ui.Cyan("5")), "📖 View CLI Command Reference ("+ui.Dim("help")+")")
		fmt.Printf("  %s  %s\n", ui.Bold(ui.Dim("0")), "✕ Exit")
		fmt.Println()

		fmt.Print(ui.Bold(ui.Cyan("  Select [1-5, 0]: ")))
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1", "p", "pack", "zip":
			interactivePack(reader)
		case "2", "u", "unpack", "unzip":
			interactiveUnpack(reader)
		case "3", "l", "list", "ls":
			interactiveList(reader)
		case "4", "b", "bench":
			interactiveBench(reader)
		case "5", "h", "help":
			fmt.Println()
			_ = RootCmd.Help()
		case "0", "q", "exit", "quit":
			fmt.Println(ui.Dim("  Goodbye! 👋"))
			return
		default:
			ui.PrintWarning("Invalid selection. Please choose an option from 0 to 5.")
		}

		fmt.Println()
		fmt.Print(ui.Dim("  Press Enter to continue..."))
		_, _ = reader.ReadString('\n')
		fmt.Println()
	}
}

func interactivePack(reader *bufio.Reader) {
	fmt.Println()
	ui.PrintSection("Interactive Pack (Zip Compression)")

	fmt.Print(ui.Bold("  Enter source folder or file path: "))
	src, _ := reader.ReadString('\n')
	src = strings.TrimSpace(src)
	if src == "" {
		ui.PrintError("Source path cannot be empty.")
		return
	}

	fmt.Print(ui.Bold("  Enter destination zip path (leave empty for auto-name): "))
	dest, _ := reader.ReadString('\n')
	dest = strings.TrimSpace(dest)

	fmt.Print(ui.Bold("  Compression Level [0=Store, 1=Fastest, 6=Default, 9=Best] (default 6): "))
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

	fmt.Print(ui.Bold("  Enter zip archive path: "))
	archive, _ := reader.ReadString('\n')
	archive = strings.TrimSpace(archive)
	if archive == "" {
		ui.PrintError("Archive path cannot be empty.")
		return
	}

	fmt.Print(ui.Bold("  Enter target extraction folder (leave empty for auto-name): "))
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
	ui.PrintSection("Interactive Archive Inspector")

	fmt.Print(ui.Bold("  Enter zip archive path to inspect: "))
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
	ui.PrintSection("Interactive Compression Benchmark")

	fmt.Print(ui.Bold("  Enter folder or file path to benchmark: "))
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
