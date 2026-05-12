package golden

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type t3BackendExpectation struct {
	supported  bool
	skipReason string
}

type t3WriteScenario struct {
	name               string
	argv               []string
	expectOriginalCopy bool
	goldenFile         string
	backendExpect      map[string]t3BackendExpectation
}

func TestExiftoolGoCLI_t3WriteOriginalUtility(t *testing.T) {
	root := RepoRoot(t)
	requireNativeExiftoolDriverOrSkip(t)
	unified := buildCLIUnderTest(t, root)

	workDir := t.TempDir()
	input := filepath.Join(workDir, "write_original_input.jpg")
	seedTempFromFixture(t, filepath.Join(root, "testdata", "images", "jpeg", "samsung_sm_g930p.jpg"), input)

	scenarios := []t3WriteScenario{
		{
			name:               "write_comment_creates_original_backup",
			argv:               []string{"-Comment=phase-g-t3-red", input},
			expectOriginalCopy: true,
			goldenFile:         "testdata/golden/phase_g_t3_write_comment.stdout",
			backendExpect: map[string]t3BackendExpectation{
				"native": {supported: true},
				"guest":  {supported: true},
			},
		},
	}

	backends := []parityBackend{
		{name: "native", bin: unified, backendArg: "native"},
		{name: "guest", bin: unified, backendArg: "wasm"},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		for _, backend := range backends {
			backend := backend
			t.Run(fmt.Sprintf("%s_%s", scenario.name, backend.name), func(t *testing.T) {
				expect, ok := scenario.backendExpect[backend.name]
				if !ok {
					t.Fatalf("missing backend expectation metadata for %q", backend.name)
				}
				if !expect.supported {
					t.Skip(expect.skipReason)
				}

				args := append([]string{"--backend=" + backend.backendArg}, scenario.argv...)
				out := runCLIBinary(t, root, backend.bin, nil, args)

				if scenario.expectOriginalCopy {
					backup := input + "_original"
					if _, err := os.Stat(backup); err != nil {
						t.Fatalf("expected original backup artifact %q: %v", backup, err)
					}
				}

				goldenPath := filepath.Join(root, scenario.goldenFile)
				want, err := os.ReadFile(goldenPath)
				if err != nil {
					t.Fatalf("expected planned T3 golden artifact %q is missing: %v", goldenPath, err)
				}
				if !bytes.Equal(bytes.TrimSpace(out), bytes.TrimSpace(want)) {
					t.Fatalf("stdout mismatch\n--- got ---\n%s\n--- want ---\n%s", string(out), string(want))
				}
			})
		}
	}
}

func seedTempFromFixture(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture %q: %v", src, err)
	}
	if err := os.WriteFile(dst, b, 0o600); err != nil {
		t.Fatalf("write temp fixture copy %q: %v", dst, err)
	}
}
