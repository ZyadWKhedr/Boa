package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"compressor/internal/stats"
	"golang.org/x/term"
)

// ProgressBar displays a sleek, responsive in-terminal progress bar during compression/extraction.
type ProgressBar struct {
	mu           sync.Mutex
	out          io.Writer
	title        string
	totalFiles   int
	totalBytes   int64
	currentFiles int
	currentBytes int64
	currentName  string
	startTime    time.Time
	lastRender   time.Time
	isTerminal   bool
	disabled     bool
	completed    bool
}

// NewProgressBar initializes a new progress bar.
// If totalFiles <= 0, an indeterminate spinner and byte counter is used.
func NewProgressBar(title string, totalFiles int, totalBytes int64) *ProgressBar {
	isTerm := false
	if f, ok := Out.(*os.File); ok {
		isTerm = term.IsTerminal(int(f.Fd()))
	}

	pb := &ProgressBar{
		out:        Out,
		title:      title,
		totalFiles: totalFiles,
		totalBytes: totalBytes,
		startTime:  time.Now(),
		isTerminal: isTerm,
		disabled:   SuppressBanner,
	}

	if !pb.disabled && pb.isTerminal {
		pb.render()
	}

	return pb
}

// Update advances the progress bar state.
func (pb *ProgressBar) Update(fileName string, currentBytes int64, currentFiles int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.disabled || pb.completed || !pb.isTerminal {
		return
	}

	pb.currentName = fileName
	pb.currentBytes = currentBytes
	pb.currentFiles = currentFiles

	// Throttle rendering to at most 30 times per second for maximum performance
	if time.Since(pb.lastRender) > 30*time.Millisecond || pb.currentFiles == pb.totalFiles {
		pb.render()
		pb.lastRender = time.Now()
	}
}

// Finish completes the progress bar and clears the line for the final summary card.
func (pb *ProgressBar) Finish() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.completed {
		return
	}
	pb.completed = true

	if !pb.disabled && pb.isTerminal {
		// Clear the progress line completely
		fmt.Fprint(pb.out, "\r\033[K")
	}
}

// render draws the progress bar to standard output. Must be called with lock held.
func (pb *ProgressBar) render() {
	width := 80
	if f, ok := pb.out.(*os.File); ok {
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 20 {
			width = w
		}
	}

	var percent float64
	if pb.totalFiles > 0 {
		percent = float64(pb.currentFiles) / float64(pb.totalFiles)
	} else if pb.totalBytes > 0 {
		percent = float64(pb.currentBytes) / float64(pb.totalBytes)
	}
	if percent > 1.0 {
		percent = 1.0
	} else if percent < 0 {
		percent = 0
	}

	barWidth := 20
	if width < 70 {
		barWidth = 10
	}

	filled := int(percent * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	unfilled := barWidth - filled

	barStr := strings.Repeat("█", filled) + strings.Repeat("░", unfilled)
	if !NoColor {
		barStr = Cyan(strings.Repeat("█", filled)) + Dim(strings.Repeat("░", unfilled))
	}

	elapsed := time.Since(pb.startTime)
	speedStr := stats.CalculateSpeed(pb.currentBytes, elapsed)
	bytesStr := stats.FormatBytes(pb.currentBytes)

	var prefix string
	var statsStr string

	if pb.totalFiles > 0 {
		prefix = fmt.Sprintf(" %s [%s] %3.0f%%", Bold(pb.title), barStr, percent*100)
		statsStr = fmt.Sprintf("(%d/%d files · %s · %s)", pb.currentFiles, pb.totalFiles, bytesStr, speedStr)
	} else {
		prefix = fmt.Sprintf(" %s [%s]", Bold(pb.title), barStr)
		statsStr = fmt.Sprintf("(%d files · %s · %s)", pb.currentFiles, bytesStr, speedStr)
	}

	// Calculate remaining width for file path
	rawLineLen := len(prefix) + len(statsStr) + 6
	maxFileLen := width - rawLineLen
	if maxFileLen < 10 {
		maxFileLen = 10
	}

	fileName := pb.currentName
	if len(fileName) > maxFileLen {
		fileName = "..." + fileName[len(fileName)-maxFileLen+3:]
	}

	fileDisplay := ""
	if fileName != "" {
		fileDisplay = fmt.Sprintf(" %s %s", Dim("→"), fileName)
	}

	outputLine := fmt.Sprintf("\r\033[K%s %s%s", prefix, statsStr, fileDisplay)

	// Clamp to terminal width
	fmt.Fprint(pb.out, outputLine)
}
