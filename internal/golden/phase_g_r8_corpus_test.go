package golden

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	r8FixturePNGPath      = "testdata/images/png/sample.png"
	r8FixtureTIFFPath     = "testdata/images/tiff/sample.tiff"
	r8FixtureWEBPPath     = "testdata/images/webp/sample.webp"
	r8ArgfilePath         = "testdata/cli/argfiles/r8_mixed_formats.args"
	r8ArgfileGoldenStdout = "testdata/golden/phase_g_r8_argfile.stdout"
	r8HEICFixtureDir      = "testdata/images/heic"
	r8RAWFixtureDir       = "testdata/images/raw"
)

func TestPhaseG_R8_requiredCorpusArtifactsExist(t *testing.T) {
	root := RepoRoot(t)

	required := []string{
		r8FixturePNGPath,
		r8FixtureTIFFPath,
		r8FixtureWEBPPath,
		r8ArgfilePath,
		r8ArgfileGoldenStdout,
	}

	for _, rel := range required {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			requireReadableRegularFile(t, root, rel)
		})
	}
}

func TestPhaseG_R8_optionalHEICCorpusRowPolicy(t *testing.T) {
	root := RepoRoot(t)
	requireOptionalCorpusRowOrSkip(
		t,
		root,
		r8HEICFixtureDir,
		[]string{".heic", ".heif"},
		"§6.3 optional row `testdata/images/heic/`",
	)
}

func TestPhaseG_R8_optionalRAWCorpusRowPolicy(t *testing.T) {
	root := RepoRoot(t)
	requireOptionalCorpusRowOrSkip(
		t,
		root,
		r8RAWFixtureDir,
		[]string{".cr2", ".cr3", ".dng", ".nef", ".arw", ".raf", ".rw2", ".orf"},
		"§6.3 optional row `testdata/images/raw/`",
	)
}

func requireReadableRegularFile(t *testing.T, root, rel string) {
	t.Helper()

	path := filepath.Join(root, rel)
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected planned Phase G R8 fixture %q is missing: %v", path, err)
	}
	if !st.Mode().IsRegular() {
		t.Fatalf("expected planned Phase G R8 fixture %q to be a regular file (mode=%s)", path, st.Mode())
	}
	if _, err := os.ReadFile(path); err != nil {
		t.Fatalf("expected planned Phase G R8 fixture %q to be readable: %v", path, err)
	}
}

func requireOptionalCorpusRowOrSkip(t *testing.T, root, relDir string, exts []string, row string) {
	t.Helper()

	dirPath := filepath.Join(root, relDir)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("skipping: %s absent; %s allows explicit skip when no sample is available", relDir, row)
		}
		t.Fatalf("read optional corpus directory %q for %s: %v", dirPath, row, err)
	}

	matches := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if stringSliceContains(exts, ext) {
			matches = append(matches, filepath.Join(relDir, entry.Name()))
		}
	}
	if len(matches) == 0 {
		t.Skipf("skipping: %s has no sample files matching %v; %s permits explicit skip when sample is unavailable", relDir, exts, row)
	}

	for _, rel := range matches {
		requireReadableRegularFile(t, root, rel)
	}
}

func stringSliceContains(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}
