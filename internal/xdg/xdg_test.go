package xdg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataDirUsesXDGDataHome(t *testing.T) {
	t.Helper()

	t.Setenv("HOME", filepath.Join(t.TempDir(), "home"))
	baseDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", baseDir)

	got, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir returned error: %v", err)
	}

	want := filepath.Join(baseDir, "exiftool-go")
	if got != want {
		t.Fatalf("DataDir: got %q, want %q", got, want)
	}

	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("DataDir path is not a directory: %q", got)
	}
}

func TestDataDirFallsBackToHome(t *testing.T) {
	t.Helper()

	homeDir := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll home: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_DATA_HOME", "")

	got, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir returned error: %v", err)
	}

	want := filepath.Join(homeDir, ".local", "share", "exiftool-go")
	if got != want {
		t.Fatalf("DataDir: got %q, want %q", got, want)
	}

	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("DataDir path is not a directory: %q", got)
	}
}
