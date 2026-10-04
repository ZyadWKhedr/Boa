package tui

import (
	"fmt"
	"strings"

	"compressor/internal/domain"
	"compressor/internal/stats"
	"compressor/internal/usecase"
	tea "github.com/charmbracelet/bubbletea"
)

type wizardStep int

const (
	stepOverview wizardStep = iota
	stepImages
	stepAudio
	stepVideo
	stepArchivePreset
	stepConsent
)

// WizardResult holds the final user choices and execution decision.
type WizardResult struct {
	Cancelled   bool
	Preferences domain.UserPreferences
}

// WizardModel implements the interactive multi-step compression wizard in Bubble Tea.
type WizardModel struct {
	scan       *domain.ScanResult
	estimator  *usecase.EstimateSavingsUseCase
	learnUC    *usecase.LearnUseCase
	prefs      domain.UserPreferences

	currentStep wizardStep
	activeRow   int // For screens with multiple rows (e.g. Mode vs Quality slider)

	// Modal help state
	showingHelp bool
	helpInfo    domain.TechniqueInfo

	// Steps sequence
	applicableSteps []wizardStep
	stepIndex       int

	// Final result
	Result WizardResult
	Done   bool
}

// NewWizardModel initializes the wizard state.
func NewWizardModel(
	scan *domain.ScanResult,
	estimator *usecase.EstimateSavingsUseCase,
	learnUC *usecase.LearnUseCase,
) WizardModel {
	prefs := domain.UserPreferences{
		ArchiveMethod: domain.MethodDeflate,
		DefaultLevel:  6,
		Images: domain.CategoryPreferences{
			EnableLossy: false,
			Quality:     80,
		},
		Audio: domain.CategoryPreferences{
			EnableLossy: false,
			Quality:     64,
		},
		Video: domain.CategoryPreferences{
			EnableLossy: false,
			Quality:     70,
		},
		StripMetadata: false,
	}

	steps := []wizardStep{stepOverview}
	if scan.Images.Count > 0 {
		steps = append(steps, stepImages)
	}
	if scan.Audio.Count > 0 {
		steps = append(steps, stepAudio)
	}
	if scan.Video.Count > 0 {
		steps = append(steps, stepVideo)
	}
	steps = append(steps, stepArchivePreset)

	return WizardModel{
		scan:            scan,
		estimator:       estimator,
		learnUC:         learnUC,
		prefs:           prefs,
		applicableSteps: steps,
		stepIndex:       0,
		currentStep:     steps[0],
		activeRow:       0,
	}
}

func (m WizardModel) Init() tea.Cmd {
	return nil
}

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 1. Handle Help Modal Overlay
		if m.showingHelp {
			switch msg.String() {
			case "esc", "enter", "q", "?", "space":
				m.showingHelp = false
				return m, nil
			}
			return m, nil
		}

		// 2. Global Key Bindings
		switch msg.String() {
		case "ctrl+c", "q":
			m.Result = WizardResult{Cancelled: true}
			m.Done = true
			return m, tea.Quit

		case "?":
			m.openContextualHelp()
			return m, nil

		case "up", "k":
			if m.activeRow > 0 {
				m.activeRow--
			}
			return m, nil

		case "down", "j":
			if m.maxRowsForCurrentStep() > 1 && m.activeRow < m.maxRowsForCurrentStep()-1 {
				m.activeRow++
			}
			return m, nil

		case "left", "h":
			m.handleLeft()
			return m, nil

		case "right", "l":
			m.handleRight()
			return m, nil

		case "enter", "space":
			return m.handleEnter()

		case "esc", "backspace":
			return m.handleBack()
		}
	}

	return m, nil
}

