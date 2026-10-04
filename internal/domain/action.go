package domain

import "fmt"

// ActionKind represents the action to apply to a file entry during compression packaging.
type ActionKind string

const (
	// ActionStore writes the file uncompressed (ZIP Method 0).
	ActionStore ActionKind = "store"

	// ActionDeflate compresses the file using standard DEFLATE (ZIP Method 8).
	ActionDeflate ActionKind = "deflate"

	// ActionZstd compresses the file using Zstandard (ZIP Method 93).
	ActionZstd ActionKind = "zstd"

	// ActionTranscode applies media-specific transcoding (e.g. image, audio, video optimization)
	// before writing to the archive. Only allowed for KindImage, KindAudio, KindVideo.
	ActionTranscode ActionKind = "transcode"
)

// TranscodeParams holds parameters for media transcoding operations.
type TranscodeParams struct {
	Kind     MediaKind `json:"kind"`
	Quality  int       `json:"quality"`  // Normalized 0-100 quality setting
	Lossless bool      `json:"lossless"` // Whether transcoding should be lossless (e.g. PNG optimization, FLAC)
}

// Action encapsulates the decided operation for a file.
type Action struct {
	Kind            ActionKind       `json:"kind"`
	DeflateLevel    int              `json:"deflate_level,omitempty"`
	ZstdLevel       int              `json:"zstd_level,omitempty"`
	TranscodeParams *TranscodeParams `json:"transcode_params,omitempty"`
}

// String returns a human-readable description of the Action.
func (a Action) String() string {
	switch a.Kind {
	case ActionStore:
		return "Store (uncompressed)"
	case ActionDeflate:
		return fmt.Sprintf("DEFLATE (level %d)", a.DeflateLevel)
	case ActionZstd:
		return fmt.Sprintf("Zstandard (level %d)", a.ZstdLevel)
	case ActionTranscode:
		if a.TranscodeParams != nil {
			if a.TranscodeParams.Lossless {
				return fmt.Sprintf("Lossless Transcode (%s)", a.TranscodeParams.Kind)
			}
			return fmt.Sprintf("Lossy Transcode (%s, quality %d%%)", a.TranscodeParams.Kind, a.TranscodeParams.Quality)
		}
		return "Transcode"
	default:
		return string(a.Kind)
	}
}
