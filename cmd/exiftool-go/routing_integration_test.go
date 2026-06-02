//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPassthroughDoesNotCreateIngestDatabase(t *testing.T) {
	t.Helper()

	repoRoot := repoRoot(t)
	bin := buildEntrypointBinary(t)
	xdgDataHome := t.TempDir()
	fixture := filepath.Join(repoRoot, "testdata", "images", "jpeg", "samsung_sm_g930p.jpg")

	cmd := exec.Command(bin, "-json", "-q", fixture)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "XDG_DATA_HOME="+xdgDataHome)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("passthrough -json: %v", err)
	}
	if !strings.Contains(string(out), "samsung_sm_g930p.jpg") {
		t.Fatalf("expected JSON for fixture, got: %q", string(out))
	}

	dbPath := filepath.Join(xdgDataHome, "exiftool-go", "exif.db")
	if _, err := os.Stat(dbPath); err == nil {
		t.Fatalf("passthrough must not create ingest database at %s", dbPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat ingest db: %v", err)
	}
}

func TestWrapperReservedArgv(t *testing.T) {
	t.Helper()

	bin := buildEntrypointBinary(t)

	t.Run("help", func(t *testing.T) {
		t.Helper()
		for _, flag := range []string{"--help", "-help"} {
			out, err := exec.Command(bin, flag).Output()
			if err != nil {
				t.Fatalf("%s: %v", flag, err)
			}
			s := string(out)
			if !strings.Contains(s, "usage: exiftool-go") {
				t.Fatalf("%s: expected wrapper usage; got %q", flag, s)
			}
		}
		longOut, err := exec.Command(bin, "--help").Output()
		if err != nil {
			t.Fatalf("--help: %v", err)
		}
		if !strings.Contains(string(longOut), "Default behavior forwards argv") {
			t.Fatalf("--help: expected passthrough note; got %q", string(longOut))
		}
	})

	t.Run("version", func(t *testing.T) {
		t.Helper()
		out, err := exec.Command(bin, "--version").Output()
		if err != nil {
			t.Fatalf("--version: %v", err)
		}
		if strings.TrimSpace(string(out)) != "exiftool-go (development build)" {
			t.Fatalf("unexpected stdout: %q", string(out))
		}
	})
}
