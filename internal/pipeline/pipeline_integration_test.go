//go:build integration

package pipeline

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/lbe/exiftool-go/internal/exifutil"
	"github.com/lbe/exiftool-go/internal/meta"
	"github.com/lbe/exiftool-go/internal/testsupport"
)

func TestC101RichFixtureContractFullStructuredCoverage(t *testing.T) {
	t.Helper()

	fixture := fixturePath(t, richEXIFFixtureName)
	rawEXIF, err := exifutil.Extract(fixture)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	assertAnyRawFieldPresent(t, rawEXIF, "Make")
	assertAnyRawFieldPresent(t, rawEXIF, "Model", "CameraModelName")
	assertAnyRawFieldPresent(t, rawEXIF, "DateTimeOriginal")
	assertAnyRawFieldPresent(t, rawEXIF, "GPSLatitude")
	assertAnyRawFieldPresent(t, rawEXIF, "GPSLongitude")
	assertAnyRawFieldPresent(t, rawEXIF, "ImageWidth", "ExifImageWidth")
	assertAnyRawFieldPresent(t, rawEXIF, "ImageHeight", "ExifImageHeight")
	assertAnyRawFieldPresent(t, rawEXIF, "ExposureTime")
	assertAnyRawFieldPresent(t, rawEXIF, "FNumber")
	assertAnyRawFieldPresent(t, rawEXIF, "ISO", "ISOSpeedRatings")
	assertAnyRawFieldPresent(t, rawEXIF, "FocalLength")

	m := metadataFromExtract(fixture, "hash-c101", rawEXIF)

	if m.EXIFStatus != meta.EXIFStatusOK {
		t.Fatalf("exif_status: got %q, want %q", m.EXIFStatus, meta.EXIFStatusOK)
	}
	if m.LastError != nil {
		t.Fatalf("last_error: got %v, want nil", m.LastError)
	}

	assertNotNilField(t, "camera_make", m.CameraMake)
	assertNotNilField(t, "camera_model", m.CameraModel)
	assertNotNilField(t, "date_time", m.DateTime)
	assertNotNilField(t, "gps_latitude", m.GPSLatitude)
	assertNotNilField(t, "gps_longitude", m.GPSLongitude)
	assertNotNilField(t, "image_width", m.ImageWidth)
	assertNotNilField(t, "image_height", m.ImageHeight)
	assertNotNilField(t, "exposure_time", m.ExposureTime)
	assertNotNilField(t, "f_number", m.FNumber)
	assertNotNilField(t, "iso_speed", m.ISOSpeed)
	assertNotNilField(t, "focal_length", m.FocalLength)
}

func TestRunProcessesFixtureDirsIntoDB(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPath := filepath.Join(t.TempDir(), "exif.db")
	withExifDir := richFixtureInputDir(t)
	noExifDir := fixtureDir(t, "no_exif")

	err := Run(ctx, []string{withExifDir, noExifDir}, dbPath, 2)
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	database := openDBForAssert(t, dbPath)
	defer database.Close()

	var totalRows int
	if err := database.QueryRow(`SELECT COUNT(*) FROM images`).Scan(&totalRows); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if totalRows < 2 {
		t.Fatalf("rows: got %d, want at least 2", totalRows)
	}

	var okRows int
	if err := database.QueryRow(`SELECT COUNT(*) FROM images WHERE exif_status = 'ok' AND camera_make IS NOT NULL AND camera_model IS NOT NULL`).Scan(&okRows); err != nil {
		t.Fatalf("count ok rows: %v", err)
	}
	if okRows == 0 {
		t.Fatalf("expected at least one row with non-null EXIF fields and exif_status='ok'")
	}

	var noExifRows int
	if err := database.QueryRow(`SELECT COUNT(*) FROM images WHERE exif_status = 'no_exif'`).Scan(&noExifRows); err != nil {
		t.Fatalf("count no_exif rows: %v", err)
	}
	if noExifRows == 0 {
		t.Fatalf("expected at least one row with exif_status='no_exif'")
	}
}

func TestRunHonorsContextCancellationWithoutDeadlock(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dbPath := filepath.Join(t.TempDir(), "cancelled.db")
	dirs := []string{fixtureDir(t, "with_exif")}

	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, dirs, dbPath, 2)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Run returned nil error with already-cancelled context; want cancellation error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run cancellation error: got %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run appears deadlocked under cancelled context")
	}
}

func TestC107PipelineRunPersistsRichFixtureWithJSONStructuredConsistency(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPath := filepath.Join(t.TempDir(), "c107.db")
	withExifDir := richFixtureInputDir(t)

	err := Run(ctx, []string{withExifDir}, dbPath, 1)
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	database := openDBForAssert(t, dbPath)
	defer database.Close()

	row := database.QueryRow(`
SELECT
	camera_make,
	camera_model,
	date_time,
	gps_latitude,
	gps_longitude,
	image_width,
	image_height,
	orientation,
	exposure_time,
	f_number,
	iso_speed,
	focal_length,
	exif_json,
	exif_status,
	last_error
FROM images
WHERE file_path LIKE '%' || ?
`, richEXIFFixtureName)

	var (
		cameraMake   sql.NullString
		cameraModel  sql.NullString
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
		exifJSON     string
		exifStatus   string
		lastError    sql.NullString
	)

	if err := row.Scan(
		&cameraMake,
		&cameraModel,
		&dateTime,
		&gpsLatitude,
		&gpsLongitude,
		&imageWidth,
		&imageHeight,
		&orientation,
		&exposureTime,
		&fNumber,
		&isoSpeed,
		&focalLength,
		&exifJSON,
		&exifStatus,
		&lastError,
	); err != nil {
		t.Fatalf("scan rich fixture row: %v", err)
	}

	if exifStatus != meta.EXIFStatusOK {
		t.Fatalf("exif_status: got %q, want %q", exifStatus, meta.EXIFStatusOK)
	}
	if lastError.Valid {
		t.Fatalf("last_error: got %v, want NULL", lastError)
	}

	parsed := testsupport.ParseEXIFJSON(t, exifJSON)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"Make"}, "camera_make", cameraMake.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"Model", "CameraModelName"}, "camera_model", cameraModel.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"DateTimeOriginal"}, "date_time", dateTime.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"GPSLatitude"}, "gps_latitude", gpsLatitude.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"GPSLongitude"}, "gps_longitude", gpsLongitude.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ImageWidth", "ExifImageWidth"}, "image_width", imageWidth.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ImageHeight", "ExifImageHeight"}, "image_height", imageHeight.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ExposureTime"}, "exposure_time", exposureTime.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"FNumber"}, "f_number", fNumber.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"ISO", "ISOSpeedRatings"}, "iso_speed", isoSpeed.Valid)
	testsupport.AssertRawPresenceImpliesDBNotNull(t, parsed, []string{"FocalLength"}, "focal_length", focalLength.Valid)
	if _, hasOrientation := parsed["Orientation"]; hasOrientation {
		if !orientation.Valid {
			t.Fatalf("orientation is NULL while exif_json contains Orientation")
		}
		expectedOrientation, ok := exifOrientation(parsed, "Orientation")
		if !ok {
			t.Fatalf("unable to parse orientation from exif_json")
		}
		if int64(expectedOrientation) != orientation.Int64 {
			t.Fatalf("orientation mismatch: got %d, want %d", orientation.Int64, expectedOrientation)
		}
	}

	assertRawEXIFMatchesDBColumns(t, parsed, cameraMake, cameraModel, dateTime, gpsLatitude, gpsLongitude, imageWidth, imageHeight, exposureTime, fNumber, isoSpeed, focalLength)
}
