package stats

import (
	"math"
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{-1024, "-1.0 KB"},
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestCalculateRatio(t *testing.T) {
	tests := []struct {
		name       string
		orig       int64
		comp       int64
		expected   float64
		closeCheck bool
	}{
		{"normal", 1000, 400, 0.4, true},
		{"identical", 500, 500, 1.0, true},
		{"zero orig", 0, 500, 1.0, true},
		{"negative orig", -10, 500, 1.0, true},
		{"expansion", 100, 150, 1.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRatio(tt.orig, tt.comp)
			if math.Abs(got-tt.expected) > 0.001 {
				t.Errorf("CalculateRatio(%d, %d) = %f, want %f", tt.orig, tt.comp, got, tt.expected)
			}
		})
	}
}

func TestCalculateSpaceSaved(t *testing.T) {
	tests := []struct {
		name        string
		orig        int64
		comp        int64
		wantSaved   int64
		wantPercent float64
	}{
		{"standard reduction", 1000, 400, 600, 60.0},
		{"identical sizes", 1000, 1000, 0, 0.0},
		{"expansion / negative saved", 1000, 1200, -200, -20.0},
		{"zero byte original", 0, 100, 0, 0.0},
		{"negative original", -50, 50, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved, pct := CalculateSpaceSaved(tt.orig, tt.comp)
			if saved != tt.wantSaved {
				t.Errorf("saved = %d, want %d", saved, tt.wantSaved)
			}
			if math.Abs(pct-tt.wantPercent) > 0.001 {
				t.Errorf("pct = %f, want %f", pct, tt.wantPercent)
			}
		})
	}
}

func TestCalculateSpeed(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		duration time.Duration
		want     string
	}{
		{"1 MB in 1s", 1024 * 1024, 1 * time.Second, "1.0 MB/s"},
		{"0 bytes", 0, 1 * time.Second, "0 B/s"},
		{"0 duration", 1024, 0, "0 B/s"},
		{"negative duration", 1024, -1 * time.Second, "0 B/s"},
		{"500 KB in 500ms", 500 * 1024, 500 * time.Millisecond, "1000.0 KB/s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSpeed(tt.bytes, tt.duration)
			if got != tt.want {
				t.Errorf("CalculateSpeed(%d, %v) = %q, want %q", tt.bytes, tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	if got := FormatDuration(500 * time.Microsecond); got != "500µs" {
		t.Errorf("got %s, want 500µs", got)
	}
	if got := FormatDuration(250 * time.Millisecond); got != "250ms" {
		t.Errorf("got %s, want 250ms", got)
	}
	if got := FormatDuration(1500 * time.Millisecond); got != "1.50s" {
		t.Errorf("got %s, want 1.50s", got)
	}
	if got := FormatDuration(75 * time.Second); got != "1m15s" {
		t.Errorf("got %s, want 1m15s", got)
	}
}
