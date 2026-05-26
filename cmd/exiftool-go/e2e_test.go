package main

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lbe/exiftool-go/internal/testsupport"

	_ "modernc.org/sqlite"
)

var e2eBinaryPath string

const richEXIFFixtureBaseName = "Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif.jpg"

func TestMain(m *testing.M) {
	os.Exit(runE2ETestMain(m))
}

func runE2ETestMain(m *testing.M) int {
	repoRoot, err := repoRootFromCaller()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo root: %v\n", err)
		return 1
	}

	buildDir, err := os.MkdirTemp("", "exiftool-go-e2e-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create e2e temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(buildDir)

	e2eBinaryPath = filepath.Join(buildDir, "exiftool-go-test")
	defer func() {
		if e2eBinaryPath != "" {
			_ = os.Remove(e2eBinaryPath)
		}
	}()

	cmd := exec.Command("go", "build", "-o", e2eBinaryPath, "./cmd/exiftool-go")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build e2e binary failed: %v\n%s", err, string(out))
		return 1
	}

	return m.Run()
}

func repoRootFromCaller() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..")), nil
}

func xdgDatabasePath(xdgDataHome string) string {
	return filepath.Join(xdgDataHome, "exiftool-go", "exif.db")
}

func runE2EBinary(t *testing.T, xdgDataHome string, stdin string, args ...string) []byte {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	cmd := exec.Command(e2eBinaryPath, args...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "XDG_DATA_HOME="+xdgDataHome)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary run failed: %v\n%s", err, string(out))
	}
	return out
}

func runE2EBinaryExpectFailure(t *testing.T, xdgDataHome string, stdin string, args ...string) []byte {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	cmd := exec.Command(e2eBinaryPath, args...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "XDG_DATA_HOME="+xdgDataHome)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit from binary")
	}
	return out
}

func openSQLiteDB(t *testing.T, dbPath string) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})
	return database
}

func TestE2EBinaryBuilt(t *testing.T) {
	t.Helper()

	if e2eBinaryPath == "" {
		t.Fatal("e2eBinaryPath should be set by TestMain")
	}
	info, err := os.Stat(e2eBinaryPath)
	if err != nil {
		t.Fatalf("stat e2e binary: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("e2e binary path is a directory: %s", e2eBinaryPath)
	}
}

func TestE2ECLIArgs(t *testing.T) {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	xdgDataHome := t.TempDir()
	withExifDir := createRichFixtureInputDir(t, repoRoot)

	_ = runE2EBinary(t, xdgDataHome, "", withExifDir)

	dbPath := xdgDatabasePath(xdgDataHome)
	assertWithExifRows(t, dbPath, repoRoot)
}

func TestE2EPipelineSubcommandIsolation(t *testing.T) {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	xdgDataHome := t.TempDir()
	withExifDir := createRichFixtureInputDir(t, repoRoot)

	cmd := exec.Command(e2eBinaryPath, "pipeline", withExifDir)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "XDG_DATA_HOME="+xdgDataHome)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected successful ingest via pipeline subcommand: %v\n%s", err, string(out))
	}
	combined := strings.ToLower(string(out))
	if strings.Contains(combined, "dir=pipeline") || strings.Contains(combined, "root=pipeline") {
		t.Fatalf("pipeline must be a subcommand, not a scan root; output:\n%s", string(out))
	}

	dbPath := xdgDatabasePath(xdgDataHome)
	assertWithExifRows(t, dbPath, repoRoot)
}

func TestE2EStdinPipe(t *testing.T) {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	xdgDataHome := t.TempDir()
	withExifDir := createRichFixtureInputDir(t, repoRoot)

	_ = runE2EBinary(t, xdgDataHome, withExifDir+"\n")

	dbPath := xdgDatabasePath(xdgDataHome)
	assertWithExifRows(t, dbPath, repoRoot)
}

