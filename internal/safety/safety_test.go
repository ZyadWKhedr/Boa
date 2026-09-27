package safety

import (
	"os"
	"testing"
)

func TestValidateDestinationPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "safety_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name       string
		destDir    string
		targetName string
		wantErr    bool
	}{
		{
			name:       "valid standard relative path",
			destDir:    tempDir,
			targetName: "folder/file.txt",
			wantErr:    false,
		},
		{
			name:       "valid simple filename",
			destDir:    tempDir,
			targetName: "file.txt",
			wantErr:    false,
		},
		{
			name:       "valid deep nested path",
			destDir:    tempDir,
			targetName: "a/b/c/d/e.txt",
			wantErr:    false,
		},
		{
			name:       "zip slip path traversal attack with ..",
			destDir:    tempDir,
			targetName: "../../../etc/passwd",
			wantErr:    true,
		},
		{
			name:       "zip slip hidden escape inside subfolder",
			destDir:    tempDir,
			targetName: "folder/../../../../secret.txt",
			wantErr:    true,
		},
		{
			name:       "zip slip absolute path escape",
			destDir:    tempDir,
			targetName: "/etc/shadow",
			wantErr:    true,
		},
		{
			name:       "null byte injection attack",
			destDir:    tempDir,
			targetName: "valid/path\x00escape.txt",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ValidateDestinationPath(tt.destDir, tt.targetName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDestinationPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && path == "" {
				t.Errorf("ValidateDestinationPath() returned empty path for valid target")
			}
		})
	}
}

func TestMatchExclude(t *testing.T) {
	excludes := []string{"*.git*", ".DS_Store", "node_modules", "*.tmp"}

	tests := []struct {
		path     string
		expected bool
	}{
		{"src/main.go", false},
		{".git/config", true},
		{".DS_Store", true},
		{"sub/.DS_Store", true},
		{"node_modules/package/index.js", true},
		{"cache/temp.tmp", true},
		{"src/component.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := MatchExclude(tt.path, excludes)
			if got != tt.expected {
				t.Errorf("MatchExclude(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
