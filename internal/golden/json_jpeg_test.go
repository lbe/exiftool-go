package golden

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExiftoolDriver_JSON_samsungJPEG_matchesNormalizedGolden(t *testing.T) {
	root := RepoRoot(t)
	img := filepath.Join(root, "testdata", "images", "jpeg", "samsung_sm_g930p.jpg")
	if _, err := os.Stat(img); err != nil {
		t.Fatalf("jpeg fixture: %v", err)
	}

	goldenPath := filepath.Join(root, "testdata", "golden", "json_samsung_sm_g930p.json")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	ex := DriverExiftool(t)
	relImg := "testdata/images/jpeg/samsung_sm_g930p.jpg"
	cmd := exec.Command(ex, "-json", "-q", relImg)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "TZ=UTC", "LC_ALL=C")
	got, err := cmd.Output()
	if err != nil {
		t.Fatalf("exiftool -json: %v", err)
	}

	gotNorm, err := NormalizeExiftoolJSONGolden(got)
	if err != nil {
		t.Fatalf("normalize actual: %v", err)
	}
	wantNorm, err := NormalizeExiftoolJSONGolden(want)
	if err != nil {
		t.Fatalf("normalize golden: %v", err)
	}
	if !bytes.Equal(gotNorm, wantNorm) {
		t.Fatalf("normalized JSON mismatch\n--- got ---\n%s\n--- want ---\n%s", string(gotNorm), string(wantNorm))
	}
}