func TestE2ENoInputTTY(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("script utility not available for pseudo-terminal test")
	}

	xdgDataHome := t.TempDir()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("script", "-q", "-c", e2eBinaryPath, "/dev/null")
	default:
		// BSD/macOS script: script [-aq] [file] [command ...]
		cmd = exec.Command("script", "-q", "/dev/null", e2eBinaryPath)
	}
	cmd.Env = append(os.Environ(), "XDG_DATA_HOME="+xdgDataHome)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit when no input is provided on TTY")
	}

	errText := strings.ToLower(string(out))
	if !strings.Contains(errText, "no directories provided") {
		t.Fatalf("expected no directories message, got: %q", string(out))
	}
}

func TestE2ENoInputEmptyPipe(t *testing.T) {
	t.Helper()

	xdgDataHome := t.TempDir()

	out := runE2EBinaryExpectFailure(t, xdgDataHome, "\n")
	errText := strings.ToLower(string(out))
	if !strings.Contains(errText, "no directories provided") {
		t.Fatalf("expected no directories message, got: %q", string(out))
	}
}

func TestE2ENonExistentDirectory(t *testing.T) {
	t.Helper()

	xdgDataHome := t.TempDir()
	missingDir := filepath.Join(t.TempDir(), "does-not-exist")

	out := runE2EBinaryExpectFailure(t, xdgDataHome, "", missingDir)
	errText := strings.ToLower(string(out))
	if !strings.Contains(errText, "no such file or directory") {
		t.Fatalf("expected missing directory error, got: %q", string(out))
	}
}

func TestE2ENoEXIFDirectory(t *testing.T) {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	xdgDataHome := t.TempDir()
	noExifDir := filepath.Join(repoRoot, "testdata", "no_exif")

	_ = runE2EBinary(t, xdgDataHome, "", noExifDir)

	dbPath := xdgDatabasePath(xdgDataHome)
	assertNoExifRows(t, dbPath)
}

func TestE2ECorruptImages(t *testing.T) {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	xdgDataHome := t.TempDir()
	corruptDir := filepath.Join(repoRoot, "testdata", "corrupt")

	_ = runE2EBinary(t, xdgDataHome, "", corruptDir)

	dbPath := xdgDatabasePath(xdgDataHome)
	assertCorruptRows(t, dbPath)
}

func TestE2ELogLevelFlag(t *testing.T) {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	withExifDir := createRichFixtureInputDir(t, repoRoot)

	t.Run("debug flag enables debug logs", func(t *testing.T) {
		xdgDataHome := t.TempDir()

		out := runE2EBinary(t, xdgDataHome, "", "-l", "DEBUG", withExifDir)
		if !strings.Contains(string(out), "DEBUG") {
			t.Fatalf("expected DEBUG logs when -l DEBUG is set, got: %q", string(out))
		}
	})

	t.Run("default level suppresses debug logs", func(t *testing.T) {
		xdgDataHome := t.TempDir()

		out := runE2EBinary(t, xdgDataHome, "", withExifDir)
		if strings.Contains(string(out), "DEBUG") {
			t.Fatalf("did not expect DEBUG logs at default level, got: %q", string(out))
		}
	})
}

func TestE2EWorkerFlag(t *testing.T) {
	t.Helper()

	repoRoot, err := repoRootFromCaller()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	withExifDir := createRichFixtureInputDir(t, repoRoot)

	t.Run("short -w sets worker count", func(t *testing.T) {
		xdgDataHome := t.TempDir()

		out := runE2EBinary(t, xdgDataHome, "", "-w", "1", withExifDir)
		if !strings.Contains(string(out), "workers=1") {
			t.Fatalf("expected workers=1 in output, got: %q", string(out))
		}

		dbPath := xdgDatabasePath(xdgDataHome)
		assertWithExifRows(t, dbPath, repoRoot)
	})

	t.Run("long --workers sets worker count", func(t *testing.T) {
		xdgDataHome := t.TempDir()

		out := runE2EBinary(t, xdgDataHome, "", "--workers", "2", withExifDir)
		if !strings.Contains(string(out), "workers=2") {
			t.Fatalf("expected workers=2 in output, got: %q", string(out))
		}

		dbPath := xdgDatabasePath(xdgDataHome)
		assertWithExifRows(t, dbPath, repoRoot)
	})
}

