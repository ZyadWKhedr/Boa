package classifier

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"compressor/internal/domain"
)

// MagicClassifier implements domain.FileClassifier using magic bytes and container signatures.
type MagicClassifier struct{}

// NewMagicClassifier creates a new instance of MagicClassifier.
func NewMagicClassifier() *MagicClassifier {
	return &MagicClassifier{}
}

// Classify inspects a file on disk by reading its initial header bytes.
func (c *MagicClassifier) Classify(path string) (domain.FileClassification, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return domain.FileClassification{}, err
	}

	if fi.IsDir() {
		return domain.FileClassification{
			MediaKind:   domain.KindOther,
			Description: "Directory",
		}, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return domain.FileClassification{}, err
	}
	defer f.Close()

	return c.ClassifyReader(f, filepath.Base(path))
}

// ClassifyReader inspects header bytes from an io.Reader and uses filenameHint for fallback.
func (c *MagicClassifier) ClassifyReader(r io.Reader, filenameHint string) (domain.FileClassification, error) {
	// Read first 512 bytes for magic number analysis
	header := make([]byte, 512)
	n, err := io.ReadFull(r, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return domain.FileClassification{}, err
	}
	header = header[:n]

	ext := strings.ToLower(filepath.Ext(filenameHint))

	// 1. Check strict Magic Byte Signatures (Content first!)
	if cls, ok := c.matchMagicBytes(header); ok {
		cls.Extension = ext
		return cls, nil
	}

	// 2. Standard library MIME sniff fallback
	sniffed := http.DetectContentType(header)
	if strings.HasPrefix(sniffed, "text/") {
		// Active text content is definitively KindOther (code/documents) regardless of filename extension.
		return domain.FileClassification{
			MIMEType:    sniffed,
			Extension:   ext,
			MediaKind:   domain.KindOther,
			Description: "Text / Source Document",
		}, nil
	}

	if cls, ok := c.matchMIMESniff(sniffed, ext); ok {
		return cls, nil
	}

	// 3. If content is present but does not match any media signature, fail closed to KindOther.
	if len(header) > 0 {
		return domain.FileClassification{
			MIMEType:    sniffed,
			Extension:   ext,
			MediaKind:   domain.KindOther,
			Description: "Unknown / Binary (" + sniffed + ")",
		}, nil
	}

	// 4. Zero-byte empty files or empty streams fallback to extension
	return c.fallbackClassification(ext), nil
}

func (c *MagicClassifier) matchMagicBytes(b []byte) (domain.FileClassification, bool) {
	if len(b) < 4 {
		return domain.FileClassification{}, false
	}

	// JPEG: FF D8 FF
	if len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF {
		return domain.FileClassification{
			MIMEType:            "image/jpeg",
			MediaKind:           domain.KindImage,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "JPEG Image",
		}, true
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return domain.FileClassification{
			MIMEType:            "image/png",
			MediaKind:           domain.KindImage,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      false,
			Description:         "PNG Image (Lossless)",
		}, true
	}

	// GIF: GIF87a or GIF89a
	if len(b) >= 6 && (bytes.Equal(b[:6], []byte("GIF87a")) || bytes.Equal(b[:6], []byte("GIF89a"))) {
		return domain.FileClassification{
			MIMEType:            "image/gif",
			MediaKind:           domain.KindImage,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "GIF Image",
		}, true
	}

	// WebP / RIFF container (RIFF....WEBP)
	if len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")) {
		return domain.FileClassification{
			MIMEType:            "image/webp",
			MediaKind:           domain.KindImage,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "WebP Image",
		}, true
	}

	// WAV Audio (RIFF....WAVE)
	if len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WAVE")) {
		return domain.FileClassification{
			MIMEType:            "audio/wav",
			MediaKind:           domain.KindAudio,
			IsAlreadyCompressed: false,
			IsAlreadyLossy:      false,
			Description:         "WAV Audio (Uncompressed)",
		}, true
	}

	// AVI Video (RIFF....AVI )
	if len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("AVI ")) {
		return domain.FileClassification{
			MIMEType:            "video/x-msvideo",
			MediaKind:           domain.KindVideo,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "AVI Video",
		}, true
	}

	// FLAC: fLaC
	if bytes.Equal(b[:4], []byte("fLaC")) {
		return domain.FileClassification{
			MIMEType:            "audio/flac",
			MediaKind:           domain.KindAudio,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      false,
			Description:         "FLAC Audio (Lossless)",
		}, true
	}

	// MP3 with ID3 tag (ID3) or sync frame FF FB / FF F3 / FF F2
	if bytes.Equal(b[:3], []byte("ID3")) || (b[0] == 0xFF && (b[1]&0xE0) == 0xE0) {
		return domain.FileClassification{
			MIMEType:            "audio/mpeg",
			MediaKind:           domain.KindAudio,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "MP3 Audio",
		}, true
	}

	// OGG container (OggS)
	if bytes.Equal(b[:4], []byte("OggS")) {
		return domain.FileClassification{
			MIMEType:            "audio/ogg",
			MediaKind:           domain.KindAudio,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "Ogg Media Stream",
		}, true
	}

	// MP4 / MOV / M4A ISO Base Media: (....ftyp)
	if len(b) >= 8 && bytes.Equal(b[4:8], []byte("ftyp")) {
		return domain.FileClassification{
			MIMEType:            "video/mp4",
			MediaKind:           domain.KindVideo,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "MP4/ISO Media Container",
		}, true
	}

	// Matroska / MKV / WebM: 1A 45 DF A3
	if len(b) >= 4 && bytes.Equal(b[:4], []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return domain.FileClassification{
			MIMEType:            "video/x-matroska",
			MediaKind:           domain.KindVideo,
			IsAlreadyCompressed: true,
			IsAlreadyLossy:      true,
			Description:         "Matroska / WebM Video",
		}, true
	}

	// PDF: %PDF-
	if len(b) >= 5 && bytes.Equal(b[:5], []byte("%PDF-")) {
		return domain.FileClassification{
			MIMEType:            "application/pdf",
			MediaKind:           domain.KindOther,
			IsAlreadyCompressed: false,
			Description:         "PDF Document",
		}, true
	}

	// ZIP: PK\x03\x04
	if len(b) >= 4 && bytes.Equal(b[:4], []byte{0x50, 0x4B, 0x03, 0x04}) {
		return domain.FileClassification{
			MIMEType:            "application/zip",
			MediaKind:           domain.KindOther,
			IsAlreadyCompressed: true,
			Description:         "ZIP Archive",
		}, true
	}

	return domain.FileClassification{}, false
}

