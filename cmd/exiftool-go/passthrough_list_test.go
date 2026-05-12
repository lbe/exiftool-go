package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestPassthroughListNonEmpty(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("exiftool"); err != nil && strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")) == "" {
		t.Skip("exiftool not on PATH and EXIFTOOL_GO_EXIFTOOL unset")
	}
	bin := buildEntrypointBinary(t)
	out, err := exec.Command(bin, "--backend=native", "-list").Output()
	if err != nil {
		t.Fatalf("-list: %v", err)
	}
	if len(strings.TrimSpace(string(out))) < 20 {
		t.Fatalf("expected substantial -list output, got %q", string(out))
	}
}
