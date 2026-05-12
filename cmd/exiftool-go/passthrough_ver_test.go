package main

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// Phase D (plan §7): argv-first passthrough must forward ExifTool-native flags
// such as -ver to the configured driver (PATH or EXIFTOOL_GO_EXIFTOOL).
func TestPassthroughForwardsExiftoolVer(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("exiftool"); err != nil && strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")) == "" {
		t.Skip("exiftool not on PATH and EXIFTOOL_GO_EXIFTOOL unset")
	}

	bin := buildEntrypointBinary(t)
	cmd := exec.Command(bin, "-ver")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("exiftool-go -ver: %v", err)
	}
	if !regexp.MustCompile(`\d+\.\d+`).MatchString(string(out)) {
		t.Fatalf("expected version-like stdout from passthrough -ver, got %q", string(out))
	}
}