func (c *MagicClassifier) matchMIMESniff(sniffed, ext string) (domain.FileClassification, bool) {
	if strings.HasPrefix(sniffed, "image/") {
		return domain.FileClassification{
			MIMEType:            sniffed,
			Extension:           ext,
			MediaKind:           domain.KindImage,
			IsAlreadyCompressed: true,
			Description:         "Image (" + sniffed + ")",
		}, true
	}
	if strings.HasPrefix(sniffed, "audio/") {
		return domain.FileClassification{
			MIMEType:            sniffed,
			Extension:           ext,
			MediaKind:           domain.KindAudio,
			IsAlreadyCompressed: true,
			Description:         "Audio (" + sniffed + ")",
		}, true
	}
	if strings.HasPrefix(sniffed, "video/") {
		return domain.FileClassification{
			MIMEType:            sniffed,
			Extension:           ext,
			MediaKind:           domain.KindVideo,
			IsAlreadyCompressed: true,
			Description:         "Video (" + sniffed + ")",
		}, true
	}
	return domain.FileClassification{}, false
}

func (c *MagicClassifier) fallbackClassification(ext string) domain.FileClassification {
	switch ext {
	case ".jpg", ".jpeg":
		return domain.FileClassification{MIMEType: "image/jpeg", Extension: ext, MediaKind: domain.KindImage, IsAlreadyCompressed: true, IsAlreadyLossy: true, Description: "JPEG Image"}
	case ".png":
		return domain.FileClassification{MIMEType: "image/png", Extension: ext, MediaKind: domain.KindImage, IsAlreadyCompressed: true, IsAlreadyLossy: false, Description: "PNG Image"}
	case ".webp":
		return domain.FileClassification{MIMEType: "image/webp", Extension: ext, MediaKind: domain.KindImage, IsAlreadyCompressed: true, IsAlreadyLossy: true, Description: "WebP Image"}
	case ".gif":
		return domain.FileClassification{MIMEType: "image/gif", Extension: ext, MediaKind: domain.KindImage, IsAlreadyCompressed: true, IsAlreadyLossy: true, Description: "GIF Image"}
	case ".mp3":
		return domain.FileClassification{MIMEType: "audio/mpeg", Extension: ext, MediaKind: domain.KindAudio, IsAlreadyCompressed: true, IsAlreadyLossy: true, Description: "MP3 Audio"}
	case ".wav":
		return domain.FileClassification{MIMEType: "audio/wav", Extension: ext, MediaKind: domain.KindAudio, IsAlreadyCompressed: false, IsAlreadyLossy: false, Description: "WAV Audio"}
	case ".flac":
		return domain.FileClassification{MIMEType: "audio/flac", Extension: ext, MediaKind: domain.KindAudio, IsAlreadyCompressed: true, IsAlreadyLossy: false, Description: "FLAC Audio"}
	case ".mp4", ".m4v", ".mov":
		return domain.FileClassification{MIMEType: "video/mp4", Extension: ext, MediaKind: domain.KindVideo, IsAlreadyCompressed: true, IsAlreadyLossy: true, Description: "MP4 Video"}
	case ".mkv", ".webm":
		return domain.FileClassification{MIMEType: "video/webm", Extension: ext, MediaKind: domain.KindVideo, IsAlreadyCompressed: true, IsAlreadyLossy: true, Description: "WebM/MKV Video"}
	case ".zip", ".tar", ".gz", ".7z", ".rar":
		return domain.FileClassification{MIMEType: "application/archive", Extension: ext, MediaKind: domain.KindOther, IsAlreadyCompressed: true, Description: "Archive File"}
	default:
		return domain.FileClassification{MIMEType: "application/octet-stream", Extension: ext, MediaKind: domain.KindOther, Description: "Other / Unknown"}
	}
}
