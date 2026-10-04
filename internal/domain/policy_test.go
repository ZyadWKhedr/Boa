package domain

import (
	"errors"
	"testing"
)

func TestLossyEligibilityPolicy(t *testing.T) {
	policy := NewLossyEligibilityPolicy()

	tests := []struct {
		name        string
		entry       FileEntry
		action      Action
		consent     LossyConsent
		wantEligible bool
		wantErr     error
	}{
		{
			name: "Document is never eligible for lossy",
			entry: FileEntry{
				Path: "report.pdf",
				Classification: FileClassification{
					MediaKind:   KindOther,
					Description: "PDF Document",
				},
			},
			action: Action{
				Kind: ActionTranscode,
				TranscodeParams: &TranscodeParams{
					Kind:     KindOther,
					Quality:  80,
					Lossless: false,
				},
			},
			consent: LossyConsent{
				AllowLossyImages: true,
				AllowLossyAudio:  true,
				AllowLossyVideo:  true,
			},
			wantEligible: false,
			wantErr:      ErrLossyDisallowedForNonMedia,
		},
		{
			name: "Source code file is never eligible for lossy",
			entry: FileEntry{
				Path: "main.go",
				Classification: FileClassification{
					MediaKind:   KindOther,
					Description: "Go Source Code",
				},
			},
			action: Action{
				Kind: ActionTranscode,
				TranscodeParams: &TranscodeParams{
					Kind:     KindOther,
					Quality:  50,
					Lossless: false,
				},
			},
			consent: LossyConsent{
				AllowLossyImages: true,
			},
			wantEligible: false,
			wantErr:      ErrLossyDisallowedForNonMedia,
		},
		{
			name: "Unknown / other binary is never eligible for lossy",
			entry: FileEntry{
				Path: "archive.tar",
				Classification: FileClassification{
					MediaKind:   KindOther,
					Description: "Tar Archive",
				},
			},
			action: Action{
				Kind: ActionTranscode,
				TranscodeParams: &TranscodeParams{
					Kind:     KindOther,
					Quality:  80,
					Lossless: false,
				},
			},
			consent:      LossyConsent{AllowLossyImages: true, AllowLossyAudio: true, AllowLossyVideo: true},
			wantEligible: false,
			wantErr:      ErrLossyDisallowedForNonMedia,
		},
		{
			name: "Image with explicit consent is allowed",
			entry: FileEntry{
				Path: "photo.jpg",
				Classification: FileClassification{
					MediaKind:   KindImage,
					Description: "JPEG Image",
				},
			},
			action: Action{
				Kind: ActionTranscode,
				TranscodeParams: &TranscodeParams{
					Kind:     KindImage,
					Quality:  80,
					Lossless: false,
				},
			},
			consent:      LossyConsent{AllowLossyImages: true},
			wantEligible: true,
			wantErr:      nil,
		},
		{
			name: "Image without consent is rejected",
			entry: FileEntry{
				Path: "photo.png",
				Classification: FileClassification{
					MediaKind:   KindImage,
					Description: "PNG Image",
				},
			},
			action: Action{
				Kind: ActionTranscode,
				TranscodeParams: &TranscodeParams{
					Kind:     KindImage,
					Quality:  80,
					Lossless: false,
				},
			},
			consent:      LossyConsent{AllowLossyImages: false},
			wantEligible: true,
			wantErr:      ErrLossyMissingConsent,
		},
		{
			name: "Audio with consent is allowed",
			entry: FileEntry{
				Path: "song.wav",
				Classification: FileClassification{
					MediaKind:   KindAudio,
					Description: "WAV Audio",
				},
			},
			action: Action{
				Kind: ActionTranscode,
				TranscodeParams: &TranscodeParams{
					Kind:     KindAudio,
					Quality:  64,
					Lossless: false,
				},
			},
			consent:      LossyConsent{AllowLossyAudio: true},
			wantEligible: true,
			wantErr:      nil,
		},
		{
			name: "Video with consent is allowed",
			entry: FileEntry{
				Path: "clip.mp4",
				Classification: FileClassification{
					MediaKind:   KindVideo,
					Description: "MP4 Video",
				},
			},
			action: Action{
				Kind: ActionTranscode,
				TranscodeParams: &TranscodeParams{
					Kind:     KindVideo,
					Quality:  70,
					Lossless: false,
				},
			},
			consent:      LossyConsent{AllowLossyVideo: true},
			wantEligible: true,
			wantErr:      nil,
		},
		{
			name: "Lossless Deflate is always allowed for any file without consent",
			entry: FileEntry{
				Path: "sensitive_data.db",
				Classification: FileClassification{
					MediaKind:   KindOther,
					Description: "Database File",
				},
			},
			action: Action{
				Kind:         ActionDeflate,
				DeflateLevel: 9,
			},
			consent:      LossyConsent{},
			wantEligible: false,
			wantErr:      nil,
		},
		{
			name: "Lossless Store is always allowed for any file",
			entry: FileEntry{
				Path: "archive.zip",
				Classification: FileClassification{
					MediaKind:   KindOther,
					Description: "Zip Archive",
				},
			},
			action: Action{
				Kind: ActionStore,
			},
			consent:      LossyConsent{},
			wantEligible: false,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eligible := policy.IsLossyEligible(tt.entry)
			if eligible != tt.wantEligible {
				t.Errorf("IsLossyEligible() = %v, want %v", eligible, tt.wantEligible)
			}

			err := policy.ValidateAction(tt.entry, tt.action, tt.consent)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("ValidateAction() expected error wrapping %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ValidateAction() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("ValidateAction() unexpected error: %v", err)
			}
		})
	}
}
