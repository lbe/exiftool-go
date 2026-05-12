package main

import (
	"testing"

	"github.com/lbe/exiftool-go/internal/backend"
)

func TestParseBackendLiteralWasmAlwaysValid(t *testing.T) {
	t.Helper()
	if _, err := backend.ParseBackendLiteral("wasm"); err != nil {
		t.Fatalf("wasm must always be valid in unified binary: %v", err)
	}
}

func TestParseBackendLiteralNativeAlwaysValid(t *testing.T) {
	t.Helper()
	if _, err := backend.ParseBackendLiteral("native"); err != nil {
		t.Fatalf("native must always be valid in unified binary: %v", err)
	}
}
