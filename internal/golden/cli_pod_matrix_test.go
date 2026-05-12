package golden

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// podMatrixDualBackendCases adds §5.1–§5.6 Pod-matrix scenarios (see docs/Code-Review-Remdiation-1-Plan.md Phase 1).
func podMatrixDualBackendCases() []dualBackendCase {
	return []dualBackendCase{
		// §5.1 — case rules (ExifTool accepts `-JSON` as `-json`).
		{
			name:       "pod_case_insensitive_JSON_flag",
			argv:       []string{"-JSON", "-q", "testdata/images/jpeg/samsung_sm_g930p.jpg"},
			goldenFile: "testdata/golden/json_samsung_sm_g930p.json",
			normalizer: NormalizeExiftoolJSONGolden,
		},
		// §5.1 — options may follow the source file (Pod §OPTIONS).
		{
			name:       "pod_json_options_after_source_file",
			argv:       []string{"testdata/images/jpeg/samsung_sm_g930p.jpg", "-json", "-q"},
			goldenFile: "testdata/golden/json_samsung_sm_g930p.json",
			normalizer: NormalizeExiftoolJSONGolden,
		},
		// §5.1 — `-sss` is one argv token (distinct from three `-s` flags in both behavior and forwarding tests).
		{
			name:       "pod_short_cluster_sss_Model_one_argv",
			argv:       []string{"-sss", "-Model", "testdata/images/jpeg/samsung_sm_g930p.jpg"},
			goldenFile: "testdata/golden/pod_sss_model_one_argv.stdout",
		},
		// §5.1 — unknown `-Foo` forwarded as a custom tag request (empty when absent).
		{
			name:       "pod_unknown_minus_Foo_tag_token",
			argv:       []string{"-Foo", "testdata/tree/alpha/a.jpg"},
			goldenFile: "testdata/golden/pod_foo_unknown.stdout",
		},
		// §5.1 — `--` end of options; path after `--` still read in order.
		{
			name:       "pod_json_end_of_options_separator",
			argv:       []string{"-json", "-q", "--", "testdata/tree/alpha/a.jpg"},
			goldenFile: "testdata/golden/pod_json_endopts_alpha.json",
			normalizer: NormalizeExiftoolJSONGolden,
		},
		// §5.1 — single-token `-ver` passthrough (baseline alongside clustered options).
		{
			name:       "pod_ver_single_token",
			argv:       []string{"-ver"},
			goldenFile: "testdata/golden/version.stdout",
		},
		// §5.2 — reading `-Model` with short print stack.
		{
			name:       "pod_read_Model_short_values",
			argv:       []string{"-s", "-s", "-s", "-Model", "testdata/images/jpeg/samsung_sm_g930p.jpg"},
			goldenFile: "testdata/golden/pod_read_model_short.stdout",
		},
		// §5.2 — `-All` text on a tiny fixture (volatile FS lines stripped; see NormalizeExiftoolTagListText).
		{
			name:       "pod_all_text_alpha_fixture",
			argv:       []string{"-q", "-All", "testdata/tree/alpha/a.jpg"},
			goldenFile: "testdata/golden/pod_all_alpha_text.stdout",
			normalizer: NormalizeExiftoolTagListText,
		},
		// §5.4 — `-X` XML with volatile System:* lines removed.
		{
			name:       "pod_xml_alpha_fixture",
			argv:       []string{"-X", "-q", "testdata/tree/alpha/a.jpg"},
			goldenFile: "testdata/golden/pod_xml_alpha.stdout",
			normalizer: NormalizeExiftoolXMLGolden,
		},
		// §5.5 — `-T` / `-S` short table and labelled short.
		{
			name:       "pod_format_table_Model",
			argv:       []string{"-T", "-Model", "testdata/images/jpeg/samsung_sm_g930p.jpg"},
			goldenFile: "testdata/golden/pod_table_model.stdout",
		},
		{
			name:       "pod_format_short_S_Model",
			argv:       []string{"-S", "-Model", "testdata/images/jpeg/samsung_sm_g930p.jpg"},
			goldenFile: "testdata/golden/pod_short_capital_S_model.stdout",
		},
		// §5.5 — `-ext` / `-ext+` on testdata/tree (sorted lines; see FilenameLinesStable).
		{
			name:       "pod_ext_jpg_tree_filenames",
			argv:       []string{"-r", "-ext", "jpg", "-T", "-Filename", "testdata/tree"},
			goldenFile: "testdata/golden/pod_ext_jpg_tree.stdout",
			normalizer: normalizeSortedFilenameTable,
		},
		{
			name:       "pod_ext_plus_jpg_tree_sorted",
			argv:       []string{"-r", "-ext+", "JPG", "-T", "-Filename", "testdata/tree"},
			goldenFile: "testdata/golden/pod_ext_plus_jpg_tree.stdout",
			normalizer: normalizeSortedFilenameTable,
		},
		// §5.5 — `-if` smoke (deterministic true condition on known dimensions).
		{
			name:       "pod_if_imagewidth_prints_filename",
			argv:       []string{"-q", "-if", "$imagewidth", "-p", "$filename", "testdata/tree/alpha/a.jpg"},
			goldenFile: "testdata/golden/pod_if_imagewidth.stdout",
		},
		// §5.6 — list capability breadth (prefix compare; full `-listx` output is huge).
		{
			name:       "pod_listw_head_prefix",
			argv:       []string{"-listw"},
			goldenFile: "testdata/golden/pod_listw_head.stdout",
			normalizer: func(b []byte) ([]byte, error) {
				return NormalizeExiftoolHeadBytes(b, 35)
			},
		},
		{
			name:       "pod_listx_head_prefix",
			argv:       []string{"-listx"},
			goldenFile: "testdata/golden/pod_listx_head.stdout",
			normalizer: func(b []byte) ([]byte, error) {
				return NormalizeExiftoolHeadBytes(b, 28)
			},
		},
	}
}

