package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDashboardModelNavigation(t *testing.T) {
	model := NewDashboardModel("v1.0.0", false)

	// Test Initial state
	if model.selectedIdx != 0 {
		t.Errorf("Expected initial selectedIdx 0, got %d", model.selectedIdx)
	}
	if model.inMoreMenu {
		t.Errorf("Expected inMoreMenu to be false")
	}

	// Navigate down
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(DashboardModel)
	if m.selectedIdx != 1 {
		t.Errorf("Expected selectedIdx 1, got %d", m.selectedIdx)
	}

	// Press '5' to open More menu
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	m = updated.(DashboardModel)
	if !m.inMoreMenu {
		t.Errorf("Expected inMoreMenu true after pressing 5")
	}
	if m.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx reset to 0 in submenu")
	}

	// Press Esc to return to main menu
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(DashboardModel)
	if m.inMoreMenu {
		t.Errorf("Expected inMoreMenu false after pressing Esc")
	}

	// Press '1' to trigger Compress
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = updated.(DashboardModel)
	if m.Action != ActionCompress {
		t.Errorf("Expected ActionCompress, got %v", m.Action)
	}
	if cmd == nil {
		t.Errorf("Expected tea.Quit command")
	}
}

func TestFilePickerModelNavigation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boa_tui_picker_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "subfolder")
	_ = os.MkdirAll(subDir, 0755)
	testZip := filepath.Join(tempDir, "archive.zip")
	_ = os.WriteFile(testZip, []byte("PK\x05\x06"), 0644)
	testTxt := filepath.Join(tempDir, "document.txt")
	_ = os.WriteFile(testTxt, []byte("hello"), 0644)

	picker := NewFilePickerModel("Test Picker", PickAny)
	picker.currentDir = tempDir
	picker.loadItems()

	if len(picker.items) == 0 {
		t.Fatalf("Expected items in test dir, got 0")
	}

	// Up/Down Navigation
	updated, _ := picker.Update(tea.KeyMsg{Type: tea.KeyDown})
	p := updated.(FilePickerModel)
	if p.selectedIdx != 1 {
		t.Errorf("Expected selectedIdx 1, got %d", p.selectedIdx)
	}

	// Test Esc cancellation
	updated, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	p = updated.(FilePickerModel)
	if !p.Cancelled || !p.Done {
		t.Errorf("Expected picker to be cancelled and done on Esc")
	}
	if cmd == nil {
		t.Errorf("Expected tea.Quit command on cancellation")
	}
}
