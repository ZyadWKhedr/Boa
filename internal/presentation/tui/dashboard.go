package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"compressor/internal/bench"
	"compressor/internal/domain"
	"compressor/internal/infrastructure/classifier"
	"compressor/internal/presentation/cli"
	"compressor/internal/ui"
	"compressor/internal/usecase"
	"compressor/pkg/types"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type menuItem struct {
	Number string
	Name   string
	Desc   string
}

var mainMenu = []menuItem{
	{Number: "1.", Name: "Compress", Desc: "Package files into a zip archive with smart media options"},
	{Number: "2.", Name: "Extract", Desc: "Safely unzip archives with Zip-Slip path defense"},
	{Number: "3.", Name: "Browse Archive", Desc: "Inspect files, sizes, ratios & lossy metadata"},
	{Number: "4.", Name: "Learn", Desc: "Educational guide: how compression algorithms work"},
	{Number: "5.", Name: "More...", Desc: "Benchmarks, level comparison, system status & tools"},
}

var moreMenu = []menuItem{
	{Number: "1.", Name: "Benchmark", Desc: "Measure compression speed, duration & throughput"},
	{Number: "2.", Name: "Compare Levels", Desc: "Side-by-side visual bar charts across compression levels"},
	{Number: "3.", Name: "System Status", Desc: "Runtime health, Go environment & platform info"},
	{Number: "4.", Name: "Uninstall", Desc: "Safely remove Boa binaries & symlinks from system"},
	{Number: "5.", Name: "Back", Desc: "Return to main menu"},
}

// Dashboard orchestrates the interactive fullscreen TUI.
type Dashboard struct {
	PackUC      *usecase.PackArchiveUseCase
	ExtractUC   *usecase.ExtractArchiveUseCase
	InspectUC   *usecase.InspectArchiveUseCase
	BenchmarkUC *usecase.BenchmarkUseCase
	LearnUC     *usecase.LearnUseCase
	Version     string
	inMoreMenu  bool
}

// NewDashboard creates a new Dashboard instance.
func NewDashboard(
	packUC *usecase.PackArchiveUseCase,
	extractUC *usecase.ExtractArchiveUseCase,
	inspectUC *usecase.InspectArchiveUseCase,
	benchUC *usecase.BenchmarkUseCase,
	learnUC *usecase.LearnUseCase,
	version string,
) *Dashboard {
	return &Dashboard{
		PackUC:      packUC,
		ExtractUC:   extractUC,
		InspectUC:   inspectUC,
		BenchmarkUC: benchUC,
		LearnUC:     learnUC,
		Version:     version,
	}
}

// Run launches the interactive terminal dashboard.
func (d *Dashboard) Run() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return
	}

	fmt.Print("\033[?1049h\033[?25l")
	defer func() {
		fmt.Print("\033[?25h\033[?1049l")
	}()

	selectedIndex := 0

	for {
		items := mainMenu
		if d.inMoreMenu {
			items = moreMenu
		}

		if selectedIndex >= len(items) {
			selectedIndex = 0
		}

		d.renderMenu(items, selectedIndex)

		oldState, err := term.MakeRaw(fd)
		if err != nil {
			d.fallbackInteractive()
			return
		}

		key, err := ReadKey()
		_ = term.Restore(fd, oldState)
		if err != nil {
			break
		}

		switch key {
		case "UP", "k", "K":
			selectedIndex = (selectedIndex - 1 + len(items)) % len(items)
		case "DOWN", "j", "J":
			selectedIndex = (selectedIndex + 1) % len(items)
		case "1":
			d.handleAction(0)
		case "2":
			d.handleAction(1)
		case "3":
			d.handleAction(2)
		case "4":
			d.handleAction(3)
		case "5":
			d.handleAction(4)
		case "?":
			if !d.inMoreMenu {
				d.handleAction(3) // Learn
			}
		case "ENTER", "SPACE":
			d.handleAction(selectedIndex)
		case "q", "Q", "ESC", "CTRL_C":
			if d.inMoreMenu {
				d.inMoreMenu = false
				selectedIndex = 0
				continue
			}
			fmt.Print("\033[?25h\033[?1049l")
			fmt.Println(ui.Dim("\n Goodbye! 🐍\n"))
			return
		}
	}
}