func assertWithExifRows(t *testing.T, dbPath string, repoRoot string) {
	t.Helper()

	database := openSQLiteDB(t, dbPath)
	t.Cleanup(func() {
		_ = database.Close()
	})

	type row struct {
		path         string
		make         sql.NullString
		model        sql.NullString
		dateTime     sql.NullString
		gpsLatitude  sql.NullFloat64
		gpsLongitude sql.NullFloat64
		imageWidth   sql.NullInt64
		imageHeight  sql.NullInt64
		orientation  sql.NullInt64
		exposureTime sql.NullString
		fNumber      sql.NullString
		isoSpeed     sql.NullInt64
		focalLength  sql.NullString
		status       string
		lastError    sql.NullString
		exifJSON     string
	}

	rows, err := database.Query(`
SELECT file_path, camera_make, camera_model, date_time, gps_latitude, gps_longitude, image_width, image_height, orientation, exposure_time, f_number, iso_speed, focal_length, exif_status, last_error, exif_json
FROM images
ORDER BY file_path
`)
	if err != nil {
		t.Fatalf("query rows: %v", err)
	}
	t.Cleanup(func() {
		_ = rows.Close()
	})

	gotRows := make([]row, 0)
	for rows.Next() {
		var r row
		if err = rows.Scan(&r.path, &r.make, &r.model, &r.dateTime, &r.gpsLatitude, &r.gpsLongitude, &r.imageWidth, &r.imageHeight, &r.orientation, &r.exposureTime, &r.fNumber, &r.isoSpeed, &r.focalLength, &r.status, &r.lastError, &r.exifJSON); err != nil {
			t.Fatalf("scan row: %v", err)
		}
		gotRows = append(gotRows, r)
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}

	if len(gotRows) != 1 {
		t.Fatalf("row count: got %d, want 1", len(gotRows))
	}

	expected := richExifExpectedFromManifest(t, repoRoot)
	r := gotRows[0]

	if !filepath.IsAbs(r.path) {
		t.Fatalf("file_path is not absolute: %q", r.path)
	}
	if clean := filepath.Clean(r.path); clean != r.path {
		t.Fatalf("file_path is not canonical: got %q, clean %q", r.path, clean)
	}

	base := filepath.Base(r.path)
	if base != richEXIFFixtureBaseName {
		t.Fatalf("unexpected file in DB: %q", base)
	}

	if !r.make.Valid || r.make.String != expected.Make {
		t.Fatalf("camera_make for %s: got %v, want %q", base, r.make, expected.Make)
	}
	if !r.model.Valid || r.model.String != expected.Model {
		t.Fatalf("camera_model for %s: got %v, want %q", base, r.model, expected.Model)
	}
	parsed := testsupport.ParseEXIFJSON(t, r.exifJSON)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"DateTimeOriginal"}, "date_time", r.dateTime.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"GPSLatitude"}, "gps_latitude", r.gpsLatitude.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"GPSLongitude"}, "gps_longitude", r.gpsLongitude.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ImageWidth", "ExifImageWidth"}, "image_width", r.imageWidth.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ImageHeight", "ExifImageHeight"}, "image_height", r.imageHeight.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"Orientation"}, "orientation", r.orientation.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ExposureTime"}, "exposure_time", r.exposureTime.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"FNumber"}, "f_number", r.fNumber.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ISO", "ISOSpeedRatings"}, "iso_speed", r.isoSpeed.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"FocalLength"}, "focal_length", r.focalLength.Valid)

	if !r.dateTime.Valid {
		t.Fatalf("date_time for %s should be valid", base)
	}
	storedDate, err := time.Parse(time.RFC3339, r.dateTime.String)
	if err != nil {
		t.Fatalf("parse stored date_time for %s: %v", base, err)
	}
	expectedDate, err := time.Parse("2006:01:02 15:04:05", expected.DateTimeOriginal)
	if err != nil {
		t.Fatalf("parse expected DateTimeOriginal for %s: %v", base, err)
	}
	if !storedDate.Equal(expectedDate.UTC()) {
		t.Fatalf("date_time for %s: got %s, want %s", base, storedDate.Format(time.RFC3339), expectedDate.UTC().Format(time.RFC3339))
	}
	if !r.gpsLatitude.Valid {
		t.Fatalf("gps_latitude for %s should be valid", base)
	}
	if !r.gpsLongitude.Valid {
		t.Fatalf("gps_longitude for %s should be valid", base)
	}
	if got, want := int(r.imageWidth.Int64), expected.ImageWidth; got != want {
		t.Fatalf("image_width for %s: got %d, want %d", base, got, want)
	}
	if got, want := int(r.imageHeight.Int64), expected.ImageHeight; got != want {
		t.Fatalf("image_height for %s: got %d, want %d", base, got, want)
	}
	if !r.exposureTime.Valid || r.exposureTime.String != expected.ExposureTime {
		t.Fatalf("exposure_time for %s: got %v, want %q", base, r.exposureTime, expected.ExposureTime)
	}
	if !r.fNumber.Valid || r.fNumber.String != expected.FNumber {
		t.Fatalf("f_number for %s: got %v, want %q", base, r.fNumber, expected.FNumber)
	}
	if got, want := int(r.isoSpeed.Int64), expected.ISO; !r.isoSpeed.Valid || got != want {
		t.Fatalf("iso_speed for %s: got %v, want %d", base, r.isoSpeed, want)
	}
	if !r.focalLength.Valid || r.focalLength.String != expected.FocalLength {
		t.Fatalf("focal_length for %s: got %v, want %q", base, r.focalLength, expected.FocalLength)
	}
	parsedLatitude := parseGPSCoordinateOrFail(t, expected.GPSLatitude)
	if !float64AlmostEqual(r.gpsLatitude.Float64, parsedLatitude, 1e-6) {
		t.Fatalf("gps_latitude for %s: got %f, want %f", base, r.gpsLatitude.Float64, parsedLatitude)
	}
	parsedLongitude := parseGPSCoordinateOrFail(t, expected.GPSLongitude)
	if !float64AlmostEqual(r.gpsLongitude.Float64, parsedLongitude, 1e-6) {
		t.Fatalf("gps_longitude for %s: got %f, want %f", base, r.gpsLongitude.Float64, parsedLongitude)
	}
	if r.status != "ok" {
		t.Fatalf("exif_status for %s: got %q, want %q", base, r.status, "ok")
	}
	if r.lastError.Valid {
		t.Fatalf("last_error for %s should be NULL, got %v", base, r.lastError)
	}

	assertExifJSONStringField(t, parsed, "Make", expected.Make, base)
	assertExifJSONStringField(t, parsed, "Model", expected.Model, base)
	assertExifJSONStringField(t, parsed, "DateTimeOriginal", expected.DateTimeOriginal, base)
	assertExifJSONStringField(t, parsed, "GPSLatitude", expected.GPSLatitude, base)
	assertExifJSONStringField(t, parsed, "GPSLongitude", expected.GPSLongitude, base)
	if got := int(asFloat64(t, getAnyJSONValue(t, parsed, "ImageWidth", "ExifImageWidth"))); got != expected.ImageWidth {
		t.Fatalf("exif_json ImageWidth for %s: got %d, want %d", base, got, expected.ImageWidth)
	}
	if got := int(asFloat64(t, getAnyJSONValue(t, parsed, "ImageHeight", "ExifImageHeight"))); got != expected.ImageHeight {
		t.Fatalf("exif_json ImageHeight for %s: got %d, want %d", base, got, expected.ImageHeight)
	}
	assertExifJSONStringField(t, parsed, "ExposureTime", expected.ExposureTime, base)
	if got := strconv.FormatFloat(asFloat64(t, parsed["FNumber"]), 'f', -1, 64); got != expected.FNumber {
		t.Fatalf("exif_json FNumber for %s: got %s, want %s", base, got, expected.FNumber)
	}
	if got := int(asFloat64(t, getAnyJSONValue(t, parsed, "ISO", "ISOSpeedRatings"))); got != expected.ISO {
		t.Fatalf("exif_json ISO for %s: got %d, want %d", base, got, expected.ISO)
	}
	assertExifJSONStringField(t, parsed, "FocalLength", expected.FocalLength, base)
}