func normalizeSortedFilenameTable(b []byte) ([]byte, error) {
	return []byte(FilenameLinesStable(string(b))), nil
}

// TestExiftoolGoCLI_podMatrix_tagsFromFile verifies §5.2 copying with temp-only fixtures (native and wasm backends).
func TestExiftoolGoCLI_podMatrix_tagsFromFileModel(t *testing.T) {
	root := RepoRoot(t)
	requireNativeExiftoolDriverOrSkip(t)
	unified := buildCLIUnderTest(t, root)

	want, err := os.ReadFile(filepath.Join(root, "testdata/golden/pod_tagsfromfile_model.stdout"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	wantNorm, err := normalizeOutput(bytes.TrimSpace(want), nil)
	if err != nil {
		t.Fatalf("normalize golden: %v", err)
	}

	backends := []parityBackend{
		{name: "native", bin: unified, backendArg: "native"},
		{name: "guest", bin: unified, backendArg: "wasm"},
	}

	for _, backend := range backends {
		backend := backend
		t.Run(backend.name, func(t *testing.T) {
			workDir := t.TempDir()
			src := filepath.Join(workDir, "src.jpg")
			dst := filepath.Join(workDir, "dst.jpg")
			seedTempFromFixture(t, filepath.Join(root, "testdata/images/jpeg/samsung_sm_g930p.jpg"), src)
			seedTempFromFixture(t, filepath.Join(root, "testdata/tree/alpha/a.jpg"), dst)

			mutArgs := append([]string{"--backend=" + backend.backendArg}, "-q", "-overwrite_original", "-tagsFromFile", "src.jpg", "dst.jpg")
			runCLIBinary(t, workDir, backend.bin, nil, mutArgs)

			readArgs := append([]string{"--backend=" + backend.backendArg}, "-q", "-s", "-s", "-s", "-Model", "dst.jpg")
			out := runCLIBinary(t, workDir, backend.bin, nil, readArgs)
			gotNorm, err := normalizeOutput(bytes.TrimSpace(out), nil)
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}
			assertNormalizedMatchesGolden(t, fmt.Sprintf("tagsFromFile_Model_%s", backend.name), gotNorm, wantNorm)
		})
	}
}
