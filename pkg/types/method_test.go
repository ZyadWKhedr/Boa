package types

import (
	"testing"
)

func TestParseMethod(t *testing.T) {
	tests := []struct {
		input     string
		want      CompressionMethod
		expectErr bool
	}{
		{"", MethodDeflate, false},
		{"deflate", MethodDeflate, false},
		{"DEFLATE", MethodDeflate, false},
		{"zip", MethodDeflate, false},
		{"store", MethodStore, false},
		{"STORE", MethodStore, false},
		{"none", MethodStore, false},
		{"0", MethodStore, false},
		{"zstd", MethodZstd, false},
		{"ZSTD", MethodZstd, false},
		{"zstandard", MethodZstd, false},
		{"bzip2", "", true},
		{"invalid_algo", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseMethod(tt.input)
			if (err != nil) != tt.expectErr {
				t.Fatalf("ParseMethod(%q) err = %v, expectErr = %v", tt.input, err, tt.expectErr)
			}
			if got != tt.want {
				t.Errorf("ParseMethod(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMethodIDAndNames(t *testing.T) {
	if MethodStore.MethodID() != ZipMethodStore {
		t.Errorf("expected Store ID %d, got %d", ZipMethodStore, MethodStore.MethodID())
	}
	if MethodDeflate.MethodID() != ZipMethodDeflate {
		t.Errorf("expected Deflate ID %d, got %d", ZipMethodDeflate, MethodDeflate.MethodID())
	}
	if MethodZstd.MethodID() != ZipMethodZstd {
		t.Errorf("expected Zstd ID %d, got %d", ZipMethodZstd, MethodZstd.MethodID())
	}

	if MethodNameFromID(0) != "Store" {
		t.Errorf("MethodNameFromID(0) = %q, want Store", MethodNameFromID(0))
	}
	if MethodNameFromID(8) != "DEFLATE" {
		t.Errorf("MethodNameFromID(8) = %q, want DEFLATE", MethodNameFromID(8))
	}
	if MethodNameFromID(93) != "Zstandard" {
		t.Errorf("MethodNameFromID(93) = %q, want Zstandard", MethodNameFromID(93))
	}
	if MethodNameFromID(999) != "Method(999)" {
		t.Errorf("MethodNameFromID(999) = %q, want Method(999)", MethodNameFromID(999))
	}
}

func TestValidateLevel(t *testing.T) {
	tests := []struct {
		name      string
		method    CompressionMethod
		level     int
		wantLevel int
		expectErr bool
	}{
		{"deflate default (-1)", MethodDeflate, -1, 6, false},
		{"deflate level 1", MethodDeflate, 1, 1, false},
		{"deflate level 9", MethodDeflate, 9, 9, false},
		{"deflate level 0 (store)", MethodDeflate, 0, 0, false},
		{"deflate invalid level 10", MethodDeflate, 10, 0, true},
		{"deflate invalid level -2", MethodDeflate, -2, 0, true},

		{"store mode (-1)", MethodStore, -1, 0, false},
		{"store mode (5)", MethodStore, 5, 0, false},

		{"zstd default (-1)", MethodZstd, -1, 3, false},
		{"zstd level 1", MethodZstd, 1, 1, false},
		{"zstd level 3", MethodZstd, 3, 3, false},
		{"zstd level 11", MethodZstd, 11, 11, false},
		{"zstd invalid level 0", MethodZstd, 0, 0, true},
		{"zstd invalid level 12", MethodZstd, 12, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateLevel(tt.method, tt.level)
			if (err != nil) != tt.expectErr {
				t.Fatalf("ValidateLevel(%v, %d) err = %v, expectErr = %v", tt.method, tt.level, err, tt.expectErr)
			}
			if !tt.expectErr && got != tt.wantLevel {
				t.Errorf("ValidateLevel(%v, %d) = %d, want %d", tt.method, tt.level, got, tt.wantLevel)
			}
		})
	}
}
