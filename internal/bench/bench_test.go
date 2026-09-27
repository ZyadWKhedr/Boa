package bench

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestWorkspace(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bench_test_workspace_*")
	if err != nil {
		t.Fatalf("failed to create temp workspace: %v", err)
	}

	_ = os.WriteFile(filepath.Join(dir, "doc.txt"), []byte("Repeated text line for compression testing.\nRepeated text line for compression testing.\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "data.json"), []byte(`{"key": "value", "items": [1,2,3,4,5,6,7,8,9,10]}`), 0o644)

	subDir := filepath.Join(dir, "nested")
	_ = os.MkdirAll(subDir, 0o755)
	_ = os.WriteFile(filepath.Join(subDir, "code.go"), []byte("package main\n\nfunc main() {\n\t// comment\n}\n"), 0o644)

	return dir
}

func TestBenchmarkAllLevels(t *testing.T) {
	ws := createTestWorkspace(t)
	defer os.RemoveAll(ws)

	runner := NewRunner()
	opts := Options{
		SourcePath:      ws,
		Level:           -1, // all levels 0-9
		Runs:            1,
		BenchmarkDecomp: true,
		KeepArchives:    false,
	}

	report, err := runner.Run(opts)
	if err != nil {
		t.Fatalf("benchmark failed: %v", err)
	}

	if len(report.Results) != 10 {
		t.Fatalf("expected 10 levels (0-9), got %d", len(report.Results))
	}

	for i, res := range report.Results {
		if res.Level != i {
			t.Errorf("expected level %d, got %d", i, res.Level)
		}
		if res.CompressedBytes <= 0 {
			t.Errorf("level %d: expected compressed bytes > 0, got %d", i, res.CompressedBytes)
		}
		if res.DecompDurationAvg <= 0 {
			t.Errorf("level %d: expected decomp duration > 0, got %v", i, res.DecompDurationAvg)
		}
	}

	// Verify Summary Highlights
	if report.Summary.SmallestSize == "" {
		t.Errorf("expected non-empty SmallestSize summary")
	}
	if report.Summary.FastestSpeed == "" {
		t.Errorf("expected non-empty FastestSpeed summary")
	}
}

func TestBenchmarkSpecificLevel(t *testing.T) {
	ws := createTestWorkspace(t)
	defer os.RemoveAll(ws)

	runner := NewRunner()
	levelsToTest := []int{0, 1, 6, 9}

	for _, lvl := range levelsToTest {
		t.Run(GetLevelName(lvl), func(t *testing.T) {
			opts := Options{
				SourcePath: ws,
				Level:      lvl,
				Runs:       2,
			}
			report, err := runner.Run(opts)
			if err != nil {
				t.Fatalf("benchmark level %d failed: %v", lvl, err)
			}
			if len(report.Results) != 1 {
				t.Fatalf("expected 1 level result, got %d", len(report.Results))
			}
			if report.Results[0].Level != lvl {
				t.Errorf("expected level %d, got %d", lvl, report.Results[0].Level)
			}
			if report.RunsPerLevel != 2 {
				t.Errorf("expected RunsPerLevel 2, got %d", report.RunsPerLevel)
			}
		})
	}
}

func TestBenchmarkJSONAndCSVRender(t *testing.T) {
	ws := createTestWorkspace(t)
	defer os.RemoveAll(ws)

	runner := NewRunner()
	opts := Options{
		SourcePath:      ws,
		Level:           6,
		Runs:            1,
		BenchmarkDecomp: true,
	}

	report, err := runner.Run(opts)
	if err != nil {
		t.Fatalf("runner.Run failed: %v", err)
	}

	// 1. Test JSON Serialization
	var jsonBuf bytes.Buffer
	if err := RenderJSON(report, &jsonBuf); err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var parsedReport Report
	if err := json.Unmarshal(jsonBuf.Bytes(), &parsedReport); err != nil {
		t.Fatalf("failed to unmarshal rendered JSON: %v", err)
	}

	if parsedReport.Input.TotalFiles != report.Input.TotalFiles {
		t.Errorf("JSON mismatch: total_files %d != %d", parsedReport.Input.TotalFiles, report.Input.TotalFiles)
	}

	// 2. Test CSV Serialization
	var csvBuf bytes.Buffer
	if err := RenderCSV(report, &csvBuf); err != nil {
		t.Fatalf("RenderCSV failed: %v", err)
	}

	csvStr := csvBuf.String()
	if !strings.Contains(csvStr, "level,duration_ms,compressed_size_bytes") {
		t.Errorf("CSV missing header: %s", csvStr)
	}
	if !strings.Contains(csvStr, "decomp_duration_ms") {
		t.Errorf("CSV missing decomp header: %s", csvStr)
	}
}

func TestBenchmarkKeepArchives(t *testing.T) {
	ws := createTestWorkspace(t)
	defer os.RemoveAll(ws)

	runner := NewRunner()
	opts := Options{
		SourcePath:   ws,
		Level:        6,
		Runs:         1,
		KeepArchives: true,
	}

	report, err := runner.Run(opts)
	if err != nil {
		t.Fatalf("runner.Run failed: %v", err)
	}
	defer os.RemoveAll(report.PreservedDir)

	if report.PreservedDir == "" {
		t.Fatal("expected PreservedDir to be set when KeepArchives is true")
	}

	if _, err := os.Stat(report.PreservedDir); os.IsNotExist(err) {
		t.Errorf("preserved dir %q does not exist", report.PreservedDir)
	}
}

func TestBenchmarkZeroAndSmallInput(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bench_empty_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	emptyFile := filepath.Join(tempDir, "empty.txt")
	_ = os.WriteFile(emptyFile, []byte(""), 0o644)

	runner := NewRunner()
	opts := Options{
		SourcePath: emptyFile,
		Level:      6,
	}

	report, err := runner.Run(opts)
	if err != nil {
		t.Fatalf("benchmark on empty file failed: %v", err)
	}

	if !report.IsSmallDataset {
		t.Errorf("expected IsSmallDataset to be true for 0-byte file")
	}
}

func TestBenchmarkInvalidSource(t *testing.T) {
	runner := NewRunner()
	opts := Options{
		SourcePath: "/path/to/nonexistent/directory/xyz123",
	}

	_, err := runner.Run(opts)
	if err == nil {
		t.Fatal("expected error for nonexistent source path, got nil")
	}
}
