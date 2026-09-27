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