func (m *WizardModel) maxRowsForCurrentStep() int {
	switch m.currentStep {
	case stepImages:
		if m.prefs.Images.EnableLossy {
			return 2
		}
		return 1
	case stepAudio:
		if m.prefs.Audio.EnableLossy {
			return 2
		}
		return 1
	case stepVideo:
		if m.prefs.Video.EnableLossy {
			return 2
		}
		return 1
	case stepArchivePreset:
		return 2
	case stepConsent:
		return 1
	default:
		return 1
	}
}

func (m *WizardModel) handleLeft() {
	switch m.currentStep {
	case stepImages:
		if m.activeRow == 0 {
			m.prefs.Images.EnableLossy = false
		} else {
			m.prefs.Images.Quality = max(10, m.prefs.Images.Quality-5)
		}
	case stepAudio:
		if m.activeRow == 0 {
			m.prefs.Audio.EnableLossy = false
		} else {
			m.prefs.Audio.Quality = max(10, m.prefs.Audio.Quality-5)
		}
	case stepVideo:
		if m.activeRow == 0 {
			m.prefs.Video.EnableLossy = false
		} else {
			m.prefs.Video.Quality = max(10, m.prefs.Video.Quality-5)
		}
	case stepArchivePreset:
		if m.activeRow == 0 {
			// Preset: Fastest (level 1) <- Balanced (level 6) <- Smallest (level 9)
			if m.prefs.DefaultLevel == 9 {
				m.prefs.DefaultLevel = 6
			} else if m.prefs.DefaultLevel == 6 {
				m.prefs.DefaultLevel = 1
			}
		} else {
			m.prefs.StripMetadata = !m.prefs.StripMetadata
		}
	case stepConsent:
		// Toggle between Confirm and Go Back
		m.activeRow = (m.activeRow + 1) % 2
	}
}

func (m *WizardModel) handleRight() {
	switch m.currentStep {
	case stepImages:
		if m.activeRow == 0 {
			m.prefs.Images.EnableLossy = true
		} else {
			m.prefs.Images.Quality = min(100, m.prefs.Images.Quality+5)
		}
	case stepAudio:
		if m.activeRow == 0 {
			m.prefs.Audio.EnableLossy = true
		} else {
			m.prefs.Audio.Quality = min(100, m.prefs.Audio.Quality+5)
		}
	case stepVideo:
		if m.activeRow == 0 {
			m.prefs.Video.EnableLossy = true
		} else {
			m.prefs.Video.Quality = min(100, m.prefs.Video.Quality+5)
		}
	case stepArchivePreset:
		if m.activeRow == 0 {
			if m.prefs.DefaultLevel == 1 {
				m.prefs.DefaultLevel = 6
			} else if m.prefs.DefaultLevel == 6 {
				m.prefs.DefaultLevel = 9
			}
		} else {
			m.prefs.StripMetadata = !m.prefs.StripMetadata
		}
	case stepConsent:
		m.activeRow = (m.activeRow + 1) % 2
	}
}

func (m WizardModel) handleEnter() (tea.Model, tea.Cmd) {
	if m.currentStep == stepConsent {
		if m.activeRow == 1 { // User chose "Go Back"
			m.stepIndex--
			m.currentStep = m.applicableSteps[m.stepIndex]
			m.activeRow = 0
			return m, nil
		}
		// Confirmed
		m.Result = WizardResult{Cancelled: false, Preferences: m.prefs}
		m.Done = true
		return m, tea.Quit
	}

	// Advance to next step
	if m.stepIndex < len(m.applicableSteps)-1 {
		m.stepIndex++
		m.currentStep = m.applicableSteps[m.stepIndex]
		m.activeRow = 0
		return m, nil
	}

	// If we finished the regular steps, check if we must show the Consent Screen
	if m.hasAnyLossy() && m.currentStep != stepConsent {
		m.currentStep = stepConsent
		m.activeRow = 0 // Default to Confirm
		return m, nil
	}

	// All done without lossy
	m.Result = WizardResult{Cancelled: false, Preferences: m.prefs}
	m.Done = true
	return m, tea.Quit
}

