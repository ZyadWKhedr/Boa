package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"compressor/internal/bench"
	"compressor/internal/domain"
	"compressor/internal/infrastructure/classifier"
	"compressor/internal/presentation/cli"
	"compressor/internal/ui"
	"compressor/internal/usecase"
	"compressor/pkg/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// MenuAction represents the user selection from the menu.
type MenuAction int

const (
	ActionNone MenuAction = iota
	ActionCompress
	ActionExtract
	ActionBrowse
	ActionLearn
	ActionBenchmark
	ActionCompare
	ActionStatus
	ActionUninstall
	ActionQuit
)

// DashboardModel is the Bubble Tea model for the main interactive menu.
type DashboardModel struct {
	version     string
	inMoreMenu  bool
	selectedIdx int
	width       int
	height      int
	Action      MenuAction
	Quitting    bool
}

// NewDashboardModel initializes the Bubble Tea dashboard model.
func NewDashboardModel(version string, inMoreMenu bool) DashboardModel {
	return DashboardModel{
		version:    version,
		inMoreMenu: inMoreMenu,
		width:      80,
		height:     24,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := mainMenu
	if m.inMoreMenu {
		items = moreMenu
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.inMoreMenu {
				m.inMoreMenu = false
				m.selectedIdx = 0
				return m, nil
			}
			m.Quitting = true
			m.Action = ActionQuit
			return m, tea.Quit

		case "esc":
			if m.inMoreMenu {
				m.inMoreMenu = false
				m.selectedIdx = 0
				return m, nil
			}
			m.Quitting = true
			m.Action = ActionQuit
			return m, tea.Quit

		case "up", "k":
			m.selectedIdx = (m.selectedIdx - 1 + len(items)) % len(items)
			return m, nil

		case "down", "j":
			m.selectedIdx = (m.selectedIdx + 1) % len(items)
			return m, nil

		case "1":
			return m.selectIndex(0)
		case "2":
			return m.selectIndex(1)
		case "3":
			return m.selectIndex(2)
		case "4":
			return m.selectIndex(3)
		case "5":
			return m.selectIndex(4)

		case "?":
			if !m.inMoreMenu {
				m.Action = ActionLearn
				return m, tea.Quit
			}

		case "enter", " ":
			return m.selectIndex(m.selectedIdx)
		}
	}

	return m, nil
}

func (m DashboardModel) selectIndex(idx int) (tea.Model, tea.Cmd) {
	if m.inMoreMenu {
		switch idx {
		case 0:
			m.Action = ActionBenchmark
			return m, tea.Quit
		case 1:
			m.Action = ActionCompare
			return m, tea.Quit
		case 2:
			m.Action = ActionStatus
			return m, tea.Quit
		case 3:
			m.Action = ActionUninstall
			return m, tea.Quit
		case 4:
			m.inMoreMenu = false
			m.selectedIdx = 0
			return m, nil
		}
	} else {
		switch idx {
		case 0:
			m.Action = ActionCompress
			return m, tea.Quit
		case 1:
			m.Action = ActionExtract
			return m, tea.Quit
		case 2:
			m.Action = ActionBrowse
			return m, tea.Quit
		case 3:
			m.Action = ActionLearn
			return m, tea.Quit
		case 4:
			m.inMoreMenu = true
			m.selectedIdx = 0
			return m, nil
		}
	}
	return m, nil
}

