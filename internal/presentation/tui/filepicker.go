package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"compressor/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilePickerMode defines filtering modes for the file explorer.
type FilePickerMode int

const (
	PickAny FilePickerMode = iota
	PickFolderOnly
	PickZipOnly
)

// FileItem represents an entry in the file browser.
type FileItem struct {
	Name     string
	Path     string
	IsDir    bool
	IsParent bool
	Size     int64
}

// FilePickerModel is a Bubble Tea model for interactive file browsing.
type FilePickerModel struct {
	title        string
	mode         FilePickerMode
	currentDir   string
	items        []FileItem
	selectedIdx  int
	width        int
	height       int
	SelectedPath string
	Cancelled    bool
	Done         bool
	err          error
}

// NewFilePickerModel creates a new FilePickerModel instance.
func NewFilePickerModel(title string, mode FilePickerMode) FilePickerModel {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	absDir, _ := filepath.Abs(dir)

	m := FilePickerModel{
		title:      title,
		mode:       mode,
		currentDir: absDir,
		width:      80,
		height:     24,
	}
	m.loadItems()
	return m
}

func (m *FilePickerModel) loadItems() {
	items, err := readDirItems(m.currentDir, m.mode)
	if err != nil {
		m.err = err
		m.items = nil
		return
	}
	m.err = nil
	m.items = items
	if m.selectedIdx >= len(items) {
		m.selectedIdx = max(0, len(items)-1)
	}
}

// Init initializes the Bubble Tea model.
func (m FilePickerModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming terminal messages and user interactions.
func (m FilePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.Cancelled = true
			m.Done = true
			return m, tea.Quit

		case "up", "k":
			if len(m.items) > 0 {
				m.selectedIdx = (m.selectedIdx - 1 + len(m.items)) % len(m.items)
			}
			return m, nil

		case "down", "j":
			if len(m.items) > 0 {
				m.selectedIdx = (m.selectedIdx + 1) % len(m.items)
			}
			return m, nil

		case "left", "h", "backspace":
			parent := filepath.Dir(m.currentDir)
			if parent != m.currentDir {
				m.currentDir = parent
				m.selectedIdx = 0
				m.loadItems()
			}
			return m, nil

		case "right", "l", "tab":
			if len(m.items) > 0 {
				chosen := m.items[m.selectedIdx]
				if chosen.IsParent {
					m.currentDir = filepath.Dir(m.currentDir)
					m.selectedIdx = 0
					m.loadItems()
				} else if chosen.IsDir && chosen.Path != m.currentDir {
					m.currentDir = chosen.Path
					m.selectedIdx = 0
					m.loadItems()
				}
			}
			return m, nil

		case " ":
			if m.mode == PickFolderOnly || m.mode == PickAny {
				m.SelectedPath = m.currentDir
				m.Done = true
				return m, tea.Quit
			}

		case "enter":
			if len(m.items) == 0 {
				return m, nil
			}
			chosen := m.items[m.selectedIdx]
			if chosen.IsParent {
				m.currentDir = filepath.Dir(m.currentDir)
				m.selectedIdx = 0
				m.loadItems()
				return m, nil
			}
			if chosen.IsDir {
				if m.mode == PickZipOnly {
					m.currentDir = chosen.Path
					m.selectedIdx = 0
					m.loadItems()
					return m, nil
				}
				// PickAny or PickFolderOnly: select folder
				m.SelectedPath = chosen.Path
				m.Done = true
				return m, tea.Quit
			}

			// File selected
			m.SelectedPath = chosen.Path
			m.Done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the interactive file browser using Lip Gloss.
func (m FilePickerModel) View() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	pathStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	muted := lipgloss.NewStyle().Foreground(mutedColor)

	b.WriteString("\n " + headerStyle.Render("📁 "+m.title) + "\n")
	b.WriteString(" " + muted.Render("Current Directory: ") + pathStyle.Render(ui.PrettyPath(m.currentDir)) + "\n\n")

	maxVisible := 14
	startIdx := 0
	if m.selectedIdx >= maxVisible {
		startIdx = m.selectedIdx - maxVisible + 1
	}
	endIdx := min(len(m.items), startIdx+maxVisible)

	if len(m.items) == 0 {
		b.WriteString("   " + muted.Render("(empty directory or no matching files)") + "\n")
	} else {
		for i := startIdx; i < endIdx; i++ {
			item := m.items[i]
			icon := "  "
			if item.IsDir || item.IsParent {
				icon = "📁"
			}

			sizeStr := ""
			if !item.IsDir && !item.IsParent {
				sizeStr = " " + muted.Render(fmt.Sprintf("(%s)", formatBytes(item.Size)))
			}

			if i == m.selectedIdx {
				cursor := lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("➤ ")
				itemText := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#0284C7")).Render(fmt.Sprintf(" %s %s%s ", icon, item.Name, sizeStr))
				b.WriteString(" " + cursor + itemText + "\n")
			} else {
				itemText := fmt.Sprintf("   %s %s%s", icon, item.Name, sizeStr)
				if item.IsDir {
					b.WriteString(lipgloss.NewStyle().Bold(true).Render(itemText) + "\n")
				} else {
					b.WriteString(itemText + "\n")
				}
			}
		}
	}

	b.WriteString("\n")
	if m.mode == PickFolderOnly || m.mode == PickAny {
		b.WriteString(" " + muted.Render("↑↓/jk Navigate  ·  →/l Open Dir  ·  Enter Select  ·  Space Current Dir  ·  Esc Cancel") + "\n")
	} else {
		b.WriteString(" " + muted.Render("↑↓/jk Navigate  ·  →/l Open Dir  ·  Enter Select Archive  ·  Esc Cancel") + "\n")
	}

	return b.String()
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// PickPath launches an interactive in-terminal File Explorer powered by Bubble Tea.
func PickPath(title string, mode FilePickerMode) (string, error) {
	model := NewFilePickerModel(title, mode)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	res, ok := finalModel.(FilePickerModel)
	if !ok || res.Cancelled || res.SelectedPath == "" {
		return "", fmt.Errorf("selection cancelled")
	}

	return res.SelectedPath, nil
}

func readDirItems(dir string, mode FilePickerMode) ([]FileItem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var items []FileItem

	absDir, _ := filepath.Abs(dir)
	if parent := filepath.Dir(absDir); parent != absDir {
		items = append(items, FileItem{
			Name:     ".. (parent directory)",
			Path:     parent,
			IsDir:    true,
			IsParent: true,
		})
	}

	var dirs []FileItem
	var files []FileItem

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && name != ".git" {
			continue
		}

		fullPath := filepath.Join(dir, name)
		info, err := e.Info()
		var size int64
		if err == nil {
			size = info.Size()
		}

		if e.IsDir() {
			dirs = append(dirs, FileItem{
				Name:  name + "/",
				Path:  fullPath,
				IsDir: true,
				Size:  size,
			})
		} else {
			if mode == PickFolderOnly {
				continue
			}
			if mode == PickZipOnly && !strings.HasSuffix(strings.ToLower(name), ".zip") {
				continue
			}
			files = append(files, FileItem{
				Name:  name,
				Path:  fullPath,
				IsDir: false,
				Size:  size,
			})
		}
	}

	sort.Slice(dirs, func(i, j int) bool { return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name) })
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name) })

	items = append(items, dirs...)
	items = append(items, files...)
	return items, nil
}
