package exifutil

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestExtractExifFromJPEG tests EXIF extraction from a JPEG file.
func TestExtractExifFromJPEG(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "sample.jpg")

	exifData, err := Extract(imagePath)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(exifData) == 0 {
		t.Errorf("Extract returned empty EXIF data")
	}

	if _, ok := exifData["SourceFile"]; !ok {
		t.Errorf("EXIF field %q not found in extracted data", "SourceFile")
	}
}

func TestExtractNoExif(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "no_exif", "blank.jpg")

	exifData, err := Extract(imagePath)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if exifData == nil {
		t.Errorf("Extract returned nil for image with no EXIF")
	}
}

func TestExtractCorruptFile(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "corrupt", "broken.jpg")

	exifData, err := Extract(imagePath)
	if err == nil {
		t.Logf("Extract succeeded on corrupt file; got %d fields", len(exifData))
	} else {
		t.Logf("Extract failed on corrupt file (expected): %v", err)
	}
}

func TestExtractNonexistentFile(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "nonexistent.jpg")

	_, err := Extract(imagePath)
	if err == nil {
		t.Errorf("Extract should fail on non-existent file, but got nil error")
	}
}