func (m DashboardModel) View() string {
	if m.Quitting {
		return lipgloss.NewStyle().Foreground(mutedColor).Render("\n Goodbye! 🐍\n")
	}

	var b strings.Builder

	bannerStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	subtitleStyle := lipgloss.NewStyle().Foreground(mutedColor)

	b.WriteString("\n " + bannerStyle.Render("🐍 Boa Interactive Compression Dashboard") + "\n")

	title := "Interactive compression toolkit"
	if m.inMoreMenu {
		title = "Advanced Tools & Configuration"
	}
	b.WriteString(" " + subtitleStyle.Render("Version "+m.version+"  ·  "+title) + "\n\n")

	items := mainMenu
	if m.inMoreMenu {
		items = moreMenu
	}

	for i, item := range items {
		if i == m.selectedIdx {
			cursor := lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("➤")
			num := lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render(item.Number)
			name := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#0284C7")).Render(fmt.Sprintf(" %-16s ", item.Name))
			desc := lipgloss.NewStyle().Foreground(accentColor).Render(item.Desc)
			b.WriteString(fmt.Sprintf(" %s %s %s %s\n", cursor, num, name, desc))
		} else {
			num := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(item.Number)
			name := fmt.Sprintf("%-18s", item.Name)
			desc := lipgloss.NewStyle().Foreground(mutedColor).Render(item.Desc)
			b.WriteString(fmt.Sprintf("   %s %s %s\n", num, name, desc))
		}
	}

	b.WriteString("\n")
	if m.inMoreMenu {
		b.WriteString(" " + subtitleStyle.Render("↑↓/jk Navigate  ·  Enter Select  ·  1-5 Jump  ·  Esc/q Back to Main Menu") + "\n")
	} else {
		b.WriteString(" " + subtitleStyle.Render("↑↓/jk Navigate  ·  Enter Select  ·  1-5 Jump  ·  ? Learn  ·  q Quit") + "\n")
	}

	return b.String()
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

// Run launches the interactive terminal dashboard powered by Bubble Tea.
func (d *Dashboard) Run() {
	reader := bufio.NewReader(os.Stdin)

	for {
		model := NewDashboardModel(d.Version, d.inMoreMenu)
		p := tea.NewProgram(model, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return
		}

		m, ok := finalModel.(DashboardModel)
		if !ok || m.Action == ActionQuit || m.Quitting {
			fmt.Println(ui.Dim("\n Goodbye! 🐍\n"))
			return
		}

		d.inMoreMenu = m.inMoreMenu

		switch m.Action {
		case ActionCompress:
			d.interactivePack(reader)
		case ActionExtract:
			d.interactiveUnpack(reader)
		case ActionBrowse:
			d.interactiveBrowse(reader)
		case ActionLearn:
			d.interactiveLearn(reader)
		case ActionBenchmark:
			d.interactiveBench(reader)
		case ActionCompare:
			d.interactiveCompare(reader)
		case ActionStatus:
			d.interactiveStatus(reader)
		case ActionUninstall:
			cli.RunUninstall(false)
			d.waitForEnter(reader)
		}
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
		d.waitForEnter(reader)
		return
	}

	wizard := NewWizardModel(scanRes, estimator, d.LearnUC)
	p := tea.NewProgram(wizard, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		ui.PrintError(fmt.Sprintf("TUI error: %v", err))
		d.waitForEnter(reader)
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
		d.waitForEnter(reader)
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

	d.waitForEnter(reader)
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

	d.waitForEnter(reader)
}

func (d *Dashboard) interactiveBrowse(reader *bufio.Reader) {
	archive, err := PickPath("Select Zip Archive to Browse", PickZipOnly)
	if err != nil {
		return
	}

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

	d.waitForEnter(reader)
}

func (d *Dashboard) interactiveBench(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Benchmark", PickAny)
	if err != nil {
		return
	}

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

	d.waitForEnter(reader)
}

func (d *Dashboard) interactiveCompare(reader *bufio.Reader) {
	src, err := PickPath("Select Folder or File to Compare", PickAny)
	if err != nil {
		return
	}

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

	d.waitForEnter(reader)
}

func (d *Dashboard) interactiveLearn(reader *bufio.Reader) {
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

	d.waitForEnter(reader)
}

func (d *Dashboard) interactiveStatus(reader *bufio.Reader) {
	ui.PrintBanner()
	fmt.Println()
	ui.PrintSection("System & Runtime Status")
	fmt.Printf("   %-16s %s\n", ui.Dim("Version"), ui.Bold(ui.Cyan(d.Version)))
	fmt.Println()
	d.waitForEnter(reader)
}

func (d *Dashboard) waitForEnter(reader *bufio.Reader) {
	fmt.Println()
	fmt.Print(ui.Dim(" Press Enter to return to menu... "))
	_, _ = reader.ReadString('\n')
}
