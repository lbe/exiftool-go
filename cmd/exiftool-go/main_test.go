package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseFlagsShortLogLevel(t *testing.T) {
	t.Helper()

	cfg, err := parseFlags([]string{"-l", "DEBUG"})
	if err != nil {
		t.Fatalf("parseFlags returned unexpected error: %v", err)
	}
	if got, want := cfg.logLevel, "DEBUG"; got != want {
		t.Fatalf("logLevel: got %q, want %q", got, want)
	}
}

func TestParseFlagsLongLogLevel(t *testing.T) {
	t.Helper()

	cfg, err := parseFlags([]string{"--log-level", "WARN"})
	if err != nil {
		t.Fatalf("parseFlags returned unexpected error: %v", err)
	}
	if got, want := cfg.logLevel, "WARN"; got != want {
		t.Fatalf("logLevel: got %q, want %q", got, want)
	}
}

func TestParseFlagsWorkerCount(t *testing.T) {
	t.Helper()

	cfg, err := parseFlags([]string{"-w", "4"})
	if err != nil {
		t.Fatalf("parseFlags returned unexpected error: %v", err)
	}
	if got, want := cfg.workers, 4; got != want {
		t.Fatalf("workers: got %d, want %d", got, want)
	}
}

func TestParseFlagsInvalidLogLevel(t *testing.T) {
	t.Helper()

	_, err := parseFlags([]string{"--log-level", "TRACE"})
	if err == nil {
		t.Fatal("parseFlags invalid log level: expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid log level") {
		t.Fatalf("invalid log level error: got %q", err)
	}
}

func TestParseFlagsInvalidWorkerCount(t *testing.T) {
	t.Helper()

	_, err := parseFlags([]string{"-w", "0"})
	if err == nil {
		t.Fatal("parseFlags invalid worker count: expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid worker") {
		t.Fatalf("invalid worker count error: got %q", err)
	}
}

func TestEntrypointBinaryBuilds(t *testing.T) {
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

	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("built binary missing: %v", err)
	}
}

func TestEntrypointNoInputReturnsError(t *testing.T) {
	t.Helper()

	binaryPath := buildEntrypointBinary(t)

	cmd := exec.Command(binaryPath)
	cmd.Stdin = strings.NewReader("")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("entrypoint without input: expected non-zero exit, got nil error")
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("entrypoint error type: got %T, want *exec.ExitError", err)
	}

	errText := strings.ToLower(stderr.String())
	if !strings.Contains(errText, "no directories provided") {
		t.Fatalf("stderr missing expected message: %q", stderr.String())
	}
}

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
