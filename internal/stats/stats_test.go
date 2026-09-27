package stats

import (
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
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

func TestCalculateSpaceSaved(t *testing.T) {
	saved, pct := CalculateSpaceSaved(1000, 400)
	if saved != 600 {
		t.Errorf("saved = %d, want 600", saved)
	}
	if pct != 60.0 {
		t.Errorf("pct = %f, want 60.0", pct)
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
}
