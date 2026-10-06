package updater

import (
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		expected  bool
	}{
		{"v0.4.0", "v0.3.2", true},
		{"v1.0.0", "v0.9.9", true},
		{"v0.3.3", "v0.3.2", true},
		{"v0.3.2", "v0.3.2", false},
		{"v0.3.1", "v0.3.2", false},
		{"v0.3.0", "v0.3.2", false},
		{"0.4.0", "0.3.2", true},
		{"v0.3.2-1-gd546887", "v0.3.2", false},
		{"v0.4.0-alpha", "v0.3.2", true},
	}

	for _, tt := range tests {
		got := IsNewer(tt.candidate, tt.current)
		if got != tt.expected {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.candidate, tt.current, got, tt.expected)
		}
	}
}

func TestParseSemver(t *testing.T) {
	parsed := parseSemver("v1.2.3")
	if parsed != [3]int{1, 2, 3} {
		t.Errorf("parseSemver(v1.2.3) = %v, want [1, 2, 3]", parsed)
	}

	parsedWithSuffix := parseSemver("v0.3.2-dev-1234")
	if parsedWithSuffix != [3]int{0, 3, 2} {
		t.Errorf("parseSemver(v0.3.2-dev-1234) = %v, want [0, 3, 2]", parsedWithSuffix)
	}
}