func (d *Dashboard) renderMenu(items []menuItem, selected int) {
	fmt.Print("\033[H\033[2J")

	ui.PrintBanner()
	fmt.Println()
	title := "Interactive compression toolkit"
	if d.inMoreMenu {
		title = "Advanced Tools & Configuration"
	}
	fmt.Printf(" %s\n\n", ui.Dim("Version "+d.Version+"  ·  "+title))

	for i, item := range items {
		if i == selected {
			pointer := ui.Bold(ui.Cyan("➤"))
			num := ui.Bold(ui.Cyan(item.Number))
			name := ui.Bold(ui.Cyan(fmt.Sprintf("%-16s", item.Name)))
			desc := ui.Cyan(item.Desc)
			fmt.Printf(" %s %s  %s %s\n", pointer, num, name, desc)
		} else {
			num := ui.Bold(item.Number)
			name := fmt.Sprintf("%-16s", item.Name)
			desc := ui.Dim(item.Desc)
			fmt.Printf("   %s  %s %s\n", num, name, desc)
		}
	}

	fmt.Println()
	if d.inMoreMenu {
		fmt.Println(ui.Dim(" ↑↓ / jk Navigate  |  Enter Confirm  |  1-5 Jump  |  Esc Back to Main Menu"))
	} else {
		fmt.Println(ui.Dim(" ↑↓ / jk Navigate  |  Enter Confirm  |  1-5 Jump  |  ? Learn  |  Q Quit"))
	}
}

func (d *Dashboard) handleAction(index int) {
	if d.inMoreMenu {
		d.handleMoreAction(index)
		return
	}

	fmt.Print("\033[?1049l\033[?25h")
	defer func() {
		fmt.Print("\033[?1049h\033[?25l")
	}()

	reader := bufio.NewReader(os.Stdin)

	switch index {
	case 0:
		d.interactivePack(reader)
	case 1:
		d.interactiveUnpack(reader)
	case 2:
		d.interactiveBrowse(reader)
	case 3:
		d.interactiveLearn(reader)
	case 4:
		d.inMoreMenu = true
	}
}

func (d *Dashboard) handleMoreAction(index int) {
	fmt.Print("\033[?1049l\033[?25h")
	defer func() {
		fmt.Print("\033[?1049h\033[?25l")
	}()

	reader := bufio.NewReader(os.Stdin)

	switch index {
	case 0:
		d.interactiveBench(reader)
	case 1:
		d.interactiveCompare(reader)
	case 2:
		d.interactiveStatus(reader)
	case 3:
		fmt.Print("\033[H\033[2J")
		cli.RunUninstall(false)
		d.waitForEnter()
	case 4:
		d.inMoreMenu = false
	}
}

func (d *Dashboard) interactivePack(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Compress", PickAny)
	if err != nil {
		return
	}

	absSrc, _ := filepath.Abs(src)
	parentDir := filepath.Dir(absSrc)
	baseName := filepath.Base(absSrc)
	defaultName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".zip"
	defaultDest := filepath.Join(parentDir, defaultName)

	scanUC := usecase.NewScanPathUseCase(classifier.NewMagicClassifier())
	estimator := usecase.NewEstimateSavingsUseCase()

	scanRes, err := scanUC.Execute(context.Background(), src, []string{".git*", ".DS_Store", "node_modules"})
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to inspect path: %v. Suggested fix: Check that the folder exists and has read permissions.", err))
		d.waitForEnter()
		return
	}

	wizard := NewWizardModel(scanRes, estimator, d.LearnUC)
	p := tea.NewProgram(wizard)
	finalModel, err := p.Run()
	if err != nil {
		ui.PrintError(fmt.Sprintf("TUI error: %v", err))
		d.waitForEnter()
		return
	}

	resModel, ok := finalModel.(WizardModel)
	if !ok || resModel.Result.Cancelled {
		return
	}

	prefs := resModel.Result.Preferences
	policy := domain.NewLossyEligibilityPolicy()
	planUC := usecase.NewBuildCompressionPlanUseCase(policy, estimator)

	plan, err := planUC.Execute(context.Background(), scanRes, prefs, defaultDest)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Cannot create compression plan: %v", err))
		d.waitForEnter()
		return
	}

	opts := types.PackOptions{
		SourcePaths:      []string{src},
		OutputZipPath:    defaultDest,
		Method:           types.MethodDeflate,
		CompressionLevel: prefs.DefaultLevel,
		ExcludePatterns:  []string{".git*", ".DS_Store", "node_modules"},
		Overwrite:        true,
	}
	if prefs.ArchiveMethod == domain.MethodZstd {
		opts.Method = types.MethodZstd
	}

	summary, err := d.PackUC.ExecutePlan(context.Background(), plan, opts)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Compression error: %v. Suggested fix: Ensure destination disk has sufficient free space.", err))
	} else {
		ui.RenderCompressionSummary(summary, false)
	}

	d.waitForEnter()
}

