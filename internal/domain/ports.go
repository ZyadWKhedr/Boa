package domain

import (
	"context"
	"io"
)

// FileClassifier inspects file content and path to determine its classification.
type FileClassifier interface {
	Classify(path string) (FileClassification, error)
	ClassifyReader(r io.Reader, filenameHint string) (FileClassification, error)
}

// ImageCodec defines operations for image optimization and lossy transcoding.
type ImageCodec interface {
	CanHandle(mimeType string) bool
	OptimizeLossless(ctx context.Context, src io.Reader, dst io.Writer, mimeType string) error
	CompressLossy(ctx context.Context, src io.Reader, dst io.Writer, mimeType string, quality int) error
}

// TranscodeProgress reports ongoing media transcoding progress.
type TranscodeProgress struct {
	Percent float64
	Stage   string
}

// MediaTranscoder defines operations for audio/video transcoding (e.g. via external tools like FFmpeg).
type MediaTranscoder interface {
	IsAvailable() bool
	InstallHint() string
	TranscodeAudio(ctx context.Context, srcPath, dstPath string, quality int, progress func(TranscodeProgress)) error
	TranscodeVideo(ctx context.Context, srcPath, dstPath string, quality int, progress func(TranscodeProgress)) error
}

// TechniqueCatalog defines operations for retrieving educational technique explanations.
type TechniqueCatalog interface {
	ListTechniques() []TechniqueInfo
	GetTechniqueInfo(id string) (TechniqueInfo, bool)
}
