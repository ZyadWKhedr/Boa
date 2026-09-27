package safety

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrZipSlip is returned when a zip file contains a path attempting directory traversal.
	ErrZipSlip = errors.New("security violation: Zip Slip path traversal attempt detected")

	// ErrInvalidPath is returned when a path contains invalid characters or null bytes.
	ErrInvalidPath = errors.New("security violation: path contains invalid characters or null bytes")

	// ErrSymlinkEscape is returned when a symlink points outside the destination root.
	ErrSymlinkEscape = errors.New("security violation: symlink points outside destination boundary")
)

// ValidateDestinationPath checks that joining targetName with destDir does not escape destDir (Zip-Slip defense).
// It returns the clean, absolute destination path on success.
func ValidateDestinationPath(destDir, targetName string) (string, error) {
	// Guard against null byte injections
	if strings.ContainsRune(targetName, 0) {
		return "", fmt.Errorf("%w: null byte detected in %q", ErrInvalidPath, targetName)
	}

	// Reject absolute paths and leading slashes/backslashes/drive letters
	if strings.HasPrefix(targetName, "/") || strings.HasPrefix(targetName, "\\") || filepath.IsAbs(targetName) || filepath.VolumeName(targetName) != "" {
		return "", fmt.Errorf("%w: absolute path or drive letter detected in %q", ErrZipSlip, targetName)
	}

	// Clean destination directory and target name
	cleanDest, err := filepath.Abs(filepath.Clean(destDir))
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute destination path: %w", err)
	}

	// Normalize separators for cross-platform (zip uses forward slashes internally)
	sanitizedName := filepath.Clean(filepath.FromSlash(targetName))

	// Recheck if cleaned relative path tries to escape
	if strings.HasPrefix(sanitizedName, ".."+string(filepath.Separator)) || sanitizedName == ".." {
		return "", fmt.Errorf("%w: path %q attempts to escape root", ErrZipSlip, targetName)
	}

	// Resolve the complete combined target path
	fullPath := filepath.Join(cleanDest, sanitizedName)
	cleanFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Verify that cleanFullPath starts with cleanDest + separator (or equals cleanDest)
	destPrefix := cleanDest + string(filepath.Separator)
	if cleanFullPath != cleanDest && !strings.HasPrefix(cleanFullPath, destPrefix) {
		return "", fmt.Errorf("%w: target %q escapes destination directory %q", ErrZipSlip, targetName, cleanDest)
	}

	return cleanFullPath, nil
}

// ValidateSymlinkTarget ensures a symlink's target does not point outside the base directory.
func ValidateSymlinkTarget(destDir, linkPath, targetPath string) error {
	var resolvedTarget string
	if filepath.IsAbs(targetPath) {
		resolvedTarget = filepath.Clean(targetPath)
	} else {
		resolvedTarget = filepath.Join(filepath.Dir(linkPath), targetPath)
	}

	cleanDest, err := filepath.Abs(filepath.Clean(destDir))
	if err != nil {
		return err
	}

	cleanTarget, err := filepath.Abs(resolvedTarget)
	if err != nil {
		return err
	}

	destPrefix := cleanDest + string(filepath.Separator)
	if cleanTarget != cleanDest && !strings.HasPrefix(cleanTarget, destPrefix) {
		return fmt.Errorf("%w: link target %q resolves to %q outside %q", ErrSymlinkEscape, targetPath, cleanTarget, cleanDest)
	}

	return nil
}

// MatchExclude returns true if the relative path matches any exclusion glob pattern.
func MatchExclude(relPath string, excludePatterns []string) bool {
	if len(excludePatterns) == 0 {
		return false
	}

	normalized := filepath.ToSlash(relPath)
	base := filepath.Base(relPath)

	for _, pattern := range excludePatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// Direct glob match on basename (e.g. *.DS_Store, *.tmp, node_modules)
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}

		// Direct glob match on full relative path
		if matched, _ := filepath.Match(pattern, normalized); matched {
			return true
		}

		// Check prefix or substring pattern (e.g. "node_modules/", ".git/")
		cleanPattern := strings.Trim(filepath.ToSlash(pattern), "/")
		pathParts := strings.Split(normalized, "/")
		for _, part := range pathParts {
			if matched, _ := filepath.Match(cleanPattern, part); matched {
				return true
			}
		}
	}

	return false
}

// FileExists checks if a file or directory exists at the specified path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