type richExifExpected struct {
	Make             string
	Model            string
	DateTimeOriginal string
	GPSLatitude      string
	GPSLongitude     string
	ImageWidth       int
	ImageHeight      int
	ExposureTime     string
	FNumber          string
	ISO              int
	FocalLength      string
}

func createRichFixtureInputDir(t *testing.T, repoRoot string) string {
	t.Helper()

	inputDir := t.TempDir()
	src := filepath.Join(repoRoot, "testdata", richEXIFFixtureBaseName)
	dst := filepath.Join(inputDir, filepath.Base(src))

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read rich fixture: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write rich fixture copy: %v", err)
	}

	return inputDir
}

func richExifExpectedFromManifest(t *testing.T, repoRoot string) richExifExpected {
	t.Helper()

	manifest := filepath.Join(repoRoot, "testdata", "Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif-exiftool.txt")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read rich fixture manifest: %v", err)
	}

	fields := parseExifToolManifest(string(data))
	width, err := strconv.Atoi(fields["Image Width"])
	if err != nil {
		t.Fatalf("parse Image Width %q: %v", fields["Image Width"], err)
	}
	height, err := strconv.Atoi(fields["Image Height"])
	if err != nil {
		t.Fatalf("parse Image Height %q: %v", fields["Image Height"], err)
	}
	iso, err := strconv.Atoi(fields["ISO"])
	if err != nil {
		t.Fatalf("parse ISO %q: %v", fields["ISO"], err)
	}

	return richExifExpected{
		Make:             fields["Make"],
		Model:            fields["Camera Model Name"],
		DateTimeOriginal: fields["Date/Time Original"],
		GPSLatitude:      fields["GPS Latitude"],
		GPSLongitude:     fields["GPS Longitude"],
		ImageWidth:       width,
		ImageHeight:      height,
		ExposureTime:     fields["Exposure Time"],
		FNumber:          fields["F Number"],
		ISO:              iso,
		FocalLength:      normalizeManifestFocalLength(fields["Focal Length"]),
	}
}

