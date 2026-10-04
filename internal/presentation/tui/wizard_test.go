package tui

import (
	"testing"

	"compressor/internal/domain"
	"compressor/internal/infrastructure/learn"
	"compressor/internal/usecase"
	tea "github.com/charmbracelet/bubbletea"
)

func TestWizardModelWorkflow(t *testing.T) {
	scan := &domain.ScanResult{
		RootPath:   "/test/media",
		TotalFiles: 4,
		TotalBytes: 50 * 1024 * 1024,
		Images: domain.CategorySummary{
			Count:      2,
			TotalBytes: 10 * 1024 * 1024,
			Formats:    []string{"jpg", "png"},
		},
		Audio: domain.CategorySummary{
			Count:      1,
			TotalBytes: 5 * 1024 * 1024,
			Formats:    []string{"mp3"},
		},
		Video: domain.CategorySummary{
			Count:      1,
			TotalBytes: 35 * 1024 * 1024,
			Formats:    []string{"mp4"},
		},
	}

	catalog, err := learn.NewEmbeddedTechniqueCatalog()
	if err != nil {
		t.Fatalf("Failed to load catalog: %v", err)
	}
	learnUC := usecase.NewLearnUseCase(catalog)
	estimator := usecase.NewEstimateSavingsUseCase()

	model := NewWizardModel(scan, estimator, learnUC)

	// 1. Initial Step: Overview
	if model.currentStep != stepOverview {
		t.Errorf("Expected stepOverview, got %v", model.currentStep)
	}

	// 2. Press Enter -> StepImages
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(WizardModel)
	if m.currentStep != stepImages {
		t.Errorf("Expected stepImages, got %v", m.currentStep)
	}

	// 3. Toggle Lossy on Images (Right arrow)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(WizardModel)
	if !m.prefs.Images.EnableLossy {
		t.Errorf("Expected Images.EnableLossy = true after Right arrow")
	}

	// 4. Move to Slider row (Down arrow)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(WizardModel)
	if m.activeRow != 1 {
		t.Errorf("Expected activeRow = 1, got %d", m.activeRow)
	}

	// 5. Adjust slider (Right arrow -> +5)
	prevQuality := m.prefs.Images.Quality
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(WizardModel)
	if m.prefs.Images.Quality != prevQuality+5 {
		t.Errorf("Expected quality %d, got %d", prevQuality+5, m.prefs.Images.Quality)
	}

	// 6. Test contextual help ('?')
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(WizardModel)
	if !m.showingHelp {
		t.Errorf("Expected showingHelp = true after '?' key")
	}
	if m.helpInfo.ID != "jpeg-quality" {
		t.Errorf("Expected helpInfo for 'jpeg-quality', got %q", m.helpInfo.ID)
	}

	// Close help with Esc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(WizardModel)
	if m.showingHelp {
		t.Errorf("Expected showingHelp = false after Esc")
	}

	// 7. Advance through remaining steps
	// Enter -> Audio
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(WizardModel)
	if m.currentStep != stepAudio {
		t.Errorf("Expected stepAudio, got %v", m.currentStep)
	}

	// Enter -> Video
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(WizardModel)
	if m.currentStep != stepVideo {
		t.Errorf("Expected stepVideo, got %v", m.currentStep)
	}

	// Enter -> Archive Preset
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(WizardModel)
	if m.currentStep != stepArchivePreset {
		t.Errorf("Expected stepArchivePreset, got %v", m.currentStep)
	}

	// Enter -> Should show Consent Screen because Images.EnableLossy == true
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(WizardModel)
	if m.currentStep != stepConsent {
		t.Errorf("Expected stepConsent, got %v", m.currentStep)
	}

	// Enter -> Confirm Consent & Finish
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(WizardModel)
	if !m.Done || m.Result.Cancelled {
		t.Errorf("Expected wizard completion with non-cancelled result")
	}
	if cmd == nil {
		t.Errorf("Expected Quit command upon completion")
	}
}
