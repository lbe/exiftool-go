//go:build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmZeroByteFileJSONPassthrough(t *testing.T) {
	t.Helper()
	skipIfNoExiftoolDriver(t)

	bin := buildEntrypointBinary(t)
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.jpg")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "-json", empty)
	out, err := cmd.Output()
	if len(out) == 0 {
		t.Fatal("expected JSON on stdout for zero-byte file")
	}
	if !strings.Contains(string(out), "File is empty") {
		t.Fatalf("stdout missing Error field text: %s", out)
	}

	var records []map[string]any
	if err := json.Unmarshal(out, &records); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}
	if code := exitCode(t, err); code != 1 {
		t.Fatalf("exit code: got %d want 1", code)
	}
}
