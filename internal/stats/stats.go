package stats

import (
	"fmt"
	"math"
	"time"
)

// FormatBytes formats a byte count into a compact human-readable string (e.g., 1.4 MB, 512 B).
func FormatBytes(b int64) string {
	if b < 0 {
		return "-" + FormatBytes(-b)
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), units[exp])
}

// FormatDuration formats a duration into a compact clean string (e.g., 45ms, 1.2s, 3m10s).
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%02ds", mins, secs)
}

// CalculateRatio computes the compression ratio (compressed / original).
func CalculateRatio(originalBytes, compressedBytes int64) float64 {
	if originalBytes <= 0 {
		return 1.0
	}
	return float64(compressedBytes) / float64(originalBytes)
}

// CalculateSpaceSaved returns bytes saved and the percentage saved.
func CalculateSpaceSaved(originalBytes, compressedBytes int64) (int64, float64) {
	saved := originalBytes - compressedBytes
	if originalBytes <= 0 {
		return 0, 0.0
	}
	pct := (float64(saved) / float64(originalBytes)) * 100.0
	if math.IsNaN(pct) || math.IsInf(pct, 0) {
		pct = 0.0
	}
	return saved, pct
}

// CalculateSpeed returns human-readable throughput string (e.g. "45.2 MB/s").
func CalculateSpeed(bytes int64, d time.Duration) string {
	if d <= 0 || bytes <= 0 {
		return "0 B/s"
	}
	seconds := d.Seconds()
	bytesPerSec := float64(bytes) / seconds
	return fmt.Sprintf("%s/s", FormatBytes(int64(bytesPerSec)))
}