func (m WizardModel) handleBack() (tea.Model, tea.Cmd) {
	if m.currentStep == stepConsent {
		m.currentStep = m.applicableSteps[len(m.applicableSteps)-1]
		m.activeRow = 0
		return m, nil
	}

	if m.stepIndex > 0 {
		m.stepIndex--
		m.currentStep = m.applicableSteps[m.stepIndex]
		m.activeRow = 0
		return m, nil
	}

	// At first step, Esc cancels
	m.Result = WizardResult{Cancelled: true}
	m.Done = true
	return m, tea.Quit
}

func (m *WizardModel) hasAnyLossy() bool {
	return m.prefs.Images.EnableLossy || m.prefs.Audio.EnableLossy || m.prefs.Video.EnableLossy
}

func (m *WizardModel) openContextualHelp() {
	if m.learnUC == nil {
		return
	}

	techniqueID := "deflate"
	switch m.currentStep {
	case stepOverview:
		techniqueID = "deflate"
	case stepImages:
		if m.prefs.Images.EnableLossy {
			techniqueID = "jpeg-quality"
		} else {
			techniqueID = "opt-png"
		}
	case stepAudio:
		if m.prefs.Audio.EnableLossy {
			techniqueID = "audio-bitrate"
		} else {
			techniqueID = "flac"
		}
	case stepVideo:
		if m.prefs.Video.EnableLossy {
			techniqueID = "video-quality"
		} else {
			techniqueID = "store-media"
		}
	case stepArchivePreset:
		if m.prefs.DefaultLevel == 0 {
			techniqueID = "store"
		} else if m.prefs.DefaultLevel == 1 {
			techniqueID = "deflate"
		} else if m.prefs.DefaultLevel == 9 {
			techniqueID = "deflate"
		}
		if m.activeRow == 1 {
			techniqueID = "strip-meta"
		}
	case stepConsent:
		techniqueID = "jpeg-quality"
	}

	info, err := m.learnUC.GetTechniqueInfo(techniqueID)
	if err == nil {
		m.helpInfo = info
		m.showingHelp = true
	}
}

