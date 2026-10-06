package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	RepoOwner = "ZyadWKhedr"
	RepoName  = "Boa"
	apiURL    = "https://api.github.com/repos/ZyadWKhedr/Boa/releases/latest"
)

// Asset represents a release binary asset in GitHub Releases.
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

var (
	cacheMu       sync.Mutex
	cachedRelease *Release
	cachedTime    time.Time
)

// CheckForUpdate queries GitHub for the latest release and compares with the current version.
func CheckForUpdate(ctx context.Context, currentVersion string) (*Release, bool, error) {
	cacheMu.Lock()
	if cachedRelease != nil && time.Since(cachedTime) < 5*time.Minute {
		rel := cachedRelease
		cacheMu.Unlock()
		isNewer := IsNewer(rel.TagName, currentVersion)
		return rel, isNewer, nil
	}
	cacheMu.Unlock()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "Boa-Updater/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, false, err
	}

	cacheMu.Lock()
	cachedRelease = &rel
	cachedTime = time.Now()
	cacheMu.Unlock()

	isNewer := IsNewer(rel.TagName, currentVersion)
	return &rel, isNewer, nil
}

// IsNewer returns true if candidate version is strictly newer than current version.
func IsNewer(candidate, current string) bool {
	candParts := parseSemver(candidate)
	currParts := parseSemver(current)

	for i := 0; i < 3; i++ {
		if candParts[i] > currParts[i] {
			return true
		}
		if candParts[i] < currParts[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	// Strip build metadata/prerelease
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	var res [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		val, _ := strconv.Atoi(parts[i])
		res[i] = val
	}
	return res
}

// DownloadAndInstall performs automatic binary replacement for the current platform.
func DownloadAndInstall(
	ctx context.Context,
	rel *Release,
	progressCb func(downloaded, total int64),
) (string, error) {
	if rel == nil || len(rel.Assets) == 0 {
		return "", errors.New("no release assets available")
	}

	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Expected binary asset naming pattern: boa-{os}-{arch} or boa-{os}-{arch}.exe
	expectedSuffix := fmt.Sprintf("boa-%s-%s", osName, archName)
	if osName == "windows" {
		expectedSuffix += ".exe"
	}

	var targetAsset *Asset
	for i := range rel.Assets {
		if strings.EqualFold(rel.Assets[i].Name, expectedSuffix) {
			targetAsset = &rel.Assets[i]
			break
		}
	}

	if targetAsset == nil {
		return "", fmt.Errorf("no pre-built binary asset found for %s-%s in release %s", osName, archName, rel.TagName)
	}

	// Download asset to temp file
	req, err := http.NewRequestWithContext(ctx, "GET", targetAsset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Boa-Updater")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download update asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	tempDir, err := os.MkdirTemp("", "boa_update_*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	tempBinaryPath := filepath.Join(tempDir, "boa_new")
	tempFile, err := os.OpenFile(tempBinaryPath, os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		return "", err
	}

	var downloaded int64
	totalBytes := targetAsset.Size
	buf := make([]byte, 32*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tempFile.Write(buf[:n]); writeErr != nil {
				tempFile.Close()
				return "", writeErr
			}
			downloaded += int64(n)
			if progressCb != nil {
				progressCb(downloaded, totalBytes)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			tempFile.Close()
			return "", readErr
		}
	}
	tempFile.Close()

	// Locate installation target
	installPath := findInstallPath()
	targetDir := filepath.Dir(installPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("cannot access target directory %q: %w", targetDir, err)
	}

	// Replace existing binary
	_ = os.Remove(installPath)
	if err := copyAndReplace(tempBinaryPath, installPath); err != nil {
		return "", fmt.Errorf("failed to install new binary to %q: %w", installPath, err)
	}
	_ = os.Chmod(installPath, 0o755)

	// Ad-hoc sign on macOS to prevent Gatekeeper / AMFI kill
	if runtime.GOOS == "darwin" {
		_ = exec.Command("codesign", "-s", "-", "-f", installPath).Run()
	}

	// Update aliases
	aliasPath := filepath.Join(targetDir, "bo")
	_ = os.Remove(aliasPath)
	_ = os.Symlink(installPath, aliasPath)

	compressorPath := filepath.Join(targetDir, "compressor")
	_ = os.Remove(compressorPath)
	_ = os.Symlink(installPath, compressorPath)

	return installPath, nil
}

func findInstallPath() string {
	if execPath, err := os.Executable(); err == nil {
		if realPath, err := filepath.EvalSymlinks(execPath); err == nil && realPath != "" {
			if canWrite(realPath) {
				return realPath
			}
		}
	}

	home, _ := os.UserHomeDir()
	localBin := filepath.Join(home, ".local", "bin", "boa")
	if canWrite(filepath.Dir(localBin)) {
		return localBin
	}

	usrBin := "/usr/local/bin/boa"
	if canWrite(filepath.Dir(usrBin)) {
		return usrBin
	}

	return localBin
}

func canWrite(path string) bool {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		testFile := filepath.Join(path, ".boa_perm_test")
		if err := os.WriteFile(testFile, []byte(""), 0o644); err == nil {
			_ = os.Remove(testFile)
			return true
		}
		return false
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0o755)
	if err == nil {
		f.Close()
		return true
	}
	return false
}

func copyAndReplace(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
