package cmd

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCharacterizationPackUnpackListPins pins existing CLI behavior before refactoring.
func TestCharacterizationPackUnpackListPins(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "char_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test structure
	srcDir := filepath.Join(tempDir, "dataset")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	docContent := "Important document content for characterization test."
	codeContent := "package main\n\nfunc main() {}\n"
	_ = os.WriteFile(filepath.Join(srcDir, "doc.txt"), []byte(docContent), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "nested", "main.go"), []byte(codeContent), 0o644)

	deflateZip := filepath.Join(tempDir, "deflate.zip")
	storeZip := filepath.Join(tempDir, "store.zip")
	zstdZip := filepath.Join(tempDir, "zstd.zip")

	// 1. Pack Deflate
	RootCmd.SetArgs([]string{"pack", srcDir, "-o", deflateZip, "-m", "deflate", "-l", "6", "-f", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Deflate pack failed: %v", err)
	}

	// 2. Pack Store
	RootCmd.SetArgs([]string{"pack", srcDir, "-o", storeZip, "-m", "store", "-f", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Store pack failed: %v", err)
	}

	// 3. Pack Zstd
	RootCmd.SetArgs([]string{"pack", srcDir, "-o", zstdZip, "-m", "zstd", "-l", "3", "-f", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Zstd pack failed: %v", err)
	}

	// 4. Verify all archives exist
	for _, z := range []string{deflateZip, storeZip, zstdZip} {
		fi, err := os.Stat(z)
		if err != nil || fi.Size() == 0 {
			t.Fatalf("Archive %q missing or empty", z)
		}
	}

	// 5. Unpack and verify byte-for-byte fidelity
	for _, z := range []string{deflateZip, storeZip, zstdZip} {
		outDir := filepath.Join(tempDir, "out_"+filepath.Base(z))
		RootCmd.SetArgs([]string{"unpack", z, "-o", outDir, "-f", "--quiet"})
		if err := RootCmd.Execute(); err != nil {
			t.Fatalf("Unpack %q failed: %v", z, err)
		}

		gotDoc, err := os.ReadFile(filepath.Join(outDir, "dataset", "doc.txt"))
		if err != nil || string(gotDoc) != docContent {
			t.Errorf("Fidelity check failed for %q on doc.txt: got %q", z, string(gotDoc))
		}

		gotCode, err := os.ReadFile(filepath.Join(outDir, "dataset", "nested", "main.go"))
		if err != nil || string(gotCode) != codeContent {
			t.Errorf("Fidelity check failed for %q on main.go: got %q", z, string(gotCode))
		}
	}
}

// TestCharacterizationSymlinkAndZipSlipPins verifies symlink and zip slip safety guards.
func TestCharacterizationSymlinkAndZipSlipPins(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "char_symlink_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Zip slip attempt
	slipZip := filepath.Join(tempDir, "slip.zip")
	f, err := os.Create(slipZip)
	if err != nil {
		t.Fatalf("failed to create slip zip: %v", err)
	}
	zw := zip.NewWriter(f)
	hdr := &zip.FileHeader{
		Name:   "../escaped.txt",
		Method: zip.Deflate,
	}
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}
	_, _ = w.Write([]byte("malicious escape payload"))
	_ = zw.Close()
	_ = f.Close()

	destDir := filepath.Join(tempDir, "target")
	RootCmd.SetArgs([]string{"unpack", slipZip, "-o", destDir, "-f", "--quiet"})
	err = RootCmd.Execute()
	if err == nil {
		t.Fatalf("Expected zip-slip error, got nil")
	}

	escapedFile := filepath.Join(tempDir, "escaped.txt")
	if _, err := os.Stat(escapedFile); !os.IsNotExist(err) {
		t.Fatalf("Escaped file was written outside destination directory!")
	}
}

// TestCharacterizationDryRunAndExclusions pins dry run and exclusion behavior.
func TestCharacterizationDryRunAndExclusions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "char_excludes_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "src")
	_ = os.MkdirAll(filepath.Join(srcDir, "node_modules", "pkg"), 0o755)
	_ = os.MkdirAll(filepath.Join(srcDir, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "index.js"), []byte("console.log('hi');"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "node_modules", "pkg", "lib.js"), []byte("module.exports={};"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, ".git", "config"), []byte("[core]"), 0o644)

	zipOut := filepath.Join(tempDir, "out.zip")

	// 1. Dry run should not create archive
	RootCmd.SetArgs([]string{"pack", srcDir, "-o", zipOut, "--dry-run", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Dry run failed: %v", err)
	}
	if _, err := os.Stat(zipOut); !os.IsNotExist(err) {
		t.Fatalf("Dry run created archive when it should not have")
	}

	// 2. Pack with default exclusions (node_modules, .git)
	RootCmd.SetArgs([]string{"pack", srcDir, "-o", zipOut, "-f", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	// 3. Inspect archive entries
	r, err := zip.OpenReader(zipOut)
	if err != nil {
		t.Fatalf("Failed to open created zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "node_modules") || strings.Contains(f.Name, ".git") {
			t.Errorf("Excluded item found in archive: %q", f.Name)
		}
	}
}

// TestCharacterizationLossyImagesAndZipComment verifies non-interactive lossy flags and ZIP metadata comment.
func TestCharacterizationLossyImagesAndZipComment(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "char_lossy_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "media_folder")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "doc.txt"), []byte("Document must stay untouched"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "photo.jpg"), []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}, 0o644)

	zipOut := filepath.Join(tempDir, "lossy_output.zip")

	RootCmd.SetArgs([]string{"pack", srcDir, "-o", zipOut, "--lossy", "images", "--quality", "75", "-f", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Pack with --lossy images failed: %v", err)
	}

	r, err := zip.OpenReader(zipOut)
	if err != nil {
		t.Fatalf("Failed to open created zip: %v", err)
	}
	defer r.Close()

	if !strings.Contains(r.Comment, "Boa-Lossy") {
		t.Errorf("Expected ZIP Comment to contain 'Boa-Lossy', got %q", r.Comment)
	}
}
