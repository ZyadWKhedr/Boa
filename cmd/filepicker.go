package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"compressor/internal/ui"
	"golang.org/x/term"
)

type FilePickerMode int

const (
	PickAny FilePickerMode = iota
	PickFolderOnly
	PickZipOnly
)

type FileItem struct {
	Name     string
	Path     string
	IsDir    bool
	IsParent bool
	Size     int64
}

// PickPath launches an interactive in-terminal File Explorer.
// title: Header description (e.g. "Select folder or file to compress")
// mode: PickFolderOnly, PickZipOnly, or PickAny
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

		key, err := readKey()
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
			// Go up to parent directory
			parent := filepath.Dir(currentDir)
			if parent != currentDir {
				currentDir = parent
				selectedIdx = 0
			}
		case "RIGHT", "TAB", "l", "L", "o", "O":
			// Step inside folder if a directory is highlighted
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

			// If selecting ".."
			if chosen.IsParent {
				currentDir = filepath.Dir(currentDir)
				selectedIdx = 0
				continue
			}

			// If selecting the top "[SELECT THIS ENTIRE FOLDER]" item
			if chosen.Path == currentDir {
				return currentDir, nil
			}

			// In PickZipOnly mode: opening a directory steps INSIDE it to find zip files
			if mode == PickZipOnly {
				if chosen.IsDir {
					currentDir = chosen.Path
					selectedIdx = 0
					continue
				}
				if strings.HasSuffix(strings.ToLower(chosen.Path), ".zip") {
					return chosen.Path, nil
				}
				continue
			}

			// If a directory is highlighted, Enter steps INSIDE it to explore its contents/files
			if chosen.IsDir {
				currentDir = chosen.Path
				selectedIdx = 0
				continue
			}

			// Return the selected file
			return chosen.Path, nil

		case "SPACE", "s", "S", "c", "C":
			if len(items) == 0 {
				continue
			}
			chosen := items[selectedIdx]
			if chosen.IsParent {
				continue
			}
			if mode == PickZipOnly {
				if strings.HasSuffix(strings.ToLower(chosen.Path), ".zip") {
					return chosen.Path, nil
				}
				continue
			}
			// Space/S/C selects the highlighted folder or file immediately
			return chosen.Path, nil

		case "q", "Q", "ESC", "CTRL_C":
			return "", fmt.Errorf("cancelled")
		}
	}
}

func readDirItems(dir string, mode FilePickerMode) ([]FileItem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var items []FileItem

	// Add "Select current directory" option if picking folders or any items
	if mode == PickFolderOnly || mode == PickAny {
		items = append(items, FileItem{
			Name:  "✔  [PACK THIS ENTIRE FOLDER: " + filepath.Base(dir) + "]",
			Path:  dir,
			IsDir: true,
		})
	}

	// Add parent directory option
	parent := filepath.Dir(dir)
	if parent != dir {
		items = append(items, FileItem{
			Name:     "📁  .. (Parent Directory)",
			Path:     parent,
			IsDir:    true,
			IsParent: true,
		})
	}

	// Separate dirs and files
	var dirs []FileItem
	var files []FileItem

	for _, entry := range entries {
		name := entry.Name()
		// Hide standard hidden files by default
		if strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(dir, name)

		if entry.IsDir() {
			dirs = append(dirs, FileItem{
				Name:  "📁  " + name + "/",
				Path:  fullPath,
				IsDir: true,
			})
		} else {
			if mode == PickZipOnly && !strings.HasSuffix(strings.ToLower(name), ".zip") {
				continue
			}

			icon := "📄  "
			if strings.HasSuffix(strings.ToLower(name), ".zip") {
				icon = "📦  "
			}

			files = append(files, FileItem{
				Name:  icon + name,
				Path:  fullPath,
				IsDir: false,
				Size:  info.Size(),
			})
		}
	}

	// Sort alphabetically
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	if mode == PickFolderOnly {
		items = append(items, dirs...)
	} else {
		items = append(items, dirs...)
		items = append(items, files...)
	}

	return items, nil
}

func renderFilePicker(title, currentDir string, items []FileItem, selectedIdx int, mode FilePickerMode) {
	fmt.Print("\033[H\033[2J") // Clear screen

	ui.PrintBanner()
	fmt.Println()
	fmt.Printf(" %s\n", ui.Bold(ui.Cyan("▶ "+title)))
	fmt.Printf(" %s %s\n\n", ui.Dim("Location:"), ui.Bold(ui.PrettyPath(currentDir)))

	// Render scroll window (up to 12 visible items)
	maxVisible := 12
	startIdx := 0
	if selectedIdx >= maxVisible {
		startIdx = selectedIdx - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(items) {
		endIdx = len(items)
	}

	hasZip := false
	for _, item := range items {
		if !item.IsDir && strings.HasSuffix(strings.ToLower(item.Path), ".zip") {
			hasZip = true
			break
		}
	}

	if mode == PickZipOnly && !hasZip {
		fmt.Printf("   %s\n\n", ui.Dim("✦ No .zip archives here. Enter a folder below or press ← to go up."))
	} else if len(items) == 0 {
		fmt.Printf("   %s\n\n", ui.Dim("(empty directory)"))
	}

	for i := startIdx; i < endIdx; i++ {
		item := items[i]
		if i == selectedIdx {
			pointer := ui.Bold(ui.Cyan("➤"))
			label := ui.Bold(ui.Cyan(item.Name))
			fmt.Printf(" %s %s\n", pointer, label)
		} else {
			fmt.Printf("   %s\n", item.Name)
		}
	}

	fmt.Println()
	fmt.Println(ui.Dim("========================================================================"))
	if mode == PickZipOnly {
		fmt.Println(ui.Dim(" ↑↓ / jk Navigate  |  Enter Open Folder / Select Zip  |  ← Back  |  Q Cancel"))
	} else {
		fmt.Println(ui.Dim(" ↑↓ / jk Navigate  |  Enter Open Folder / Select File  |  Space Select Folder  |  ← Back  |  Q Cancel"))
	}
}