func (d *Dashboard) interactiveUnpack(reader *bufio.Reader) {
	archive, err := PickPath("Select Zip Archive to Extract", PickZipOnly)
	if err != nil {
		return
	}

	absArc, _ := filepath.Abs(archive)
	parentDir := filepath.Dir(absArc)
	baseName := strings.TrimSuffix(filepath.Base(absArc), filepath.Ext(absArc))
	defaultDest := filepath.Join(parentDir, baseName)

	dest := defaultDest

	fmt.Print("\033[H\033[2J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection("Extract Archive")
	fmt.Printf("   %-16s %s\n", ui.Dim("Source Archive"), ui.Bold(ui.Cyan(ui.PrettyPath(archive))))
	fmt.Printf("   %-16s %s\n\n", ui.Dim("Destination"), ui.Bold(ui.Cyan(ui.PrettyPath(defaultDest))))

	fmt.Print(ui.Dim(" Press Enter to extract (or 'c' to customize destination, 'q' to cancel): "))
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

	opts := types.UnpackOptions{
		ArchivePath:    archive,
		DestinationDir: dest,
		Overwrite:      true,
	}

	summary, err := d.ExtractUC.Execute(context.Background(), opts)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Extraction failed: %v. Suggested fix: Verify the file is a valid zip archive.", err))
	} else {
		ui.RenderExtractionSummary(summary, dest, false)
	}

	d.waitForEnter()
}

func (d *Dashboard) interactiveBrowse(reader *bufio.Reader) {
	archive, err := PickPath("Select Zip Archive to Browse", PickZipOnly)
	if err != nil {
		return
	}

	fmt.Print("\033[H\033[2J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection(fmt.Sprintf("Archive Browser: %s", ui.PrettyPath(archive)))
	fmt.Println()

	summary, err := d.InspectUC.Execute(context.Background(), archive)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Could not open archive: %v. Suggested fix: Check file permissions.", err))
	} else {
		ui.RenderArchiveList(summary, true)
	}

	d.waitForEnter()
}

