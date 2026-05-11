package scanner

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestScanDirFindsImagesRecursively(t *testing.T) {
	t.Helper()

	root := filepath.Join("..", "..", "testdata")

	got, err := ScanDir(root)
	if err != nil {
		t.Fatalf("ScanDir returned error: %v", err)
	}

	for index := range got {
		got[index] = filepath.ToSlash(got[index])
	}

	want := []string{
		filepath.ToSlash(filepath.Join(root, "Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif.jpg")),
		filepath.ToSlash(filepath.Join(root, "base.jpg")),
		filepath.ToSlash(filepath.Join(root, "corrupt", "broken.jpg")),
		filepath.ToSlash(filepath.Join(root, "no_exif", "blank.jpg")),
		filepath.ToSlash(filepath.Join(root, "sample.jpg")),
		filepath.ToSlash(filepath.Join(root, "sample.png")),
		filepath.ToSlash(filepath.Join(root, "sample.tiff")),
		filepath.ToSlash(filepath.Join(root, "sample_gps.jpg")),
		filepath.ToSlash(filepath.Join(root, "sample_noexif.jpg")),
		filepath.ToSlash(filepath.Join(root, "test.jpg")),
		filepath.ToSlash(filepath.Join(root, "with_exif", "nested", "another.jpg")),
		filepath.ToSlash(filepath.Join(root, "with_exif", "photo.jpg")),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ScanDir mismatch:\ngot:  %v\nwant: %v", got, want)
	}
}
