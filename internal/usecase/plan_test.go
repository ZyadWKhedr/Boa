package usecase

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"compressor/internal/domain"
	"compressor/internal/infrastructure/classifier"
)

func TestPlanAndEstimatorWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create a sample dataset
	_ = os.WriteFile(filepath.Join(tempDir, "doc.pdf"), []byte("%PDF-1.4 sample pdf document"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "code.go"), []byte("package main\n\nfunc main() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "photo.jpg"), []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}, 0o644)

	cls := classifier.NewMagicClassifier()
	scanUC := NewScanPathUseCase(cls)
	policy := domain.NewLossyEligibilityPolicy()
	estimator := NewEstimateSavingsUseCase()
	planUC := NewBuildCompressionPlanUseCase(policy, estimator)

	ctx := context.Background()

	// 2. Test ScanPath
	scanRes, err := scanUC.Execute(ctx, tempDir, nil)
	if err != nil {
		t.Fatalf("ScanPath failed: %v", err)
	}

	if scanRes.TotalFiles != 3 {
		t.Errorf("Expected 3 scanned files, got %d", scanRes.TotalFiles)
	}
	if scanRes.Images.Count != 1 {
		t.Errorf("Expected 1 image, got %d", scanRes.Images.Count)
	}
	if scanRes.Other.Count != 2 {
		t.Errorf("Expected 2 other files, got %d", scanRes.Other.Count)
	}

	// 3. Test Lossless Plan (Default)
	losslessPrefs := domain.UserPreferences{
		ArchiveMethod: domain.MethodDeflate,
		DefaultLevel:  6,
		Images:        domain.CategoryPreferences{EnableLossy: false},
	}

	losslessPlan, err := planUC.Execute(ctx, scanRes, losslessPrefs, "out.zip")
	if err != nil {
		t.Fatalf("BuildCompressionPlan (Lossless) failed: %v", err)
	}

	if losslessPlan.HasLossyMedia {
		t.Errorf("Lossless plan unexpectedly marked HasLossyMedia = true")
	}
	if losslessPlan.LossyFileCount != 0 {
		t.Errorf("Expected 0 lossy files in lossless plan, got %d", losslessPlan.LossyFileCount)
	}

	// 4. Test Lossy Images Plan
	lossyPrefs := domain.UserPreferences{
		ArchiveMethod: domain.MethodDeflate,
		DefaultLevel:  6,
		Images: domain.CategoryPreferences{
			EnableLossy: true,
			Quality:     80,
		},
	}

	lossyPlan, err := planUC.Execute(ctx, scanRes, lossyPrefs, "out.zip")
	if err != nil {
		t.Fatalf("BuildCompressionPlan (Lossy) failed: %v", err)
	}

	if !lossyPlan.HasLossyMedia {
		t.Errorf("Expected HasLossyMedia = true for lossy image plan")
	}
	if lossyPlan.LossyFileCount != 1 {
		t.Errorf("Expected 1 lossy file, got %d", lossyPlan.LossyFileCount)
	}

	// 5. Verify documents/code were NOT marked lossy even in lossy mode
	for _, f := range lossyPlan.Files {
		if f.Entry.RelPath == "doc.pdf" || f.Entry.RelPath == "code.go" {
			if f.IsLossy || f.Action.Kind == domain.ActionTranscode {
				t.Errorf("Document/code file %q was illegally assigned lossy action: %v", f.Entry.RelPath, f.Action)
			}
		}
	}
}
