package image

import (
	"bytes"
	"context"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestStandardImageCodec(t *testing.T) {
	codec := NewStandardImageCodec()
	ctx := context.Background()

	// 1. Create a test image
	testImg := CreateTestRGBImage(128, 128)

	// Encode to raw high-quality JPEG
	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, testImg, &jpeg.Options{Quality: 98}); err != nil {
		t.Fatalf("Failed to encode test jpeg: %v", err)
	}
	rawJPEG := jpegBuf.Bytes()

	// Encode to raw PNG
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, testImg); err != nil {
		t.Fatalf("Failed to encode test png: %v", err)
	}
	rawPNG := pngBuf.Bytes()

	t.Run("CanHandle", func(t *testing.T) {
		if !codec.CanHandle("image/jpeg") {
			t.Errorf("Expected CanHandle(image/jpeg) = true")
		}
		if !codec.CanHandle("image/png") {
			t.Errorf("Expected CanHandle(image/png) = true")
		}
		if codec.CanHandle("application/pdf") {
			t.Errorf("Expected CanHandle(application/pdf) = false")
		}
	})

	t.Run("JPEG Lossy Compression at Quality 40 vs 90", func(t *testing.T) {
		var lowQBuf bytes.Buffer
		if err := codec.CompressLossy(ctx, bytes.NewReader(rawJPEG), &lowQBuf, "image/jpeg", 40); err != nil {
			t.Fatalf("CompressLossy(40) failed: %v", err)
		}

		var highQBuf bytes.Buffer
		if err := codec.CompressLossy(ctx, bytes.NewReader(rawJPEG), &highQBuf, "image/jpeg", 90); err != nil {
			t.Fatalf("CompressLossy(90) failed: %v", err)
		}

		if lowQBuf.Len() >= highQBuf.Len() {
			t.Errorf("Expected quality 40 (%d bytes) to be smaller than quality 90 (%d bytes)",
				lowQBuf.Len(), highQBuf.Len())
		}
	})

	t.Run("PNG Palette Quantization at Quality 60", func(t *testing.T) {
		var palettedBuf bytes.Buffer
		if err := codec.CompressLossy(ctx, bytes.NewReader(rawPNG), &palettedBuf, "image/png", 60); err != nil {
			t.Fatalf("CompressLossy(60) on PNG failed: %v", err)
		}

		if palettedBuf.Len() == 0 {
			t.Fatalf("Output paletted PNG is empty")
		}
	})

	t.Run("Lossless PNG Optimization", func(t *testing.T) {
		var optBuf bytes.Buffer
		if err := codec.OptimizeLossless(ctx, bytes.NewReader(rawPNG), &optBuf, "image/png"); err != nil {
			t.Fatalf("OptimizeLossless on PNG failed: %v", err)
		}

		if optBuf.Len() == 0 {
			t.Fatalf("Optimized PNG output is empty")
		}
	})
}
