package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCmdPackAndUnpackE2E(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cli_e2e_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "sample_dir")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("Antigravity CLI Compressor Test"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "file2.txt"), []byte("Second file with some repeated repeated content content"), 0o644)

	zipOut := filepath.Join(tempDir, "packed.zip")
	destUnpack := filepath.Join(tempDir, "unpacked")

	// 1. Test Pack
	RootCmd.SetArgs([]string{"pack", srcDir, "-o", zipOut, "-l", "9", "-f", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("RootCmd pack failed: %v", err)
	}

	if _, err := os.Stat(zipOut); os.IsNotExist(err) {
		t.Fatalf("Packed zip file does not exist on disk")
	}

	// 2. Test List
	RootCmd.SetArgs([]string{"list", zipOut, "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("RootCmd list failed: %v", err)
	}

	// 3. Test Bench
	RootCmd.SetArgs([]string{"bench", srcDir})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("RootCmd bench failed: %v", err)
	}

	// 4. Test Unpack
	RootCmd.SetArgs([]string{"unpack", zipOut, "-o", destUnpack, "-f", "--quiet"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("RootCmd unpack failed: %v", err)
	}

	extractedFile := filepath.Join(destUnpack, "sample_dir", "file1.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Fatalf("Extracted file %q does not exist", extractedFile)
	}
}
