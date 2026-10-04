package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color Palette
	primaryColor   = lipgloss.Color("#10B981") // Emerald Green
	accentColor    = lipgloss.Color("#06B6D4") // Cyan
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	dangerColor    = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Slate Muted
	subtleBgColor  = lipgloss.Color("#1E293B") // Dark Slate
	cardBorderColor = lipgloss.Color("#334155")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	badgeLossless = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#065F46")).
			Foreground(lipgloss.Color("#A7F3D0")).
			Padding(0, 1)

	badgeLossy = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#78350F")).
			Foreground(lipgloss.Color("#FDE68A")).
			Padding(0, 1)

	badgeDanger = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#991B1B")).
			Foreground(lipgloss.Color("#FECACA")).
			Padding(0, 1)

	selectedOptionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#0284C7")).
				Padding(0, 1)

	unselectedOptionStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cardBorderColor).
			Padding(1, 2)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(accentColor).
			Padding(1, 2)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

// RenderProgressBar renders a sleek terminal slider bar [########----]
func RenderSlider(value, maxVal, width int) string {
	if maxVal <= 0 {
		maxVal = 100
	}
	if value < 0 {
		value = 0
	}
	if value > maxVal {
		value = maxVal
	}

	filledLen := (value * width) / maxVal
	emptyLen := width - filledLen

	filled := lipgloss.NewStyle().Foreground(accentColor).Render(strings.Repeat("█", filledLen))
	empty := lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("░", emptyLen))

	return fmt.Sprintf("[%s%s] %3d%%", filled, empty, value)
}
