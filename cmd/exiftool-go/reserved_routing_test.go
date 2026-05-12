package main

import (
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/lbe/exiftool-go/internal/backend"
)

func TestParseBackendLiteralInvalidValue(t *testing.T) {
	t.Helper()
	_, err := backend.ParseBackendLiteral("nope")
	if err == nil {
		t.Fatal("expected error for invalid backend")
	}
	if !strings.Contains(err.Error(), "invalid --backend value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStripLeadingBackendFlagsMissingValue(t *testing.T) {
	t.Helper()
	_, _, err := backend.StripLeadingBackendFlags([]string{"--backend"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--backend requires a value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStripLeadingBackendFlagsValidChoice(t *testing.T) {
	t.Helper()
	t.Run("equals_form_native", func(t *testing.T) {
		t.Helper()
		rest, kind, err := backend.StripLeadingBackendFlags([]string{"--backend=native", "x"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kind != backend.BackendNative {
			t.Fatalf("kind: got %v, want BackendNative", kind)
		}
		if got, wantRest := rest, []string{"x"}; !reflect.DeepEqual(got, wantRest) {
			t.Fatalf("rest: got %#v, want %#v", got, wantRest)
		}
	})

	t.Run("two_token_form_wasm", func(t *testing.T) {
		t.Helper()
		rest, kind, err := backend.StripLeadingBackendFlags([]string{"--backend", "wasm", "-ver"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kind != backend.BackendWasm {
			t.Fatalf("kind: got %v, want BackendWasm", kind)
		}
		if got, wantRest := rest, []string{"-ver"}; !reflect.DeepEqual(got, wantRest) {
			t.Fatalf("rest: got %#v, want %#v", got, wantRest)
		}
	})
}

func TestStripLeadingBackendFlagsStripsOnlyLeadingReserved(t *testing.T) {
	t.Helper()
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "single_eq_then_rest",
			in:   []string{"--backend=native", "-json", "a.jpg", "b.jpg"},
			want: []string{"-json", "a.jpg", "b.jpg"},
		},
		{
			name: "two_tokens_then_rest",
			in:   []string{"--backend", "wasm", "pipeline", "-l", "INFO"},
			want: []string{"pipeline", "-l", "INFO"},
		},
		{
			name: "double_leading_backend",
			in:   []string{"--backend=native", "--backend=wasm", "-w", "3"},
			want: []string{"-w", "3"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got, _, err := backend.StripLeadingBackendFlags(tc.in)
			if err != nil {
				t.Fatalf("StripLeadingBackendFlags: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestPipelineSubcommandArgvAfterPeelReachesParseFlags(t *testing.T) {
	t.Helper()
	rest, _, err := backend.StripLeadingBackendFlags([]string{"--backend=native", "pipeline", "-l", "DEBUG", "/tmp/in"})
	if err != nil {
		t.Fatalf("StripLeadingBackendFlags: %v", err)
	}
	if len(rest) < 1 || rest[0] != "pipeline" {
		t.Fatalf("expected pipeline subcommand first, got %#v", rest)
	}
	cfg, err := parseFlags(rest[1:])
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if cfg.logLevel != "DEBUG" {
		t.Fatalf("logLevel: got %q, want DEBUG", cfg.logLevel)
	}
	if len(cfg.dirs) != 1 || cfg.dirs[0] != "/tmp/in" {
		t.Fatalf("dirs: got %#v, want [/tmp/in]", cfg.dirs)
	}
}

func TestWrapperHelpAndVersionNotExiftoolPassthrough(t *testing.T) {
	t.Helper()
	bin := buildEntrypointBinary(t)

	t.Run("help_long", func(t *testing.T) {
		t.Helper()
		out, err := exec.Command(bin, "--help").Output()
		if err != nil {
			t.Fatalf("--help: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "Reserved before forward") || !strings.Contains(s, "usage: exiftool-go") {
			t.Fatalf("expected wrapper global usage; got %q", s)
		}
	})

	t.Run("help_short", func(t *testing.T) {
		t.Helper()
		out, err := exec.Command(bin, "-help").Output()
		if err != nil {
			t.Fatalf("-help: %v", err)
		}
		if !strings.Contains(string(out), "usage: exiftool-go") {
			t.Fatalf("expected wrapper usage; got %q", string(out))
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

func TestPassthroughForwardsArgsInOrderAfterBackendPeel(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("exiftool"); err != nil && strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")) == "" {
		t.Skip("exiftool not on PATH and EXIFTOOL_GO_EXIFTOOL unset")
	}
	bin := buildEntrypointBinary(t)

	cmd := exec.Command(bin, "--backend=native", "-ver")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("wrapper: %v", err)
	}
	cmd2 := exec.Command("exiftool", "-ver")
	if p := strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")); p != "" {
		cmd2 = exec.Command(p, "-ver")
	}
	wantOut, err := cmd2.Output()
	if err != nil {
		t.Fatalf("driver -ver: %v", err)
	}
	if string(out) != string(wantOut) {
		t.Fatalf("stdout mismatch after --backend peel\nwrapper: %q\ndirect:  %q", string(out), string(wantOut))
	}
}
