package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	{Number: "1.", Name: "Pack", Desc: "Compress folders or files into dense archives"},
	{Number: "2.", Name: "Unpack", Desc: "Safely extract zip archives with Zip-Slip defense"},
	{Number: "3.", Name: "List", Desc: "List files, sizes & directory contents of an archive"},
	{Number: "4.", Name: "Inspect", Desc: "Explore archive structure, ratios & metadata"},
	{Number: "5.", Name: "Benchmark", Desc: "Measure compression speed, duration & throughput"},
	{Number: "6.", Name: "Compare", Desc: "Side-by-side visual bar charts across compression levels"},
	{Number: "7.", Name: "Status", Desc: "Runtime health, Go environment & platform info"},
	{Number: "8.", Name: "Uninstall", Desc: "Safely remove Boa binaries & symlinks from system"},
}

// RunInteractiveDashboard launches real-time interactive arrow-key navigation.
func RunInteractiveDashboard() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		_ = RootCmd.Help()
		return
	}

	// Enter alternate screen buffer & hide cursor for a crisp fullscreen TUI experience
	fmt.Print("\033[?1049h\033[?25l")
	defer func() {
		// Restore cursor & leave alternate screen buffer on exit
		fmt.Print("\033[?25h\033[?1049l")
	}()

	selectedIndex := 0

	for {
		renderMenu(selectedIndex)

		oldState, err := term.MakeRaw(fd)
		if err != nil {
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
		case "6":
			selectedIndex = 5
			handleAction(5)
		case "7":
			selectedIndex = 6
			handleAction(6)
		case "8", "u", "U":
			selectedIndex = 7
			handleAction(7)
		case "ENTER", "SPACE":
			handleAction(selectedIndex)
		case "v", "V":
			selectedIndex = 6
			handleAction(6)
		case "h", "H", "?":
			fmt.Print("\033[H\033[2J\033[3J")
			ui.PrintBanner()
			fmt.Println()
			_ = RootCmd.Help()
			waitForEnter()
		case "q", "Q", "ESC", "CTRL_C":
			fmt.Print("\033[?25h\033[?1049l")
			fmt.Println(ui.Dim("\n Goodbye! 🐍\n"))
			return
		}
	}
}

func renderMenu(selected int) {
	fmt.Print("\033[H\033[2J\033[3J")

	ui.PrintBanner()
	fmt.Println()
	fmt.Printf(" %s\n\n", ui.Dim("Version "+Version+"  ·  Interactive compression toolkit"))

	for i, item := range menuItems {
		if i == selected {
			pointer := ui.Bold(ui.Cyan("➤"))
			num := ui.Bold(ui.Cyan(item.Number))
			name := ui.Bold(ui.Cyan(fmt.Sprintf("%-12s", item.Name)))
			desc := ui.Cyan(item.Desc)
			fmt.Printf(" %s %s  %s %s\n", pointer, num, name, desc)
		} else {
			num := ui.Bold(item.Number)
			name := fmt.Sprintf("%-12s", item.Name)
			desc := ui.Dim(item.Desc)
			fmt.Printf("   %s  %s %s\n", num, name, desc)
		}
	}

	fmt.Println()
	fmt.Println(ui.Dim(" ↑↓ / jk Navigate  |  Enter Confirm  |  1-8 Jump  |  V Version  |  Q Quit"))
}

func handleAction(index int) {
	fmt.Print("\033[?25h")        // Show cursor during action execution
	defer fmt.Print("\033[?25l") // Hide cursor when returning to dashboard menu

	reader := bufio.NewReader(os.Stdin)

	switch index {
	case 0:
		interactivePack(reader)
	case 1:
		interactiveUnpack(reader)
	case 2:
		interactiveList(reader)
	case 3:
		interactiveInspect(reader)
	case 4:
		interactiveBench(reader)
	case 5:
		interactiveCompare(reader)
	case 6:
		fmt.Print("\033[H\033[2J\033[3J")
		ui.PrintBanner()
		fmt.Println()
		ui.SuppressBanner = true
		RootCmd.SetArgs([]string{"version"})
		_ = RootCmd.Execute()
		ui.SuppressBanner = false
		waitForEnter()
	case 7:
		interactiveUninstall(reader)
	}
}

