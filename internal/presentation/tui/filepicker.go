package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"compressor/internal/ui"
	"golang.org/x/term"
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

// PickPath launches an interactive in-terminal File Explorer.
func PickPath(title string, mode FilePickerMode) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("terminal not attached")
	}

	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	selectedIdx := 0

	for {
		items, err := readDirItems(currentDir, mode)
		if err != nil {
			return "", err
		}

		if selectedIdx >= len(items) {
			selectedIdx = max(0, len(items)-1)
		}

		renderFilePicker(title, currentDir, items, selectedIdx, mode)

		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return "", err
		}

		key, err := ReadKey()
		_ = term.Restore(fd, oldState)
		if err != nil {
			return "", err
		}

		switch key {
		case "UP", "k", "K":
			if len(items) > 0 {
				selectedIdx = (selectedIdx - 1 + len(items)) % len(items)
			}
		case "DOWN", "j", "J":
			if len(items) > 0 {
				selectedIdx = (selectedIdx + 1) % len(items)
			}
		case "LEFT", "h", "H", "BACKSPACE":
			parent := filepath.Dir(currentDir)
			if parent != currentDir {
				currentDir = parent
				selectedIdx = 0
			}
		case "RIGHT", "TAB", "l", "L", "o", "O":
			if len(items) > 0 {
				chosen := items[selectedIdx]
				if chosen.IsParent {
					currentDir = filepath.Dir(currentDir)
					selectedIdx = 0
				} else if chosen.IsDir && chosen.Path != currentDir {
					currentDir = chosen.Path
					selectedIdx = 0
				}
			}
		case "ENTER":
			if len(items) == 0 {
				continue
			}
			chosen := items[selectedIdx]
			if chosen.IsParent {
				currentDir = filepath.Dir(currentDir)
				selectedIdx = 0
				continue
			}
			if chosen.IsDir && mode == PickZipOnly {
				currentDir = chosen.Path
				selectedIdx = 0
				continue
			}
			return chosen.Path, nil
		case "SPACE":
			if mode == PickFolderOnly || mode == PickAny {
				return currentDir, nil
			}
		case "ESC", "q", "Q", "CTRL_C":
			return "", fmt.Errorf("selection cancelled")
		}
	}
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

func renderFilePicker(title, currentDir string, items []FileItem, selectedIdx int, mode FilePickerMode) {
	fmt.Print("\033[H\033[2J")

	ui.PrintBanner()
	fmt.Println()

	ui.PrintSection(title)
	fmt.Printf("   %s %s\n\n", ui.Dim("Current Directory:"), ui.Bold(ui.Cyan(ui.PrettyPath(currentDir))))

	maxVisible := 14
	startIdx := 0
	if selectedIdx >= maxVisible {
		startIdx = selectedIdx - maxVisible + 1
	}
	endIdx := min(len(items), startIdx+maxVisible)

	for i := startIdx; i < endIdx; i++ {
		item := items[i]
		icon := "📄"
		if item.IsDir {
			icon = "📁"
		}
		if item.IsParent {
			icon = "⤴ "
		}

		if i == selectedIdx {
			cursor := ui.Bold(ui.Cyan("➤"))
			nameStr := ui.Bold(ui.Cyan(fmt.Sprintf("%s %-32s", icon, item.Name)))
			fmt.Printf(" %s %s\n", cursor, nameStr)
		} else {
			nameStr := fmt.Sprintf("   %s %-32s", icon, item.Name)
			if item.IsDir {
				fmt.Printf("%s\n", ui.Bold(nameStr))
			} else {
				fmt.Printf("%s\n", nameStr)
			}
		}
	}

	if len(items) == 0 {
		fmt.Printf("   %s\n", ui.Dim("(empty directory or no matching files)"))
	}

	fmt.Println()
	if mode == PickFolderOnly || mode == PickAny {
		fmt.Println(ui.Dim(" ↑↓ Navigate  |  → / Enter Open  |  Space Select Directory  |  Esc Cancel"))
	} else {
		fmt.Println(ui.Dim(" ↑↓ Navigate  |  → Open  |  Enter Select Zip  |  Esc Cancel"))
	}
}

// ReadKey reads a single keypress or ANSI escape sequence from terminal.
func ReadKey() (string, error) {
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
		case '?':
			return "?", nil
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
