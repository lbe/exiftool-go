package main

import (
	"testing"

	"github.com/lbe/exiftool-go/internal/backend"
)

func TestPipelineSubcommandAfterBackendPeel(t *testing.T) {
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