// View renders the TUI screen.
func (m WizardModel) View() string {
	if m.showingHelp {
		return m.renderHelpModal()
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(" " + titleStyle.Render("Boa Smart Archiving Wizard 🐍") + "\n\n")

	switch m.currentStep {
	case stepOverview:
		sb.WriteString(m.renderOverview())
	case stepImages:
		sb.WriteString(m.renderCategoryScreen("Images", m.scan.Images, m.prefs.Images, "jpeg-quality"))
	case stepAudio:
		sb.WriteString(m.renderCategoryScreen("Audio", m.scan.Audio, m.prefs.Audio, "audio-bitrate"))
	case stepVideo:
		sb.WriteString(m.renderCategoryScreen("Video", m.scan.Video, m.prefs.Video, "video-quality"))
	case stepArchivePreset:
		sb.WriteString(m.renderArchivePresetScreen())
	case stepConsent:
		sb.WriteString(m.renderConsentScreen())
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderFooter())
	return sb.String()
}

func (m WizardModel) renderOverview() string {
	var sb strings.Builder
	sb.WriteString(cardStyle.Render(fmt.Sprintf(
		"Target: %s (%s in %d files)\n\n"+
			"Discovered content breakdown:\n"+
			"  • Images:    %d files (%s)  [%s]\n"+
			"  • Audio:     %d files (%s)  [%s]\n"+
			"  • Video:     %d files (%s)  [%s]\n"+
			"  • Other:     %d files (%s)  (Documents, code, databases - 100%% Lossless)",
		m.scan.RootPath, stats.FormatBytes(m.scan.TotalBytes), m.scan.TotalFiles,
		m.scan.Images.Count, stats.FormatBytes(m.scan.Images.TotalBytes), strings.Join(m.scan.Images.Formats, ", "),
		m.scan.Audio.Count, stats.FormatBytes(m.scan.Audio.TotalBytes), strings.Join(m.scan.Audio.Formats, ", "),
		m.scan.Video.Count, stats.FormatBytes(m.scan.Video.TotalBytes), strings.Join(m.scan.Video.Formats, ", "),
		m.scan.Other.Count, stats.FormatBytes(m.scan.Other.TotalBytes),
	)))
	return sb.String()
}

func (m WizardModel) renderCategoryScreen(title string, summary domain.CategorySummary, prefs domain.CategoryPreferences, _ string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(" %s Settings (%d files, %s)\n", titleStyle.Render(title), summary.Count, stats.FormatBytes(summary.TotalBytes)))
	if len(summary.Formats) > 0 {
		sb.WriteString(fmt.Sprintf(" %s\n\n", keyHintStyle.Render("Detected formats: "+strings.Join(summary.Formats, ", "))))
	}

	// Row 0: Mode selection (Volume control style)
	optLossless := unselectedOptionStyle.Render("  Keep original quality (lossless)  ")
	optLossy := unselectedOptionStyle.Render("  Smaller files (some quality loss)  ")

	if !prefs.EnableLossy {
		optLossless = selectedOptionStyle.Render("✔ Keep original quality (lossless)")
	} else {
		optLossy = selectedOptionStyle.Render("⚡ Smaller files (some quality loss)")
	}

	cursor0 := "  "
	if m.activeRow == 0 {
		cursor0 = "➤ "
	}
	sb.WriteString(fmt.Sprintf("%sMode:  %s  %s\n\n", cursor0, optLossless, optLossy))

	// Row 1: Quality slider (if lossy selected)
	if prefs.EnableLossy {
		cursor1 := "  "
		if m.activeRow == 1 {
			cursor1 = "➤ "
		}
		slider := RenderSlider(prefs.Quality, 100, 24)
		sb.WriteString(fmt.Sprintf("%sQuality: %s  (← / → adjusts)\n\n", cursor1, slider))
	} else {
		sb.WriteString(fmt.Sprintf("   %s\n\n", keyHintStyle.Render("Files will be stored with 100% exact fidelity (lossless).")))
	}

	// Live savings estimation
	if m.estimator != nil {
		est := m.estimator.Estimate(m.scan, m.prefs)
		sb.WriteString(cardStyle.Render(fmt.Sprintf(
			"📊 Live Estimated Archive Size: %s – %s  (avg saving: ~%.1f%%)",
			stats.FormatBytes(est.EstimatedMin), stats.FormatBytes(est.EstimatedMax), est.PercentSavedAvg,
		)))
	}

	return sb.String()
}

func (m WizardModel) renderArchivePresetScreen() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(" %s\n\n", titleStyle.Render("Archive Preset & Options")))

	// Row 0: Speed/Density Preset
	optFast := unselectedOptionStyle.Render(" Fastest ")
	optBal := unselectedOptionStyle.Render(" Balanced (Default) ")
	optSmall := unselectedOptionStyle.Render(" Smallest ")

	if m.prefs.DefaultLevel == 1 {
		optFast = selectedOptionStyle.Render("✔ Fastest")
	} else if m.prefs.DefaultLevel == 9 {
		optSmall = selectedOptionStyle.Render("✔ Smallest")
	} else {
		optBal = selectedOptionStyle.Render("✔ Balanced (Default)")
	}

	cursor0 := "  "
	if m.activeRow == 0 {
		cursor0 = "➤ "
	}
	sb.WriteString(fmt.Sprintf("%sCompression Preset:  %s  %s  %s\n\n", cursor0, optFast, optBal, optSmall))

	// Row 1: Strip Metadata toggle
	optMetaKeep := unselectedOptionStyle.Render(" Keep metadata (EXIF/tags) ")
	optMetaStrip := unselectedOptionStyle.Render(" Strip non-essential tags ")

	if !m.prefs.StripMetadata {
		optMetaKeep = selectedOptionStyle.Render("✔ Keep metadata (EXIF/tags)")
	} else {
		optMetaStrip = selectedOptionStyle.Render("⚡ Strip non-essential tags")
	}

	cursor1 := "  "
	if m.activeRow == 1 {
		cursor1 = "➤ "
	}
	sb.WriteString(fmt.Sprintf("%sPrivacy & Metadata: %s  %s\n\n", cursor1, optMetaKeep, optMetaStrip))

	return sb.String()
}

