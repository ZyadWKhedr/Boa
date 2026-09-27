package types

import (
	"fmt"
	"strings"
)

// CompressionMethod defines supported compression algorithms within the ZIP container.
type CompressionMethod string

const (
	// MethodDeflate is standard DEFLATE compression (Method ID 8, RFC 1951). Universal compatibility.
	MethodDeflate CompressionMethod = "deflate"

	// MethodStore is raw uncompressed packaging (Method ID 0). Disk I/O bound throughput.
	MethodStore CompressionMethod = "store"

	// MethodZstd is modern Zstandard compression (Method ID 93, PKZIP spec v6.3.8+). High throughput and density.
	MethodZstd CompressionMethod = "zstd"
)

// PKZIP compression method IDs according to APPNOTE.TXT.
const (
	ZipMethodStore   uint16 = 0
	ZipMethodDeflate uint16 = 8
	ZipMethodZstd    uint16 = 93
)

// ParseMethod parses and normalizes a method string into a typed CompressionMethod.
func ParseMethod(s string) (CompressionMethod, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "deflate", "zip":
		return MethodDeflate, nil
	case "store", "none", "copy", "0":
		return MethodStore, nil
	case "zstd", "zstandard":
		return MethodZstd, nil
	default:
		return "", fmt.Errorf("unsupported compression method %q (supported: deflate, store, zstd)", s)
	}
}

// MethodID returns the official PKZIP uint16 method ID for a given CompressionMethod.
func (m CompressionMethod) MethodID() uint16 {
	switch m {
	case MethodStore:
		return ZipMethodStore
	case MethodZstd:
		return ZipMethodZstd
	default:
		return ZipMethodDeflate
	}
}

// DisplayName returns a clean, capitalized display name.
func (m CompressionMethod) DisplayName() string {
	switch m {
	case MethodStore:
		return "Store"
	case MethodZstd:
		return "Zstandard"
	default:
		return "DEFLATE"
	}
}

// MethodNameFromID resolves a raw ZIP method ID to its human-readable algorithm name.
func MethodNameFromID(methodID uint16) string {
	switch methodID {
	case ZipMethodStore:
		return "Store"
	case ZipMethodDeflate:
		return "DEFLATE"
	case ZipMethodZstd:
		return "Zstandard"
	default:
		return fmt.Sprintf("Method(%d)", methodID)
	}
}
