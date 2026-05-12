package golden

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lbe/exiftool-go/internal/backend"
)

type dualBackendCase struct {
	name       string
	argv       []string
	stdin      []byte
	goldenFile string
	normalizer func([]byte) ([]byte, error)
	skipReason string
	// workDir is Cmd.Dir for exiftool-go. Empty means the repository root.
	workDir string
	// skipGuestReason, when non-empty, skips the wasm backend subtest after native vs golden passes
	// (e.g. guest does not yet implement -ext traversal parity for this scenario).
	skipGuestReason string
}

type parityBackend struct {
	name       string
	bin        string
	backendArg string
}

func TestExiftoolGoCLI_dualBackendParity(t *testing.T) {
	root := RepoRoot(t)
	requireNativeExiftoolDriverOrSkip(t)
	unified := buildCLIUnderTest(t, root)

	testCases := append([]dualBackendCase{
		{
			name:       "json_samsung_sm_g930p",
			argv:       []string{"-json", "-q", "testdata/images/jpeg/samsung_sm_g930p.jpg"},
			goldenFile: "testdata/golden/json_samsung_sm_g930p.json",
			normalizer: NormalizeExiftoolJSONGolden,
		},
		{
			name:       "api_userparam_ver_passthrough",
			argv:       []string{"-ver", "-api", "RequestAll=3", "-userParam", "Foo=Bar"},
			goldenFile: "testdata/golden/api_userparam_ver.stdout",
		},
		{
			name:       "execute_sequence_ver",
			argv:       []string{"-ver", "-execute", "-ver"},
			goldenFile: "testdata/golden/execute_sequence_ver.stdout",
		},
		{
			name: "phase_g_r8_argfile_mixed_formats",
			argv: []string{
				"-@", "testdata/cli/argfiles/r8_mixed_formats.args",
				"-json", "-q",
			},
			goldenFile: "testdata/golden/phase_g_r8_argfile.stdout",
			normalizer: NormalizeExiftoolJSONGolden,
		},
	}, podMatrixDualBackendCases()...)

	backends := struct {
		native parityBackend
		guest  parityBackend
	}{
		native: parityBackend{name: "native", bin: unified, backendArg: "native"},
		guest:  parityBackend{name: "guest", bin: unified, backendArg: "wasm"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipReason != "" {
				t.Skip(tc.skipReason)
			}
			passthrough, err := stripReservedBackendArgs(tc.argv)
			if err != nil {
				t.Fatalf("strip reserved backend args for %q: %v", tc.name, err)
			}

			want, err := os.ReadFile(filepath.Join(root, tc.goldenFile))
			if err != nil {
				t.Fatalf("read golden %q: %v", tc.goldenFile, err)
			}

			cmdDir := root
			if tc.workDir != "" {
				cmdDir = tc.workDir
			}

			nativeNorm := runNormalizedBackendOutput(t, cmdDir, backends.native, tc.stdin, passthrough, tc.normalizer)

			wantNorm, err := normalizeOutput(want, tc.normalizer)
			if err != nil {
				t.Fatalf("normalize golden output: %v", err)
			}

			assertNormalizedMatchesGolden(t, tc.name, nativeNorm, wantNorm)

			if tc.skipGuestReason != "" {
				t.Run("guest_backend", func(t *testing.T) {
					t.Skip(tc.skipGuestReason)
				})
				return
			}

			guestNorm := runNormalizedBackendOutput(t, cmdDir, backends.guest, tc.stdin, passthrough, tc.normalizer)
			assertNormalizedBackendParity(t, tc.name, nativeNorm, guestNorm)
		})
	}
}

func assertNormalizedBackendParity(t *testing.T, testName string, nativeNorm, guestNorm []byte) {
	t.Helper()
	if !bytes.Equal(nativeNorm, guestNorm) {
		t.Fatalf(
			"normalized backend mismatch for %q\n--- native ---\n%s\n--- guest ---\n%s",
			testName,
			string(nativeNorm),
			string(guestNorm),
		)
	}
}

func assertNormalizedMatchesGolden(t *testing.T, testName string, gotNorm, wantNorm []byte) {
	t.Helper()
	if !bytes.Equal(gotNorm, wantNorm) {
		t.Fatalf(
			"normalized output mismatch vs golden for %q\n--- got ---\n%s\n--- want ---\n%s",
			testName,
			string(gotNorm),
			string(wantNorm),
		)
	}
}

func requireNativeExiftoolDriverOrSkip(t *testing.T) {
	t.Helper()
	if p := strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")); p != "" {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			t.Skipf("skipping: EXIFTOOL_GO_EXIFTOOL %q unavailable: %v", p, err)
		}
		return
	}
	if _, err := exec.LookPath("exiftool"); err != nil {
		t.Skipf("skipping: exiftool not on PATH (set EXIFTOOL_GO_EXIFTOOL): %v", err)
	}
}

func buildCLIUnderTest(t *testing.T, root string) string {
	t.Helper()
	tmpDir := t.TempDir()
	bin := filepath.Join(tmpDir, "exiftool-go")
	buildCLI(t, root, bin)
	return bin
}

func buildCLI(t *testing.T, root, output string, buildArgs ...string) {
	t.Helper()
	args := []string{"build"}
	args = append(args, buildArgs...)
	args = append(args, "-o", output, "./cmd/exiftool-go")
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

func runCLIBinary(t *testing.T, workDir, bin string, stdin []byte, args []string) []byte {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "TZ=UTC", "LC_ALL=C")
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run %q with args %v: %v\n%s", bin, args, err, string(out))
	}
	return out
}

func runNormalizedBackendOutput(
	t *testing.T,
	workDir string,
	backend parityBackend,
	stdin []byte,
	passthrough []string,
	normalizer func([]byte) ([]byte, error),
) []byte {
	t.Helper()
	args := append([]string{"--backend=" + backend.backendArg}, passthrough...)
	out := runCLIBinary(t, workDir, backend.bin, stdin, args)
	norm, err := normalizeOutput(out, normalizer)
	if err != nil {
		t.Fatalf("normalize %s output: %v", backend.name, err)
	}
	return norm
}

func stripReservedBackendArgs(argv []string) ([]string, error) {
	rest := append([]string(nil), argv...)
	for len(rest) > 0 {
		a := rest[0]
		switch {
		case strings.HasPrefix(a, "--backend="):
			if err := validateBackendFlagValue(strings.TrimPrefix(a, "--backend=")); err != nil {
				return nil, err
			}
			rest = rest[1:]
		case a == "--backend":
			if len(rest) < 2 {
				return nil, fmt.Errorf("--backend requires a value")
			}
			if err := validateBackendFlagValue(rest[1]); err != nil {
				return nil, err
			}
			rest = rest[2:]
		default:
			return rest, nil
		}
	}
	return rest, nil
}

func validateBackendFlagValue(v string) error {
	_, err := backend.ParseBackendLiteral(v)
	return err
}

func normalizeOutput(in []byte, normalizer func([]byte) ([]byte, error)) ([]byte, error) {
	if normalizer == nil {
		return bytes.TrimSpace(in), nil
	}
	return normalizer(in)
}
