package golden

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// RepoRoot returns the repository root (directory containing go.mod).
func RepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// DriverExiftool returns the ExifTool driver path: EXIFTOOL_GO_EXIFTOOL if set,
// otherwise exiftool from PATH.
func DriverExiftool(t *testing.T) string {
	t.Helper()
	if p := strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")); p != "" {
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			t.Fatalf("EXIFTOOL_GO_EXIFTOOL %q: %v", p, err)
		}
		return p
	}
	path, err := exec.LookPath("exiftool")
	if err != nil {
		t.Fatalf("exiftool not on PATH (set EXIFTOOL_GO_EXIFTOOL): %v", err)
	}
	return path
}