func (m WizardModel) renderConsentScreen() string {
	var sb strings.Builder
	warningHeader := badgeDanger.Render("⚠️ EXPLICIT QUALITY LOSS CONSENT REQUIRED")
	sb.WriteString(fmt.Sprintf(" %s\n\n", warningHeader))

	var affected []string
	if m.prefs.Images.EnableLossy {
		affected = append(affected, fmt.Sprintf("• Images (%d files, %s) at %d%% quality",
			m.scan.Images.Count, stats.FormatBytes(m.scan.Images.TotalBytes), m.prefs.Images.Quality))
	}
	if m.prefs.Audio.EnableLossy {
		affected = append(affected, fmt.Sprintf("• Audio (%d files, %s) at %d%% bitrate quality",
			m.scan.Audio.Count, stats.FormatBytes(m.scan.Audio.TotalBytes), m.prefs.Audio.Quality))
	}
	if m.prefs.Video.EnableLossy {
		affected = append(affected, fmt.Sprintf("• Video (%d files, %s) at %d%% CRF quality",
			m.scan.Video.Count, stats.FormatBytes(m.scan.Video.TotalBytes), m.prefs.Video.Quality))
	}

	notice := fmt.Sprintf(
		"You have chosen to compress media files lossy to save space.\n"+
			"The following files will undergo quality reduction:\n\n"+
			"  %s\n\n"+
			"Note: Original files on disk will NEVER be modified.\n"+
			"All documents, code, and text files remain 100%% byte-for-byte lossless.",
		strings.Join(affected, "\n  "),
	)
	sb.WriteString(cardStyle.Render(notice))
	sb.WriteString("\n\n")

	optConfirm := unselectedOptionStyle.Render(" Confirm & Compress ")
	optBack := unselectedOptionStyle.Render(" Go Back ")

	if m.activeRow == 0 {
		optConfirm = selectedOptionStyle.Render("✔ Confirm & Compress")
	} else {
		optBack = selectedOptionStyle.Render("← Go Back")
	}

	sb.WriteString(fmt.Sprintf(" Action:  %s    %s\n", optConfirm, optBack))
	return sb.String()
}

func (m WizardModel) renderHelpModal() string {
	var sb strings.Builder
	info := m.helpInfo

	catBadge := badgeLossless.Render("LOSSLESS")
	if info.Category == domain.CategoryLossy {
		catBadge = badgeLossy.Render("LOSSY MEDIA")
	}

	body := fmt.Sprintf(
		"%s  %s\n\n"+
			"%s\n\n"+
			"💡 Analogy: %s\n\n"+
			"✔ Gain:  %s\n"+
			"✖ Loss:  %s\n\n"+
			"🎯 Best For:  %s\n"+
			"⚠️ Avoid For: %s\n\n"+
			"Press Esc, Enter, or ? to return to wizard",
		catBadge, titleStyle.Render(info.Name),
		info.Summary,
		info.Analogy,
		info.WhatYouGain,
		info.WhatYouLose,
		info.BestFor,
		info.AvoidFor,
	)

	sb.WriteString("\n")
	sb.WriteString(modalStyle.Render(body))
	return sb.String()
}

func (m WizardModel) renderFooter() string {
	return keyHintStyle.Render(" ←/→ Change option  |  ↑/↓ Move row  |  Enter Next  |  Esc Back  |  ? Help  |  q Cancel")
}
