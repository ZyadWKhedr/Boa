package compress

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"compressor/pkg/types"
)

func TestPackAndInspect(t *testing.T) {
	// Create temporary source structure
	srcDir, err := os.MkdirTemp("", "compress_src_*")
	if err != nil {
		t.Fatalf("failed to create src temp dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	subDir := filepath.Join(srcDir, "subfolder")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subfolder: %v", err)
	}

	file1 := filepath.Join(srcDir, "hello.txt")
	if err := os.WriteFile(file1, []byte("Hello World from Compressor CLI!"), 0o644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	file2 := filepath.Join(subDir, "nested.txt")
	if err := os.WriteFile(file2, []byte("Nested content data inside subfolder"), 0o644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	zipOut := filepath.Join(srcDir, "output.zip")

	engine := New()
	opts := types.PackOptions{
		SourcePaths:      []string{srcDir},
		OutputZipPath:    zipOut,
		CompressionLevel: 6,
		Overwrite:        true,
	}

	summary, err := engine.Pack(opts)
	if err != nil {
		t.Fatalf("Pack() failed: %v", err)
	}

	if summary.TotalFiles < 2 {
		t.Errorf("Expected at least 2 files packed, got %d", summary.TotalFiles)
	}

	// Verify zip contents
	reader, err := zip.OpenReader(zipOut)
	if err != nil {
		t.Fatalf("Failed to open generated zip: %v", err)
	}
	defer reader.Close()

	foundHello := false
	for _, f := range reader.File {
		if filepath.Base(f.Name) == "hello.txt" {
			foundHello = true
		}
	}

	if !foundHello {
		t.Errorf("hello.txt not found inside zip archive")
	}
}

func TestPackDryRun(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "compress_dryrun_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	testFile := filepath.Join(srcDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("Dry run test data"), 0o644)

	zipOut := filepath.Join(srcDir, "never_created.zip")

	engine := New()
	opts := types.PackOptions{
		SourcePaths:   []string{srcDir},
		OutputZipPath: zipOut,
		DryRun:        true,
	}

	summary, err := engine.Pack(opts)
	if err != nil {
		t.Fatalf("Pack() in dry-run mode failed: %v", err)
	}

	if summary.TotalFiles != 1 {
		t.Errorf("Expected 1 file in dry-run summary, got %d", summary.TotalFiles)
	}

	if _, err := os.Stat(zipOut); !os.IsNotExist(err) {
		t.Errorf("Dry-run created file on disk when it should not have")
	}
}

func TestPackStoreMode(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "compress_store_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	file1 := filepath.Join(srcDir, "raw.txt")
	testData := []byte("Uncompressed raw storage testing data string with repeated repetitive patterns 123456789")
	if err := os.WriteFile(file1, testData, 0o644); err != nil {
		t.Fatalf("failed to write raw.txt: %v", err)
	}

	zipOut := filepath.Join(srcDir, "store.zip")

	engine := New()
	opts := types.PackOptions{
		SourcePaths:      []string{file1},
		OutputZipPath:    zipOut,
		Method:           types.MethodStore,
		CompressionLevel: 0,
		Overwrite:        true,
	}

	summary, err := engine.Pack(opts)
	if err != nil {
		t.Fatalf("Pack with MethodStore failed: %v", err)
	}

	if summary.CompressionMethod != "Store" {
		t.Errorf("Expected summary CompressionMethod 'Store', got %q", summary.CompressionMethod)
	}

	// Verify that header method is zip.Store (0)
	reader, err := zip.OpenReader(zipOut)
	if err != nil {
		t.Fatalf("Failed to open store.zip: %v", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		t.Fatalf("Expected at least 1 entry in store.zip")
	}

	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			if f.Method != zip.Store {
				t.Errorf("Expected zip method Store (%d), got %d", zip.Store, f.Method)
			}
		}
	}
}

func TestPackZstdMode(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "compress_zstd_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	file1 := filepath.Join(srcDir, "payload.txt")
	testData := []byte("Zstandard modern compression testing data string with lots of repetition repetition repetition 123456789")
	if err := os.WriteFile(file1, testData, 0o644); err != nil {
		t.Fatalf("failed to write payload.txt: %v", err)
	}

	zipOut := filepath.Join(srcDir, "zstd.zip")

	engine := New()
	opts := types.PackOptions{
		SourcePaths:      []string{file1},
		OutputZipPath:    zipOut,
		Method:           types.MethodZstd,
		CompressionLevel: 3,
		Overwrite:        true,
	}

	summary, err := engine.Pack(opts)
	if err != nil {
		t.Fatalf("Pack with MethodZstd failed: %v", err)
	}

	if summary.CompressionMethod != "Zstandard" {
		t.Errorf("Expected summary CompressionMethod 'Zstandard', got %q", summary.CompressionMethod)
	}

	// Verify that header method is 93 (ZipMethodZstd)
	reader, err := zip.OpenReader(zipOut)
	if err != nil {
		t.Fatalf("Failed to open zstd.zip: %v", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		t.Fatalf("Expected at least 1 entry in zstd.zip")
	}

	for _, f := range reader.File {
		if !f.FileInfo().IsDir() {
			if f.Method != types.ZipMethodZstd {
				t.Errorf("Expected zip method Zstd (%d), got %d", types.ZipMethodZstd, f.Method)
			}
		}
	}
}
