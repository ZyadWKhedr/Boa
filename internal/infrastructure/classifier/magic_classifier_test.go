package classifier

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"compressor/internal/domain"
)

func TestMagicClassifier(t *testing.T) {
	classifier := NewMagicClassifier()

	tests := []struct {
		name         string
		filenameHint string
		content      []byte
		wantKind     domain.MediaKind
		wantLossy    bool
	}{
		{
			name:         "Real JPEG image",
			filenameHint: "photo.jpg",
			content:      []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			wantKind:     domain.KindImage,
			wantLossy:    true,
		},
		{
			name:         "Deceptive file: .txt that is actually JPEG image",
			filenameHint: "secret.txt",
			content:      []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			wantKind:     domain.KindImage,
			wantLossy:    true,
		},
		{
			name:         "Deceptive file: .jpg that is actually Go source code (must NOT be classified as image)",
			filenameHint: "malicious.jpg",
			content:      []byte("package main\n\nfunc main() {}\n"),
			wantKind:     domain.KindOther,
			wantLossy:    false,
		},
		{
			name:         "Real PNG image",
			filenameHint: "graphic.png",
			content:      []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00},
			wantKind:     domain.KindImage,
			wantLossy:    false,
		},
		{
			name:         "Real WAV uncompressed audio",
			filenameHint: "audio.wav",
			content:      []byte("RIFF____WAVEfmt "),
			wantKind:     domain.KindAudio,
			wantLossy:    false,
		},
		{
			name:         "Real FLAC lossless audio",
			filenameHint: "track.flac",
			content:      []byte("fLaC\x00\x00\x00\x22"),
			wantKind:     domain.KindAudio,
			wantLossy:    false,
		},
		{
			name:         "Real MP4 video container",
			filenameHint: "movie.mp4",
			content:      []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32},
			wantKind:     domain.KindVideo,
			wantLossy:    true,
		},
		{
			name:         "PDF document",
			filenameHint: "contract.pdf",
			content:      []byte("%PDF-1.7\n%âãÏÓ\n"),
			wantKind:     domain.KindOther,
			wantLossy:    false,
		},
		{
			name:         "ZIP archive",
			filenameHint: "bundle.zip",
			content:      []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00, 0x00, 0x00},
			wantKind:     domain.KindOther,
			wantLossy:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cls, err := classifier.ClassifyReader(bytes.NewReader(tt.content), tt.filenameHint)
			if err != nil {
				t.Fatalf("ClassifyReader() unexpected error: %v", err)
			}

			if cls.MediaKind != tt.wantKind {
				t.Errorf("MediaKind = %v, want %v", cls.MediaKind, tt.wantKind)
			}
			if cls.IsAlreadyLossy != tt.wantLossy {
				t.Errorf("IsAlreadyLossy = %v, want %v", cls.IsAlreadyLossy, tt.wantLossy)
			}
		})
	}
}

func TestMagicClassifierFileOnDisk(t *testing.T) {
	classifier := NewMagicClassifier()
	tempDir := t.TempDir()

	jpegPath := filepath.Join(tempDir, "photo_ disguised_as_txt.txt")
	_ = os.WriteFile(jpegPath, []byte{0xFF, 0xD8, 0xFF, 0xE1}, 0o644)

	cls, err := classifier.Classify(jpegPath)
	if err != nil {
		t.Fatalf("Classify() error: %v", err)
	}
	if cls.MediaKind != domain.KindImage {
		t.Errorf("Expected KindImage for disguised jpeg, got %v", cls.MediaKind)
	}
}
