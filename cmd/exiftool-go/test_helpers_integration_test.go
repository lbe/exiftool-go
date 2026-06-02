//go:build integration

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildEntrypointBinary(t *testing.T) string {
	t.Helper()

	repoRoot := repoRoot(t)
	buildDir := t.TempDir()
	binaryPath := filepath.Join(buildDir, "exiftool-go-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/exiftool-go")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	return binaryPath
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func skipIfNoExiftoolDriver(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("exiftool"); err != nil && strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")) == "" {
		t.Skip("exiftool not on PATH and EXIFTOOL_GO_EXIFTOOL unset")
	}
}

func exiftoolDriverPath(t *testing.T) string {
	t.Helper()
	if p := strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")); p != "" {
		return p
	}
	p, err := exec.LookPath("exiftool")
	if err != nil {
		t.Skip("exiftool not on PATH and EXIFTOOL_GO_EXIFTOOL unset")
	}
	return p
}

func exitCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	return ee.ExitCode()
}

func stderrMarker(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}
