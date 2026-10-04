package media

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"compressor/internal/domain"
)

var (
	// ErrFFmpegNotFound is returned when ffmpeg is not available on PATH.
	ErrFFmpegNotFound = errors.New("ffmpeg executable not found on system PATH")
)

// FFmpegTranscoder implements domain.MediaTranscoder using an external ffmpeg binary.
type FFmpegTranscoder struct {
	ffmpegPath string
}

// NewFFmpegTranscoder discovers ffmpeg on system PATH.
func NewFFmpegTranscoder() *FFmpegTranscoder {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return &FFmpegTranscoder{ffmpegPath: ""}
	}
	return &FFmpegTranscoder{ffmpegPath: path}
}

// IsAvailable returns true if ffmpeg is found and executable.
func (t *FFmpegTranscoder) IsAvailable() bool {
	return t.ffmpegPath != ""
}

// InstallHint returns friendly, OS-specific instructions to install FFmpeg.
func (t *FFmpegTranscoder) InstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "To enable smaller audio/video files, install FFmpeg using Homebrew: brew install ffmpeg"
	case "windows":
		return "To enable smaller audio/video files, install FFmpeg using Windows Package Manager: winget install Gyan.FFmpeg"
	case "linux":
		return "To enable smaller audio/video files, install FFmpeg using your package manager: sudo apt install ffmpeg"
	default:
		return "To enable smaller audio/video files, please install FFmpeg from https://ffmpeg.org/download.html"
	}
}

// TranscodeAudio transcodes audio using AAC or MP3 based on quality slider (0-100).
func (t *FFmpegTranscoder) TranscodeAudio(
	ctx context.Context,
	srcPath, dstPath string,
	quality int,
	progress func(domain.TranscodeProgress),
) error {
	if !t.IsAvailable() {
		return ErrFFmpegNotFound
	}

	bitrate := t.mapAudioBitrate(quality)

	// Safe argument slice - never shell formatted
	args := []string{
		"-y", // overwrite output
		"-v", "error",
		"-progress", "pipe:1",
		"-i", srcPath,
		"-vn", // drop video streams
		"-c:a", "aac",
		"-b:a", bitrate,
		dstPath,
	}

	return t.runFFmpeg(ctx, args, progress)
}

// TranscodeVideo transcodes video using H.264 CRF encoding with optional downscaling.
func (t *FFmpegTranscoder) TranscodeVideo(
	ctx context.Context,
	srcPath, dstPath string,
	quality int,
	progress func(domain.TranscodeProgress),
) error {
	if !t.IsAvailable() {
		return ErrFFmpegNotFound
	}

	crf := t.mapVideoCRF(quality)

	args := []string{
		"-y",
		"-v", "error",
		"-progress", "pipe:1",
		"-i", srcPath,
		"-c:v", "libx264",
		"-preset", "faster",
		"-crf", strconv.Itoa(crf),
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		dstPath,
	}

	return t.runFFmpeg(ctx, args, progress)
}

func (t *FFmpegTranscoder) mapAudioBitrate(quality int) string {
	switch {
	case quality <= 40:
		return "96k" // Small
	case quality <= 75:
		return "160k" // Balanced
	default:
		return "256k" // High
	}
}

func (t *FFmpegTranscoder) mapVideoCRF(quality int) int {
	// CRF scale: 18 (near lossless) to 36 (very small/lower quality)
	// Quality 100 -> CRF 18
	// Quality 70  -> CRF 23
	// Quality 50  -> CRF 28
	// Quality 10  -> CRF 35
	if quality <= 0 {
		quality = 70
	}
	if quality > 100 {
		quality = 100
	}

	crf := 36 - int(float64(quality)*(18.0/100.0))
	if crf < 18 {
		crf = 18
	}
	if crf > 36 {
		crf = 36
	}
	return crf
}

func (t *FFmpegTranscoder) runFFmpeg(
	ctx context.Context,
	args []string,
	progress func(domain.TranscodeProgress),
) error {
	cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, t.ffmpegPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	if progress != nil {
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "out_time_ms=") {
					valStr := strings.TrimPrefix(line, "out_time_ms=")
					if us, err := strconv.ParseInt(valStr, 10, 64); err == nil {
						progress(domain.TranscodeProgress{
							Percent: float64(us) / 1000.0,
							Stage:   "encoding",
						})
					}
				}
			}
		}()
	}

	if err := cmd.Wait(); err != nil {
		if errors.Is(cmdCtx.Err(), context.Canceled) {
			return errors.New("transcoding cancelled by user")
		}
		if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) {
			return errors.New("transcoding timed out")
		}
		return fmt.Errorf("ffmpeg execution failed: %w", err)
	}

	return nil
}

var _ domain.MediaTranscoder = (*FFmpegTranscoder)(nil)
