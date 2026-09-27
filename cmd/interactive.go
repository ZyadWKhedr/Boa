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
	"golang.org/x/term"
)

type menuItem struct {
	Number string
	Name   string
	Desc   string
}

var menuItems = []menuItem{
	{Number: "1.", Name: "Pack", Desc: "Compress folders into dense archives"},
	{Number: "2.", Name: "Unpack", Desc: "Safely extract zip archives"},
	{Number: "3.", Name: "Inspect", Desc: "Explore archive structure & metadata"},
	{Number: "4.", Name: "Benchmark", Desc: "Compare compression speed & ratios"},
	{Number: "5.", Name: "Status", Desc: "Runtime health & system info"},
}

// RunInteractiveDashboard launches real-time interactive arrow-key navigation.
func RunInteractiveDashboard() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		_ = RootCmd.Help()
		return
	}

	selectedIndex := 0

	for {
		renderMenu(selectedIndex)

		// Switch stdin to raw mode to read single keypresses immediately
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			// Fallback to line mode if raw mode fails
			fallbackInteractive()
			return
		}

		key, err := readKey()
		_ = term.Restore(fd, oldState)
		if err != nil {
			break
		}

		switch key {
		case "UP", "k", "K":
			selectedIndex = (selectedIndex - 1 + len(menuItems)) % len(menuItems)
		case "DOWN", "j", "J":
			selectedIndex = (selectedIndex + 1) % len(menuItems)
		case "1":
			selectedIndex = 0
			handleAction(0)
		case "2":
			selectedIndex = 1
			handleAction(1)
		case "3":
			selectedIndex = 2
			handleAction(2)
		case "4":
			selectedIndex = 3
			handleAction(3)
		case "5":
			selectedIndex = 4
			handleAction(4)
		case "ENTER", "SPACE":
			handleAction(selectedIndex)
		case "v", "V":
			selectedIndex = 4
			handleAction(4)
		case "h", "H", "?":
			fmt.Print("\033[H\033[2J") // Clear screen
			ui.PrintBanner()
			fmt.Println()
			_ = RootCmd.Help()
			waitForEnter()
		case "q", "Q", "ESC", "CTRL_C":
			fmt.Print("\033[H\033[2J")
			ui.PrintBanner()
			fmt.Println(ui.Dim("\n Goodbye! 🐍\n"))
			return
		}
	}
}

func renderMenu(selected int) {
	// ANSI Clear screen and move cursor to top-left
	fmt.Print("\033[H\033[2J")

	ui.PrintBanner()
	fmt.Println()
	fmt.Printf(" %s\n\n", ui.Dim("Update 1.0.0 available, run bo update"))

	for i, item := range menuItems {
		if i == selected {
			// Highlighted active row with pointer ➤
			pointer := ui.Bold(ui.Cyan("➤"))
			num := ui.Bold(ui.Cyan(item.Number))
			name := ui.Bold(ui.Cyan(fmt.Sprintf("%-12s", item.Name)))
			desc := ui.Cyan(item.Desc)
			fmt.Printf(" %s %s  %s %s\n", pointer, num, name, desc)
		} else {
			// Inactive row
			num := ui.Bold(item.Number)
			name := fmt.Sprintf("%-12s", item.Name)
			desc := ui.Dim(item.Desc)
			fmt.Printf("   %s  %s %s\n", num, name, desc)
		}
	}

	fmt.Println()
	fmt.Println(ui.Dim(" ↑↓ / jk Navigate  |  Enter Confirm  |  1-5 Jump  |  V Version  |  Q Quit"))
}

func handleAction(index int) {
	fmt.Print("\033[H\033[2J")
	ui.PrintBanner()

	reader := bufio.NewReader(os.Stdin)

	switch index {
	case 0:
		interactivePack(reader)
	case 1:
		interactiveUnpack(reader)
	case 2:
		interactiveList(reader)
	case 3:
		interactiveBench(reader)
	case 4:
		fmt.Println()
		RootCmd.SetArgs([]string{"version"})
		_ = RootCmd.Execute()
	}

	waitForEnter()
}

func waitForEnter() {
	fmt.Println()
	fmt.Print(ui.Dim(" Press Enter to return to main menu..."))
	buf := bufio.NewReader(os.Stdin)
	_, _ = buf.ReadString('\n')
}

// readKey reads and decodes a single key or ANSI escape sequence in raw mode.
func readKey() (string, error) {
	var buf [3]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return "", err
	}

	if n == 1 {
		switch buf[0] {
		case 3: // Ctrl+C
			return "CTRL_C", nil
		case 13, 10: // Enter
			return "ENTER", nil
		case 27: // Esc
			return "ESC", nil
		case 32: // Space
			return "SPACE", nil
		default:
			return string(buf[:1]), nil
		}
	}

	if n == 3 && buf[0] == 27 && buf[1] == 91 { // ANSI escape \033[
		switch buf[2] {
		case 'A':
			return "UP", nil
		case 'B':
			return "DOWN", nil
		case 'C':
			return "RIGHT", nil
		case 'D':
			return "LEFT", nil
		}
	}

	return "", nil
}

func fallbackInteractive() {
	reader := bufio.NewReader(os.Stdin)
	for {
		ui.PrintBanner()
		fmt.Println()
		for _, item := range menuItems {
			fmt.Printf("  %s %-12s %s\n", item.Number, item.Name, item.Desc)
		}
		fmt.Print("\n Select option [1-5, Q]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.ToLower(strings.TrimSpace(choice))
		if choice == "q" || choice == "exit" {
			return
		}
		if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= 5 {
			handleAction(idx - 1)
		}
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
