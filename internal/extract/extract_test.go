package extract

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"compressor/internal/compress"
	"compressor/pkg/types"
)

func TestUnpackAndVerify(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "extract_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "extracted")
	zipPath := filepath.Join(tempDir, "archive.zip")

	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "doc.txt"), []byte("Document content inside zip test"), 0o644)

	compEngine := compress.New()
	_, err = compEngine.Pack(types.PackOptions{
		SourcePaths:   []string{srcDir},
		OutputZipPath: zipPath,
		Overwrite:     true,
	})
	if err != nil {
		t.Fatalf("Pack failed: %v", err)
	}

	extractEngine := New()
	summary, err := extractEngine.Unpack(types.UnpackOptions{
		ArchivePath:    zipPath,
		DestinationDir: destDir,
		Overwrite:      true,
	})
	if err != nil {
		t.Fatalf("Unpack failed: %v", err)
	}

	if summary.TotalFiles < 1 {
		t.Errorf("Expected at least 1 file extracted, got %d", summary.TotalFiles)
	}

	// Verify extracted file exists and has correct content
	extractedDoc := filepath.Join(destDir, "source", "doc.txt")
	content, err := os.ReadFile(extractedDoc)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(content) != "Document content inside zip test" {
		t.Errorf("Extracted content mismatch: got %q", string(content))
	}
}

func TestZipSlipAttackBlocked(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zipslip_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	maliciousZip := filepath.Join(tempDir, "malicious.zip")
	destDir := filepath.Join(tempDir, "target")

	// Craft a malicious zip containing a traversal path "../evil.txt"
	f, err := os.Create(maliciousZip)
	if err != nil {
		t.Fatalf("failed to create malicious zip: %v", err)
	}
	zw := zip.NewWriter(f)

	header := &zip.FileHeader{
		Name:   "../../evil.txt",
		Method: zip.Store,
	}
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("failed to create malicious header: %v", err)
	}
	_, _ = w.Write([]byte("malicious payload outside target folder"))
	_ = zw.Close()
	_ = f.Close()

	extractEngine := New()
	_, err = extractEngine.Unpack(types.UnpackOptions{
		ArchivePath:    maliciousZip,
		DestinationDir: destDir,
		Overwrite:      true,
	})

	if err == nil {
		t.Fatalf("Expected Zip Slip attack to be rejected, but Unpack succeeded!")
	}

	if !strings.Contains(err.Error(), "Zip Slip") && !strings.Contains(err.Error(), "security validation") {
		t.Errorf("Expected security violation error, got: %v", err)
	}

	// Verify evil file was NOT created outside target
	evilFile := filepath.Join(tempDir, "evil.txt")
	if _, err := os.Stat(evilFile); !os.IsNotExist(err) {
		t.Fatalf("Security failure: evil.txt was written outside target boundary!")
	}
}

func TestUnpackStoreMode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "extract_store_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "extracted")
	zipPath := filepath.Join(tempDir, "store.zip")

	_ = os.MkdirAll(srcDir, 0o755)
	testContent := []byte("Store mode extraction payload data content string")
	_ = os.WriteFile(filepath.Join(srcDir, "store_test.txt"), testContent, 0o644)

	compEngine := compress.New()
	_, err = compEngine.Pack(types.PackOptions{
		SourcePaths:      []string{srcDir},
		OutputZipPath:    zipPath,
		Method:           types.MethodStore,
		CompressionLevel: 0,
		Overwrite:        true,
	})
	if err != nil {
		t.Fatalf("Pack in Store mode failed: %v", err)
	}

	extractEngine := New()
	summary, err := extractEngine.Unpack(types.UnpackOptions{
		ArchivePath:    zipPath,
		DestinationDir: destDir,
		Overwrite:      true,
	})
	if err != nil {
		t.Fatalf("Unpack of Store archive failed: %v", err)
	}

	if summary.TotalFiles < 1 {
		t.Errorf("Expected at least 1 file extracted, got %d", summary.TotalFiles)
	}

	extractedFile := filepath.Join(destDir, "source", "store_test.txt")
	gotContent, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(gotContent) != string(testContent) {
		t.Errorf("Content mismatch: expected %q, got %q", string(testContent), string(gotContent))
	}
}

func TestUnpackZstdMode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "extract_zstd_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "extracted")
	zipPath := filepath.Join(tempDir, "zstd.zip")

	_ = os.MkdirAll(srcDir, 0o755)
	testContent := []byte("Zstandard Method 93 extraction test payload data with lots of repeating tokens 123456789")
	_ = os.WriteFile(filepath.Join(srcDir, "zstd_test.txt"), testContent, 0o644)

	compEngine := compress.New()
	_, err = compEngine.Pack(types.PackOptions{
		SourcePaths:      []string{srcDir},
		OutputZipPath:    zipPath,
		Method:           types.MethodZstd,
		CompressionLevel: 3,
		Overwrite:        true,
	})
	if err != nil {
		t.Fatalf("Pack in Zstandard mode failed: %v", err)
	}

	extractEngine := New()
	summary, err := extractEngine.Unpack(types.UnpackOptions{
		ArchivePath:    zipPath,
		DestinationDir: destDir,
		Overwrite:      true,
	})
	if err != nil {
		t.Fatalf("Unpack of Zstandard archive failed: %v", err)
	}

	if summary.TotalFiles < 1 {
		t.Errorf("Expected at least 1 file extracted, got %d", summary.TotalFiles)
	}

	extractedFile := filepath.Join(destDir, "source", "zstd_test.txt")
	gotContent, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Failed to read extracted zstd file: %v", err)
	}
	if string(gotContent) != string(testContent) {
		t.Errorf("Content mismatch: expected %q, got %q", string(testContent), string(gotContent))
	}

	// Test InspectArchive on Zstd zip
	inspectSummary, err := extractEngine.InspectArchive(zipPath)
	if err != nil {
		t.Fatalf("InspectArchive on Zstd zip failed: %v", err)
	}
	if inspectSummary.CompressionMethod != "Zstandard" {
		t.Errorf("Expected inspect summary CompressionMethod 'Zstandard', got %q", inspectSummary.CompressionMethod)
	}
}
