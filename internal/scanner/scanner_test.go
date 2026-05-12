package scanner

import (
	"path/filepath"
	"slices"
	"testing"
)

// mustContainImages are paths relative to testdata root that ScanDir must
// always discover; the corpus may grow (new dirs/files) without breaking this test.
var mustContainImages = []string{
	"Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif.jpg",
	"base.jpg",
	"corrupt/broken.jpg",
	"no_exif/blank.jpg",
	"sample.jpg",
	"sample.png",
	"sample.tiff",
	"sample_gps.jpg",
	"sample_noexif.jpg",
	"test.jpg",
	"with_exif/nested/another.jpg",
	"with_exif/photo.jpg",
	"images/jpeg/samsung_sm_g930p.jpg",
	"images/jpeg/samsung_sm_g930p_altpath.jpg",
	"tree/alpha/a.jpg",
	"tree/alpha/nested/b.jpg",
	"tree/bravo/c.jpg",
}

func TestScanDirFindsImagesRecursively(t *testing.T) {
	t.Helper()

	root := filepath.Join("..", "..", "testdata")

	got, err := ScanDir(root)
	if err != nil {
		t.Fatalf("ScanDir returned error: %v", err)
	}

	for i := range got {
		got[i] = filepath.ToSlash(filepath.Clean(got[i]))
	}
	slices.Sort(got)

	for _, rel := range mustContainImages {
		want := filepath.ToSlash(filepath.Clean(filepath.Join(root, rel)))
		if !slices.Contains(got, want) {
			t.Fatalf("ScanDir missing required image %q (under %q)", rel, root)
		}
	}

	if len(got) < len(mustContainImages) {
		t.Fatalf("ScanDir returned %d paths, want at least %d", len(got), len(mustContainImages))
	}
}