func (d *Dashboard) interactiveBench(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Benchmark", PickAny)
	if err != nil {
		return
	}

	fmt.Print("\033[H\033[2J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintInfo(fmt.Sprintf("Benchmarking target: %s", ui.Bold(ui.PrettyPath(src))))
	fmt.Println(ui.Dim(" Analyzing compression speed, ratio & throughput across levels..."))
	fmt.Println()

	opts := bench.Options{
		SourcePath:    src,
		Level:         -1,
		Runs:          1,
		CompareVisual: false,
	}

	report, err := d.BenchmarkUC.Execute(context.Background(), opts)
	if err != nil {
		ui.PrintError(err.Error())
	} else {
		bench.RenderText(report, false, os.Stdout)
	}

	d.waitForEnter()
}

func (d *Dashboard) interactiveCompare(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Compare", PickAny)
	if err != nil {
		return
	}

	fmt.Print("\033[H\033[2J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintInfo(fmt.Sprintf("Comparing compression levels for: %s", ui.Bold(ui.PrettyPath(src))))
	fmt.Println(ui.Dim(" Generating side-by-side trade-off matrices & visual comparison charts..."))
	fmt.Println()

	opts := bench.Options{
		SourcePath:    src,
		Level:         -1,
		Runs:          2,
		CompareVisual: true,
	}

	report, err := d.BenchmarkUC.Execute(context.Background(), opts)
	if err != nil {
		ui.PrintError(err.Error())
	} else {
		bench.RenderText(report, true, os.Stdout)
	}

	d.waitForEnter()
}

func (d *Dashboard) interactiveLearn(reader *bufio.Reader) {
	fmt.Print("\033[H\033[2J")
	ui.PrintBanner()
	fmt.Println()

	ui.PrintSection("Boa Compression Knowledge Base 📚")
	techniques := d.LearnUC.ListTechniques()

	var lossless []domain.TechniqueInfo
	var lossy []domain.TechniqueInfo

	for _, t := range techniques {
		if t.Category == domain.CategoryLossy {
			lossy = append(lossy, t)
		} else {
			lossless = append(lossless, t)
		}
	}

	fmt.Println(ui.Bold(ui.Green("▶ Lossless Techniques (Bit-for-bit exact)")))
	for _, t := range lossless {
		fmt.Printf("   %-16s %s\n", ui.Bold(ui.Cyan(t.ID)), t.Name)
	}

	fmt.Println()
	fmt.Println(ui.Bold(ui.Yellow("▶ Lossy Media Techniques (Perceptual size optimization)")))
	for _, t := range lossy {
		fmt.Printf("   %-16s %s\n", ui.Bold(ui.Yellow(t.ID)), t.Name)
	}

	fmt.Println()
	fmt.Print(ui.Bold(" Enter technique ID to read full explanation (or Enter to go back): "))
	chosenID, _ := reader.ReadString('\n')
	chosenID = strings.TrimSpace(chosenID)

	if chosenID != "" {
		info, err := d.LearnUC.GetTechniqueInfo(chosenID)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Unknown technique %q. Type 'boa learn' to view available IDs.", chosenID))
		} else {
			fmt.Println()
			fmt.Printf(" %s %s\n", ui.BadgeInfo("LEARN"), ui.Bold(ui.Cyan(info.Name)))
			fmt.Printf(" %s\n\n", ui.Dim(info.Summary))
			fmt.Printf(" %s %s\n\n", ui.Bold("💡 Analogy:"), info.Analogy)
			fmt.Printf(" %s %s\n", ui.Bold(ui.Green("✔ Gain:")), info.WhatYouGain)
			fmt.Printf(" %s %s\n\n", ui.Bold(ui.Red("✖ Loss:")), info.WhatYouLose)
			fmt.Printf(" %s %s\n", ui.Bold("🎯 Best:"), info.BestFor)
			fmt.Printf(" %s %s\n\n", ui.Bold("⚠️ Avoid:"), info.AvoidFor)
			if info.GoDeeper != "" {
				fmt.Printf(" %s\n   %s\n\n", ui.Bold(ui.Magenta("🔬 Go Deeper:")), ui.Dim(info.GoDeeper))
			}
		}
	}

	d.waitForEnter()
}

func (d *Dashboard) interactiveStatus(reader *bufio.Reader) {
	fmt.Print("\033[H\033[2J")
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection("System & Runtime Status")
	fmt.Printf("   %-16s %s\n", ui.Dim("Version"), ui.Bold(ui.Cyan(d.Version)))
	fmt.Println()
	d.waitForEnter()
}

func (d *Dashboard) waitForEnter() {
	fmt.Println()
	fmt.Print(ui.Dim(" Press Enter or 'q' to return to menu... "))
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err == nil {
			defer term.Restore(fd, oldState)
			_, _ = ReadKey()
			return
		}
	}
	buf := bufio.NewReader(os.Stdin)
	_, _ = buf.ReadString('\n')
}

func (d *Dashboard) fallbackInteractive() {
	reader := bufio.NewReader(os.Stdin)
	items := mainMenu
	if d.inMoreMenu {
		items = moreMenu
	}
	for {
		ui.PrintBanner()
		fmt.Println()
		for _, item := range items {
			fmt.Printf("  %s %-16s %s\n", item.Number, item.Name, item.Desc)
		}
		fmt.Print("\n Select option [1-5, Q]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.ToLower(strings.TrimSpace(choice))
		if choice == "q" || choice == "exit" {
			return
		}
		if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(items) {
			d.handleAction(idx - 1)
		}
	}
}
