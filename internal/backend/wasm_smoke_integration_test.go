//go:build integration

package backend

import (
	"context"
	"regexp"
	"testing"
)

func TestWasmDriverVerSmoke(t *testing.T) {
	t.Helper()
	var d WasmDriver
	out, err := d.CommandContext(context.Background(), nil, "-ver")
	if err != nil {
		t.Fatalf("CommandContext -ver: %v", err)
	}
	if !regexp.MustCompile(`\d+\.\d+`).MatchString(string(out)) {
		t.Fatalf("unexpected -ver output: %q", string(out))
	}
}
