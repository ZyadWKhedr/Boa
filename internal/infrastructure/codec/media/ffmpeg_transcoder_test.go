package media

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestFFmpegTranscoderGracefulDegradation(t *testing.T) {
	// Simulate missing ffmpeg
	transcoder := &FFmpegTranscoder{ffmpegPath: ""}

	if transcoder.IsAvailable() {
		t.Errorf("Expected IsAvailable() = false for empty path")
	}

	hint := transcoder.InstallHint()
	if !strings.Contains(hint, "install FFmpeg") && !strings.Contains(hint, "brew") && !strings.Contains(hint, "apt") && !strings.Contains(hint, "winget") {
		t.Errorf("Expected friendly install hint, got %q", hint)
	}

	ctx := context.Background()
	err := transcoder.TranscodeAudio(ctx, "input.wav", "output.aac", 80, nil)
	if !errors.Is(err, ErrFFmpegNotFound) {
		t.Errorf("Expected ErrFFmpegNotFound, got %v", err)
	}

	err = transcoder.TranscodeVideo(ctx, "input.mov", "output.mp4", 75, nil)
	if !errors.Is(err, ErrFFmpegNotFound) {
		t.Errorf("Expected ErrFFmpegNotFound, got %v", err)
	}
}

func TestParameterMappings(t *testing.T) {
	transcoder := &FFmpegTranscoder{}

	t.Run("Audio Bitrates", func(t *testing.T) {
		if transcoder.mapAudioBitrate(20) != "96k" {
			t.Errorf("Expected 96k for quality 20, got %s", transcoder.mapAudioBitrate(20))
		}
		if transcoder.mapAudioBitrate(60) != "160k" {
			t.Errorf("Expected 160k for quality 60, got %s", transcoder.mapAudioBitrate(60))
		}
		if transcoder.mapAudioBitrate(90) != "256k" {
			t.Errorf("Expected 256k for quality 90, got %s", transcoder.mapAudioBitrate(90))
		}
	})

	t.Run("Video CRF", func(t *testing.T) {
		crfHigh := transcoder.mapVideoCRF(100)
		crfLow := transcoder.mapVideoCRF(20)

		if crfHigh != 18 {
			t.Errorf("Expected CRF 18 for quality 100, got %d", crfHigh)
		}
		if crfLow <= crfHigh {
			t.Errorf("Expected CRF for low quality (%d) to be higher than high quality (%d)", crfLow, crfHigh)
		}
	})
}
