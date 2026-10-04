package domain

// MediaKind categorizes files for compression policy decisions.
// The domain enforces an allowlist: only Image, Audio, and Video may be eligible for lossy processing.
// All other or unknown formats are categorized as Other and are strictly lossless.
type MediaKind string

const (
	// KindImage represents raster or vector image formats.
	KindImage MediaKind = "image"

	// KindAudio represents audio stream formats.
	KindAudio MediaKind = "audio"

	// KindVideo represents video stream / container formats.
	KindVideo MediaKind = "video"

	// KindOther represents documents, code, databases, archives, executables, fonts, or unknown files.
	// KindOther files can NEVER be processed lossy.
	KindOther MediaKind = "other"
)

// String returns the string representation of the MediaKind.
func (k MediaKind) String() string {
	return string(k)
}

// IsMedia returns true if the kind is Image, Audio, or Video.
func (k MediaKind) IsMedia() bool {
	switch k {
	case KindImage, KindAudio, KindVideo:
		return true
	default:
		return false
	}
}