func normalizeManifestFocalLength(value string) string {
	if idx := strings.Index(value, " ("); idx != -1 {
		return strings.TrimSpace(value[:idx])
	}
	return strings.TrimSpace(value)
}

func parseExifToolManifest(content string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" || val == "" {
			continue
		}
		fields[key] = val
	}
	return fields
}

func asFloat64(t *testing.T, v any) float64 {
	t.Helper()

	f, ok := v.(float64)
	if !ok {
		t.Fatalf("expected float64 value in exif_json, got %T (%v)", v, v)
	}
	return f
}

func getAnyJSONValue(t *testing.T, parsed map[string]any, keys ...string) any {
	t.Helper()

	for _, key := range keys {
		if v, ok := parsed[key]; ok {
			return v
		}
	}
	t.Fatalf("exif_json missing any of keys: %v", keys)
	return nil
}

func assertExifJSONStringField(t *testing.T, parsed map[string]any, key, want, base string) {
	t.Helper()

	if parsed[key] != want {
		t.Fatalf("exif_json %s for %s: got %v, want %q", key, base, parsed[key], want)
	}
}

var dmsCoordinatePattern = regexp.MustCompile(`(?i)^\s*([+-]?\d+(?:\.\d+)?)\s*deg\s*(\d+(?:\.\d+)?)'\s*(\d+(?:\.\d+)?)"\s*([NSEW])\s*$`)

func parseGPSCoordinateOrFail(t *testing.T, raw string) float64 {
	t.Helper()
	parsed, ok := parseGPSCoordinate(raw)
	if !ok {
		t.Fatalf("parse GPS coordinate %q", raw)
	}
	return parsed
}

