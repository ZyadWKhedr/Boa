package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"compressor/internal/domain"
)

// StandardImageCodec implements domain.ImageCodec using Go's standard library.
type StandardImageCodec struct{}

// NewStandardImageCodec creates a new instance of StandardImageCodec.
func NewStandardImageCodec() *StandardImageCodec {
	return &StandardImageCodec{}
}

// CanHandle returns true for supported image MIME types.
func (c *StandardImageCodec) CanHandle(mimeType string) bool {
	switch strings.ToLower(mimeType) {
	case "image/jpeg", "image/jpg", "image/png":
		return true
	default:
		return false
	}
}

// OptimizeLossless re-encodes images losslessly where possible (e.g., PNG).
// For JPEG, lossless recompression is not possible in standard Go, so it streams untouched.
func (c *StandardImageCodec) OptimizeLossless(ctx context.Context, src io.Reader, dst io.Writer, mimeType string) error {
	rawBytes, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	if strings.ToLower(mimeType) == "image/png" {
		img, err := png.Decode(bytes.NewReader(rawBytes))
		if err == nil {
			var buf bytes.Buffer
			encoder := png.Encoder{
				CompressionLevel: png.BestCompression,
			}
			if err := encoder.Encode(&buf, img); err == nil {
				// Only use optimized output if it actually shrank the size
				if buf.Len() < len(rawBytes) {
					_, err = io.Copy(dst, &buf)
					return err
				}
			}
		}
	}

	// Default fallback: copy untouched
	_, err = io.Copy(dst, bytes.NewReader(rawBytes))
	return err
}

// CompressLossy applies lossy image compression with quality 1-100.
func (c *StandardImageCodec) CompressLossy(ctx context.Context, src io.Reader, dst io.Writer, mimeType string, quality int) error {
	if quality <= 0 {
		quality = 75
	}
	if quality > 100 {
		quality = 100
	}

	rawBytes, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	mimeLower := strings.ToLower(mimeType)

	switch mimeLower {
	case "image/jpeg", "image/jpg":
		img, err := jpeg.Decode(bytes.NewReader(rawBytes))
		if err != nil {
			// Cannot decode: fallback to raw stream
			_, copyErr := io.Copy(dst, bytes.NewReader(rawBytes))
			return copyErr
		}

		var buf bytes.Buffer
		opts := &jpeg.Options{Quality: quality}
		if err := jpeg.Encode(&buf, img, opts); err != nil {
			return fmt.Errorf("jpeg encode failed: %w", err)
		}

		// Only write if not larger than original
		if buf.Len() < len(rawBytes) || len(rawBytes) == 0 {
			_, err = io.Copy(dst, &buf)
			return err
		}
		_, err = io.Copy(dst, bytes.NewReader(rawBytes))
		return err

	case "image/png":
		img, err := png.Decode(bytes.NewReader(rawBytes))
		if err != nil {
			_, copyErr := io.Copy(dst, bytes.NewReader(rawBytes))
			return copyErr
		}

		var buf bytes.Buffer
		// If quality < 85, quantize to 256-color palette image
		if quality < 85 {
			bounds := img.Bounds()
			paletted := image.NewPaletted(bounds, palette.Plan9)
			draw.FloydSteinberg.Draw(paletted, bounds, img, image.Point{})
			encoder := png.Encoder{CompressionLevel: png.BestCompression}
			if err := encoder.Encode(&buf, paletted); err == nil && (buf.Len() < len(rawBytes) || len(rawBytes) == 0) {
				_, err = io.Copy(dst, &buf)
				return err
			}
		}

		// Otherwise, lossless PNG compression
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&buf, img); err == nil && (buf.Len() < len(rawBytes) || len(rawBytes) == 0) {
			_, err = io.Copy(dst, &buf)
			return err
		}

		_, err = io.Copy(dst, bytes.NewReader(rawBytes))
		return err

	default:
		// Unsupported lossy format: stream as-is
		_, err = io.Copy(dst, bytes.NewReader(rawBytes))
		return err
	}
}

var _ domain.ImageCodec = (*StandardImageCodec)(nil)

// Helper to generate a solid or gradient test image for tests.
func CreateTestRGBImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: uint8((x + y) % 255),
				A: 255,
			})
		}
	}
	return img
}
