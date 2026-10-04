package domain

import "time"

// FileEntry represents a discovered file or directory to be packaged or analyzed.
type FileEntry struct {
	Path             string             `json:"path"`
	RelPath          string             `json:"rel_path"`
	IsDir            bool               `json:"is_dir"`
	SizeBytes        int64              `json:"size_bytes"`
	ModTime          time.Time          `json:"mod_time"`
	Mode             uint32             `json:"mode"`
	Classification   FileClassification `json:"classification"`
	AssignedAction   Action             `json:"assigned_action"`
}

// FileClassification details content-probed media properties.
type FileClassification struct {
	MIMEType           string    `json:"mime_type"`
	Extension          string    `json:"extension"`
	MediaKind          MediaKind `json:"media_kind"`
	IsAlreadyCompressed bool     `json:"is_already_compressed"`
	IsAlreadyLossy     bool      `json:"is_already_lossy"`
	Description        string    `json:"description"`
}