func waitForEnter() {
	fmt.Println()
	fmt.Print(ui.Dim(" Press Enter or 'q' to return to main menu... "))
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err == nil {
			defer term.Restore(fd, oldState)
			_, _ = readKey()
			return
		}
	}
	buf := bufio.NewReader(os.Stdin)
	_, _ = buf.ReadString('\n')
}

func readKey() (string, error) {
	var buf [3]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return "", err
	}

	if n == 1 {
		switch buf[0] {
		case 3:
			return "CTRL_C", nil
		case 9:
			return "TAB", nil
		case 13, 10:
			return "ENTER", nil
		case 27:
			return "ESC", nil
		case 32:
			return "SPACE", nil
		case 127:
			return "BACKSPACE", nil
		default:
			return string(buf[:1]), nil
		}
	}

	if n == 3 && buf[0] == 27 && buf[1] == 91 {
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
		fmt.Print("\n Select option [1-7, Q]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.ToLower(strings.TrimSpace(choice))
		if choice == "q" || choice == "exit" {
			return
		}
		if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(menuItems) {
			handleAction(idx - 1)
		}
	}
}

func interactivePack(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Compress", PickAny)
	if err != nil {
		return
	}

	// Calculate default zip name in same directory
	absSrc, _ := filepath.Abs(src)
	parentDir := filepath.Dir(absSrc)
	baseName := filepath.Base(absSrc)
	defaultName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".zip"
	defaultDest := filepath.Join(parentDir, defaultName)

	level := 6
	dest := defaultDest

	fmt.Print("\033[H\033[2J\033[3J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection("Pack Configuration")
	fmt.Printf("   %-16s %s\n", ui.Dim("Source"), ui.Bold(ui.Cyan(ui.PrettyPath(src))))
	fmt.Printf("   %-16s %s\n", ui.Dim("Destination"), ui.Bold(ui.Cyan(ui.PrettyPath(defaultDest))))
	fmt.Printf("   %-16s %s\n\n", ui.Dim("Level"), "6 (Default)")

	fmt.Print(ui.Dim(" Press Enter to compress (or 'c' to customize, 'q' to cancel): "))
	optChoice, _ := reader.ReadString('\n')
	optChoice = strings.ToLower(strings.TrimSpace(optChoice))

	if optChoice == "q" || optChoice == "cancel" {
		return
	}

	if optChoice == "c" || optChoice == "custom" {
		fmt.Println()
		fmt.Print(ui.Bold(" Custom destination archive name (leave empty for default): "))
		customDest, _ := reader.ReadString('\n')
		customDest = strings.TrimSpace(customDest)
		if customDest != "" {
			dest = customDest
		}

		fmt.Print(ui.Bold(" Compression level [0=Store, 1=Fast, 6=Default, 9=Best]: "))
		lvlStr, _ := reader.ReadString('\n')
		lvlStr = strings.TrimSpace(lvlStr)
		if lvlStr != "" {
			if val, err := strconv.Atoi(lvlStr); err == nil && val >= 0 && val <= 9 {
				level = val
			}
		}
	}

	args := []string{"pack", src, "-l", strconv.Itoa(level), "-o", dest, "-f"}

	ui.SuppressBanner = true
	RootCmd.SetArgs(args)
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
	}
	ui.SuppressBanner = false

	waitForEnter()
}

