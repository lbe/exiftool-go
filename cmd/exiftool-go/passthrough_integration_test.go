//go:build integration

package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPassthrough covers wrapper passthrough contracts not exercised by golden parity:
// native stdout matches the driver on success, and subprocess failures surface exit
// code and stderr (see main.exitError).
func TestPassthrough(t *testing.T) {
	t.Helper()
	skipIfNoExiftoolDriver(t)

	bin := buildEntrypointBinary(t)
	driver := exiftoolDriverPath(t)

	t.Run("native_stdout_matches_driver", func(t *testing.T) {
		t.Helper()

		out, err := exec.Command(bin, "--backend=native", "-ver").Output()
		if err != nil {
			t.Fatalf("wrapper -ver: %v", err)
		}

		wantOut, err := exec.Command(driver, "-ver").Output()
		if err != nil {
			t.Fatalf("driver -ver: %v", err)
		}
		if string(out) != string(wantOut) {
			t.Fatalf("stdout mismatch\nwrapper: %q\ndriver:  %q", string(out), string(wantOut))
		}
	})

	t.Run("failure_exit_and_stderr", func(t *testing.T) {
		t.Helper()

		missing := filepath.Join(t.TempDir(), "absent.jpg")

		var driverStderr bytes.Buffer
		cmdDriver := exec.Command(driver, "-q", missing)
		cmdDriver.Stderr = &driverStderr
		driverErr := cmdDriver.Run()
		if driverErr == nil {
			t.Fatalf("driver: expected failure for missing file, got nil (stderr=%q)", driverStderr.String())
		}

		var wrapStderr bytes.Buffer
		cmdWrap := exec.Command(bin, "--backend=native", "-q", missing)
		cmdWrap.Stderr = &wrapStderr
		wrapErr := cmdWrap.Run()
		if wrapErr == nil {
			t.Fatalf("wrapper: expected failure for missing file, got nil (stderr=%q)", wrapStderr.String())
		}

		if got, want := exitCode(t, wrapErr), exitCode(t, driverErr); got != want {
			t.Fatalf("exit code: got %d want %d (wrapper stderr=%q driver stderr=%q)", got, want, wrapStderr.String(), driverStderr.String())
		}

		wrapStderrText := wrapStderr.String()
		if strings.TrimSpace(wrapStderrText) == "" {
			t.Fatalf("wrapper stderr empty; driver stderr was %q", driverStderr.String())
		}

		if marker := stderrMarker(driverStderr.String()); marker != "" && !strings.Contains(wrapStderrText, marker) {
			t.Fatalf("wrapper stderr should include driver error text; want substring %q; got wrapper=%q driver=%q", marker, wrapStderrText, driverStderr.String())
		}
	})
}
