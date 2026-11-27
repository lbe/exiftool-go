//go:build integration

package pipeline

import (
	"database/sql"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

const richEXIFFixtureName = "Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif.jpg"

func fixtureDir(t *testing.T, subdir string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	return filepath.Join(root, "testdata", subdir)
}

func openDBForAssert(t *testing.T, dbPath string) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	return database
}

func fixturePath(t *testing.T, name string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	return filepath.Join(root, "testdata", name)
}

func assertAnyRawFieldPresent(t *testing.T, raw map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		v, ok := raw[key]
		if !ok {
			continue
		}
		if v == nil {
			t.Fatalf("raw exif key %q is nil", key)
		}

		if s, ok := v.(string); ok && s == "" {
			t.Fatalf("raw exif key %q is empty string", key)
		}
		return
	}
	t.Fatalf("raw exif missing all candidate keys: %v", keys)
}

func richFixtureInputDir(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	richFixture := fixturePath(t, richEXIFFixtureName)
	dst := filepath.Join(tmpDir, filepath.Base(richFixture))

	srcFile, err := os.Open(richFixture)
	if err != nil {
		t.Fatalf("open rich fixture: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create copied fixture: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		t.Fatalf("copy rich fixture: %v", err)
	}

	return tmpDir
}

func assertRawEXIFMatchesDBColumns(
	t *testing.T,
	parsed map[string]any,
	cameraMake sql.NullString,
	cameraModel sql.NullString,
	dateTime sql.NullString,
	gpsLatitude sql.NullFloat64,
	gpsLongitude sql.NullFloat64,
	imageWidth sql.NullInt64,
	imageHeight sql.NullInt64,
	exposureTime sql.NullString,
	fNumber sql.NullString,
	isoSpeed sql.NullInt64,
	focalLength sql.NullString,
) {
	t.Helper()

	if v, ok := parsed["Make"].(string); ok && cameraMake.Valid && cameraMake.String != v {
		t.Fatalf("camera_make mismatch: got %q, want %q", cameraMake.String, v)
	}

	if v, ok := parsed["Model"].(string); ok && cameraModel.Valid && cameraModel.String != v {
		t.Fatalf("camera_model mismatch from Model: got %q, want %q", cameraModel.String, v)
	}
	if v, ok := parsed["CameraModelName"].(string); ok && cameraModel.Valid && cameraModel.String != v {
		t.Fatalf("camera_model mismatch from CameraModelName: got %q, want %q", cameraModel.String, v)
	}

	if rawDate, ok := parsed["DateTimeOriginal"].(string); ok && dateTime.Valid {
		parsedDate, ok := parseDateTime(rawDate)
		if !ok {
			t.Fatalf("unable to parse DateTimeOriginal from exif_json: %q", rawDate)
		}
		storedDate, err := time.Parse(time.RFC3339, dateTime.String)
		if err != nil {
			t.Fatalf("unable to parse stored date_time %q: %v", dateTime.String, err)
		}
		if !storedDate.Equal(parsedDate.UTC()) {
			t.Fatalf("date_time mismatch: got %s, want %s", storedDate.Format(time.RFC3339), parsedDate.UTC().Format(time.RFC3339))
		}
	}

	if rawLat, ok := parsed["GPSLatitude"].(string); ok && gpsLatitude.Valid {
		expected, ok := parseGPSCoordinate(rawLat)
		if !ok {
			t.Fatalf("unable to parse GPSLatitude from exif_json: %q", rawLat)
		}
		if math.Abs(gpsLatitude.Float64-expected) > 1e-9 {
			t.Fatalf("gps_latitude mismatch: got %.12f, want %.12f", gpsLatitude.Float64, expected)
		}
	}

	if rawLon, ok := parsed["GPSLongitude"].(string); ok && gpsLongitude.Valid {
		expected, ok := parseGPSCoordinate(rawLon)
		if !ok {
			t.Fatalf("unable to parse GPSLongitude from exif_json: %q", rawLon)
		}
		if math.Abs(gpsLongitude.Float64-expected) > 1e-9 {
			t.Fatalf("gps_longitude mismatch: got %.12f, want %.12f", gpsLongitude.Float64, expected)
		}
	}

	if rawWidth, ok := parsed["ImageWidth"].(float64); ok && imageWidth.Valid && int64(math.Round(rawWidth)) != imageWidth.Int64 {
		t.Fatalf("image_width mismatch: got %d, want %d", imageWidth.Int64, int64(math.Round(rawWidth)))
	}
	if rawHeight, ok := parsed["ImageHeight"].(float64); ok && imageHeight.Valid && int64(math.Round(rawHeight)) != imageHeight.Int64 {
		t.Fatalf("image_height mismatch: got %d, want %d", imageHeight.Int64, int64(math.Round(rawHeight)))
	}

	if rawExposure, ok := parsed["ExposureTime"].(string); ok && exposureTime.Valid && exposureTime.String != rawExposure {
		t.Fatalf("exposure_time mismatch: got %q, want %q", exposureTime.String, rawExposure)
	}
	if rawFNumber, ok := parsed["FNumber"].(float64); ok && fNumber.Valid {
		expected := asStringForCompare(rawFNumber)
		if fNumber.String != expected {
			t.Fatalf("f_number mismatch: got %q, want %q", fNumber.String, expected)
		}
	}
	if rawISO, ok := parsed["ISO"].(float64); ok && isoSpeed.Valid && int64(math.Round(rawISO)) != isoSpeed.Int64 {
		t.Fatalf("iso_speed mismatch: got %d, want %d", isoSpeed.Int64, int64(math.Round(rawISO)))
	}
	if rawFocalLength, ok := parsed["FocalLength"].(string); ok && focalLength.Valid && focalLength.String != rawFocalLength {
		t.Fatalf("focal_length mismatch: got %q, want %q", focalLength.String, rawFocalLength)
	}
}

func asStringForCompare(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