func parseGPSCoordinate(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, false
	}
	if m := dmsCoordinatePattern.FindStringSubmatch(s); len(m) == 5 {
		degrees, err1 := strconv.ParseFloat(m[1], 64)
		minutes, err2 := strconv.ParseFloat(m[2], 64)
		seconds, err3 := strconv.ParseFloat(m[3], 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return 0, false
		}

		decimal := math.Abs(degrees) + (minutes / 60.0) + (seconds / 3600.0)
		hemisphere := strings.ToUpper(m[4])
		if hemisphere == "S" || hemisphere == "W" {
			decimal = -decimal
		}
		return decimal, true
	}
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

func float64AlmostEqual(a, b, eps float64) bool {
	if a == b {
		return true
	}
	delta := a - b
	if delta < 0 {
		delta = -delta
	}
	return delta <= eps
}

func assertNoExifRows(t *testing.T, dbPath string) {
	t.Helper()

	database := openSQLiteDB(t, dbPath)
	t.Cleanup(func() {
		_ = database.Close()
	})

	var (
		path      string
		makeVal   sql.NullString
		modelVal  sql.NullString
		dateVal   sql.NullString
		status    string
		lastError sql.NullString
		exifJSON  string
		err       error
	)

	err = database.QueryRow(`
SELECT file_path, camera_make, camera_model, date_time, exif_status, last_error, exif_json
FROM images
`).Scan(&path, &makeVal, &modelVal, &dateVal, &status, &lastError, &exifJSON)
	if err != nil {
		t.Fatalf("query row: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Fatalf("file_path is not absolute: %q", path)
	}
	if filepath.Base(path) != "blank.jpg" {
		t.Fatalf("unexpected file_path: %q", path)
	}
	if makeVal.Valid {
		t.Fatalf("camera_make should be NULL, got %v", makeVal)
	}
	if modelVal.Valid {
		t.Fatalf("camera_model should be NULL, got %v", modelVal)
	}
	if dateVal.Valid {
		t.Fatalf("date_time should be NULL, got %v", dateVal)
	}
	if status != "no_exif" {
		t.Fatalf("exif_status: got %q, want %q", status, "no_exif")
	}
	if lastError.Valid {
		t.Fatalf("last_error should be NULL, got %v", lastError)
	}
	if strings.TrimSpace(exifJSON) != "{}" {
		t.Fatalf("exif_json: got %q, want {}", exifJSON)
	}
}

func assertCorruptRows(t *testing.T, dbPath string) {
	t.Helper()

	database := openSQLiteDB(t, dbPath)
	t.Cleanup(func() {
		_ = database.Close()
	})

	var (
		path      string
		makeVal   sql.NullString
		modelVal  sql.NullString
		dateVal   sql.NullString
		status    string
		lastError sql.NullString
		exifJSON  string
		err       error
	)

	err = database.QueryRow(`
SELECT file_path, camera_make, camera_model, date_time, exif_status, last_error, exif_json
FROM images
`).Scan(&path, &makeVal, &modelVal, &dateVal, &status, &lastError, &exifJSON)
	if err != nil {
		t.Fatalf("query row: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Fatalf("file_path is not absolute: %q", path)
	}
	if filepath.Base(path) != "broken.jpg" {
		t.Fatalf("unexpected file_path: %q", path)
	}
	if makeVal.Valid {
		t.Fatalf("camera_make should be NULL, got %v", makeVal)
	}
	if modelVal.Valid {
		t.Fatalf("camera_model should be NULL, got %v", modelVal)
	}
	if dateVal.Valid {
		t.Fatalf("date_time should be NULL, got %v", dateVal)
	}
	if status != "parse_error" && status != "no_exif" {
		t.Fatalf("exif_status: got %q, want parse_error or no_exif", status)
	}
	if status == "parse_error" && (!lastError.Valid || strings.TrimSpace(lastError.String) == "") {
		t.Fatal("last_error should be non-empty when exif_status is parse_error")
	}
	if status == "no_exif" && lastError.Valid {
		t.Fatalf("last_error should be NULL when exif_status is no_exif, got %v", lastError)
	}
	if strings.TrimSpace(exifJSON) != "{}" {
		t.Fatalf("exif_json: got %q, want {}", exifJSON)
	}
}
