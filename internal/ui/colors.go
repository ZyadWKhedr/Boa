package ui

import (
	"os"
	"strings"
)

var (
	// NoColor controls whether ANSI color escape codes are emitted.
	NoColor = false
)

func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		NoColor = true
	}
	if term := os.Getenv("TERM"); term == "dumb" {
		NoColor = true
	}
}

// ANSI Escape Codes
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	italic    = "\033[3m"
	underline = "\033[4m"

	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	white     = "\033[37m"
	gray      = "\033[90m"

	bgCyan    = "\033[46;30m"
	bgGreen   = "\033[42;30m"
	bgYellow  = "\033[43;30m"
	bgRed     = "\033[41;37m"
	bgBlue    = "\033[44;37m"
)

func colorize(code, text string) string {
	if NoColor || text == "" {
		return text
	}
	return code + text + reset
}

// Color and formatting helpers
func Bold(s string) string      { return colorize(bold, s) }
func Dim(s string) string       { return colorize(dim, s) }
func Italic(s string) string    { return colorize(italic, s) }
func Red(s string) string       { return colorize(red, s) }
func Green(s string) string     { return colorize(green, s) }
func Yellow(s string) string    { return colorize(yellow, s) }
func Blue(s string) string      { return colorize(blue, s) }
func Magenta(s string) string   { return colorize(magenta, s) }
func Cyan(s string) string      { return colorize(cyan, s) }
func Gray(s string) string      { return colorize(gray, s) }
func White(s string) string     { return colorize(white, s) }

// Badge helpers for dense, clean status reporting
func BadgeOK(text string) string {
	if NoColor {
		return "[" + text + "]"
	}
	return colorize(green+bold, "✔ ") + colorize(bold, text)
}

func BadgeInfo(text string) string {
	if NoColor {
		return "[" + text + "]"
	}
	return colorize(cyan+bold, "ℹ ") + colorize(bold, text)
}

func BadgeWarn(text string) string {
	if NoColor {
		return "[" + text + "]"
	}
	return colorize(yellow+bold, "▲ ") + colorize(yellow+bold, text)
}

func BadgeErr(text string) string {
	if NoColor {
		return "[" + text + "]"
	}
	return colorize(red+bold, "✖ ") + colorize(red+bold, text)
}

func BadgeDryRun(text string) string {
	if NoColor {
		return "[DRY-RUN " + text + "]"
	}
	return colorize(magenta+bold, "✦ DRY-RUN: ") + colorize(bold, text)
}

// StripAnsi removes ANSI escape codes from a string (useful for calculating exact column width).
func StripAnsi(str string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (str[i] >= 'A' && str[i] <= 'Z') || (str[i] >= 'a' && str[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		b.WriteByte(str[i])
	}
	return b.String()
}
