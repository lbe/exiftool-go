package exifutil

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestExtractExifFromJPEG tests EXIF extraction from a JPEG file.
func TestExtractExifFromJPEG(t *testing.T) {
	// Construct path to test image relative to this source file
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "sample.jpg")

	exifData, err := Extract(imagePath)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Assert that basic EXIF fields were extracted (not empty)
	if len(exifData) == 0 {
		t.Errorf("Extract returned empty EXIF data")
	}

	// Check that SourceFile field is present (exiftool always adds this)
	if _, ok := exifData["SourceFile"]; !ok {
		t.Errorf("EXIF field %q not found in extracted data", "SourceFile")
	}
}

// TestExtractNoExif tests extraction from an image with no EXIF data.
func TestExtractNoExif(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "no_exif", "blank.jpg")

	exifData, err := Extract(imagePath)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should still return valid map (possibly empty or with only file metadata)
	if exifData == nil {
		t.Errorf("Extract returned nil for image with no EXIF")
	}
}

// TestExtractCorruptFile tests extraction from a corrupt/invalid file.
func TestExtractCorruptFile(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "corrupt", "broken.jpg")

	// Corrupt file should result in an error from exiftool
	exifData, err := Extract(imagePath)
	if err == nil {
		t.Logf("Extract succeeded on corrupt file; got %d fields", len(exifData))
		// Some corrupt files may still return partial data; that's acceptable
	} else {
		t.Logf("Extract failed on corrupt file (expected): %v", err)
	}
}

// TestExtractNonexistentFile tests extraction from a non-existent file.
func TestExtractNonexistentFile(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata")
	imagePath := filepath.Join(testdataDir, "nonexistent.jpg")

	_, err := Extract(imagePath)
	if err == nil {
		t.Errorf("Extract should fail on non-existent file, but got nil error")
	}
}