func interactiveUnpack(reader *bufio.Reader) {
	archive, err := PickPath("Select Zip Archive to Extract", PickZipOnly)
	if err != nil {
		return
	}

	absArc, _ := filepath.Abs(archive)
	parentDir := filepath.Dir(absArc)
	baseName := strings.TrimSuffix(filepath.Base(absArc), filepath.Ext(absArc))
	defaultDest := filepath.Join(parentDir, baseName)

	dest := defaultDest

	fmt.Print("\033[H\033[2J\033[3J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection("Unpack Configuration")
	fmt.Printf("   %-16s %s\n", ui.Dim("Archive"), ui.Bold(ui.Cyan(ui.PrettyPath(archive))))
	fmt.Printf("   %-16s %s\n\n", ui.Dim("Extract To"), ui.Bold(ui.Cyan(ui.PrettyPath(defaultDest))))

	fmt.Print(ui.Dim(" Press Enter to extract (or 'c' to customize, 'q' to cancel): "))
	optChoice, _ := reader.ReadString('\n')
	optChoice = strings.ToLower(strings.TrimSpace(optChoice))

	if optChoice == "q" || optChoice == "cancel" {
		return
	}

	if optChoice == "c" || optChoice == "custom" {
		fmt.Println()
		fmt.Print(ui.Bold(" Custom output directory (leave empty for default): "))
		customDest, _ := reader.ReadString('\n')
		customDest = strings.TrimSpace(customDest)
		if customDest != "" {
			dest = customDest
		}
	}

	args := []string{"unpack", archive, "-o", dest, "-f"}

	ui.SuppressBanner = true
	RootCmd.SetArgs(args)
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
	}
	ui.SuppressBanner = false

	waitForEnter()
}

func interactiveList(reader *bufio.Reader) {
	archive, err := PickPath("Select Zip Archive to List", PickZipOnly)
	if err != nil {
		return
	}

	fmt.Print("\033[H\033[2J\033[3J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection(fmt.Sprintf("Archive Contents: %s", ui.PrettyPath(archive)))
	fmt.Println()

	engine := extract.New()
	summary, err := engine.InspectArchive(archive)
	if err != nil {
		ui.PrintError(err.Error())
		waitForEnter()
		return
	}

	ui.RenderArchiveList(summary, false)
	waitForEnter()
}

func interactiveInspect(reader *bufio.Reader) {
	archive, err := PickPath("Select Zip Archive to Inspect", PickZipOnly)
	if err != nil {
		return
	}

	fmt.Print("\033[H\033[2J\033[3J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection(fmt.Sprintf("Deep Inspection: %s", ui.PrettyPath(archive)))
	fmt.Println()

	engine := extract.New()
	summary, err := engine.InspectArchive(archive)
	if err != nil {
		ui.PrintError(err.Error())
		waitForEnter()
		return
	}

	ui.RenderArchiveList(summary, false)
	waitForEnter()
}

func interactiveBench(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Benchmark", PickAny)
	if err != nil {
		return
	}

	fmt.Print("\033[H\033[2J\033[3J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintInfo(fmt.Sprintf("Benchmarking target: %s", ui.Bold(ui.PrettyPath(src))))
	fmt.Println(ui.Dim(" Analyzing compression speed, ratio & throughput across levels..."))
	fmt.Println()

	ui.SuppressBanner = true
	RootCmd.SetArgs([]string{"bench", src})
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
	}
	ui.SuppressBanner = false

	waitForEnter()
}

func interactiveCompare(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Compare", PickAny)
	if err != nil {
		return
	}

	fmt.Print("\033[H\033[2J\033[3J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintInfo(fmt.Sprintf("Comparing compression levels for: %s", ui.Bold(ui.PrettyPath(src))))
	fmt.Println(ui.Dim(" Generating side-by-side trade-off matrices & visual comparison charts..."))
	fmt.Println()

	ui.SuppressBanner = true
	RootCmd.SetArgs([]string{"bench", src, "--compare"})
	if err := RootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
	}
	ui.SuppressBanner = false

	waitForEnter()
}

func interactiveUninstall(reader *bufio.Reader) {
	fmt.Print("\033[H\033[2J\033[3J")
	RunUninstall(false)
	waitForEnter()
}

