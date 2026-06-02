//go:build integration

package golden

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhaseG_R8_minimumJPEGCamerasDistinct(t *testing.T) {
	root := RepoRoot(t)
	ex := DriverExiftool(t)

	fixtures := []string{
		"testdata/images/jpeg/samsung_sm_g930p.jpg",
		"testdata/sample.jpg",
	}

	signatures := make(map[string]struct{}, len(fixtures))
	for _, rel := range fixtures {
		sig := cameraSignatureFromFixture(t, root, ex, rel)
		signatures[sig] = struct{}{}
	}
	if len(signatures) < 2 {
		t.Fatalf("expected at least 2 distinct camera signatures across minimum JPEG corpus; got %d (%v)", len(signatures), fixtures)
	}
}

func cameraSignatureFromFixture(t *testing.T, root, exiftoolPath, rel string) string {
	t.Helper()
	abs := filepath.Join(root, rel)
	out, err := exec.Command(exiftoolPath, "-s3", "-Make", "-Model", abs).CombinedOutput()
	if err != nil {
		t.Fatalf("read make/model from %q: %v\n%s", rel, err, string(out))
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected make/model lines for %q, got %q", rel, string(out))
	}
	return strings.TrimSpace(lines[0]) + "::" + strings.TrimSpace(lines[1])
}
