//go:build integration

package golden

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExiftoolDriver_r_recursive_tree_filenames_matchGolden(t *testing.T) {
	root := RepoRoot(t)
	treeDir := filepath.Join(root, "testdata", "tree")
	if _, err := os.Stat(treeDir); err != nil {
		t.Fatalf("testdata/tree: %v", err)
	}

	goldenPath := filepath.Join(root, "testdata", "golden", "tree_r_filename_sorted.stdout")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	ex := DriverExiftool(t)
	cmd := exec.Command(ex, "-r", "-T", "-Filename", "testdata/tree")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "TZ=UTC", "LC_ALL=C")
	got, err := cmd.Output()
	if err != nil {
		t.Fatalf("exiftool -r: %v", err)
	}

	gotNorm := FilenameLinesStable(string(got))
	wantNorm := FilenameLinesStable(string(want))
	if !bytes.Equal([]byte(gotNorm), []byte(wantNorm)) {
		t.Fatalf("sorted filename list mismatch\n--- got ---\n%s\n--- want ---\n%s", gotNorm, wantNorm)
	}
}
