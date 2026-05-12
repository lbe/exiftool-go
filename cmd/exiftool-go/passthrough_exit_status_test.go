package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Phase 2 (docs/Code-Review-Remdiation-1-Plan.md §6.6): passthrough must exit
// with non-zero status classes, forward subprocess stderr, and still exit 0 on
// success — subprocess driver is PATH exiftool or EXIFTOOL_GO_EXIFTOOL (see
// internal/backend/native.go).

func TestPassthroughMissingFileExitMatchesExiftoolDriver(t *testing.T) {
	t.Helper()
	skipIfNoExiftoolDriver(t)

	bin := buildEntrypointBinary(t)
	missing := filepath.Join(t.TempDir(), "absent-phase-t.jpg")

	driver := exiftoolDriverPath(t)

	// -q routes argv away from legacy ingest (single bare path is treated as pipeline dirs).
	cmdW := exec.Command(bin, "--backend=native", "-q", missing)
	var wstderr bytes.Buffer
	cmdW.Stderr = &wstderr
	wrapErr := cmdW.Run()

	cmdE := exec.Command(driver, "-q", missing)
	var estderr bytes.Buffer
	cmdE.Stderr = &estderr
	exErr := cmdE.Run()

	if exErr == nil {
		t.Fatalf("baseline exiftool: expected failure for missing file, got nil (stderr=%q)", estderr.String())
	}
	if wrapErr == nil {
		t.Fatalf("wrapper: expected failure for missing file, got nil (stderr=%q)", wstderr.String())
	}

	want := exitCode(t, exErr)
	got := exitCode(t, wrapErr)
	if got != want {
		t.Fatalf("exit code: got %d want %d (wrapper stderr=%q driver stderr=%q)", got, want, wstderr.String(), estderr.String())
	}
}

func TestPassthroughMissingFileStderrSurfaced(t *testing.T) {
	t.Helper()
	skipIfNoExiftoolDriver(t)

	bin := buildEntrypointBinary(t)
	missing := filepath.Join(t.TempDir(), "missing-for-stderr-phase-t.jpg")
	driver := exiftoolDriverPath(t)

	var estderr bytes.Buffer
	cmdE := exec.Command(driver, "-q", missing)
	cmdE.Stderr = &estderr
	if err := cmdE.Run(); err == nil {
		t.Fatal("baseline exiftool: expected failure for missing file")
	}
	driverStderr := estderr.String()

	var wstderr bytes.Buffer
	cmdW := exec.Command(bin, "--backend=native", "-q", missing)
	cmdW.Stderr = &wstderr
	if err := cmdW.Run(); err == nil {
		t.Fatal("wrapper: expected failure for missing file")
	}
	wrapStderr := wstderr.String()
	if strings.TrimSpace(wrapStderr) == "" {
		t.Fatalf("wrapper stderr empty; expected subprocess stderr from exitError (driver stderr was %q)", driverStderr)
	}

	marker := stderrMarker(driverStderr)
	if marker != "" && !strings.Contains(wrapStderr, marker) {
		t.Fatalf("wrapper stderr should include driver error text; want substring %q; got wrapper=%q driver=%q", marker, wrapStderr, driverStderr)
	}
}

func TestPassthroughVerExitZeroAfterFailureCase(t *testing.T) {
	t.Helper()
	skipIfNoExiftoolDriver(t)

	bin := buildEntrypointBinary(t)
	cmd := exec.Command(bin, "--backend=native", "-ver")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("wrapper -ver: %v stderr=%q", err, stderr.String())
	}
	if len(bytes.TrimSpace(out)) == 0 {
		t.Fatalf("expected non-empty stdout from -ver")
	}
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
