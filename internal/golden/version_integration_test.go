//go:build integration

package golden

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExiftoolDriver_ver_stdout_matchesGolden(t *testing.T) {
	root := RepoRoot(t)
	goldenPath := filepath.Join(root, "testdata", "golden", "version.stdout")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenPath, err)
	}

	ex := DriverExiftool(t)
	cmd := exec.Command(ex, "-ver")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "TZ=UTC", "LC_ALL=C")
	got, err := cmd.Output()
	if err != nil {
		t.Fatalf("run exiftool -ver: %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Fatalf("stdout mismatch\ngot:  %q\nwant: %q", string(got), string(want))
	}
}
