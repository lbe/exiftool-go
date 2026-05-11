package db

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/lbe/exiftool-go/internal/meta"
)

func TestInsertImageRoundTrip(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "exif.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := database.Close(); closeErr != nil {
			t.Fatalf("Close returned error: %v", closeErr)
		}
	})

	makeValue := "Canon"
	modelValue := "EOS 80D"
	latValue := 40.7128
	lonValue := -74.0060
	widthValue := 6000
	heightValue := 4000
	orientationValue := 1
	exposureValue := "1/125"
	fNumberValue := "5.6"
	isoValue := 100
	focalLengthValue := "35.0 mm"
	statusValue := meta.EXIFStatusOK
	dateTimeValue := time.Date(2024, time.January, 1, 12, 30, 45, 0, time.UTC)

	input := meta.ImageMeta{
		FilePath:     "/tmp/photo.jpg",
		FileHash:     "abc123",
		CameraMake:   &makeValue,
		CameraModel:  &modelValue,
		DateTime:     &dateTimeValue,
		GPSLatitude:  &latValue,
		GPSLongitude: &lonValue,
		ImageWidth:   &widthValue,
		ImageHeight:  &heightValue,
		Orientation:  &orientationValue,
		ExposureTime: &exposureValue,
		FNumber:      &fNumberValue,
		ISOSpeed:     &isoValue,
		FocalLength:  &focalLengthValue,
		EXIFStatus:   statusValue,
		RawEXIF: map[string]any{
			"Make":             makeValue,
			"Model":            modelValue,
			"DateTimeOriginal": "2024:01:01 12:30:45",
		},
	}

	if err := InsertImage(database, input); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}

	stored := readStoredImage(t, database, input.FilePath)
	assertStoredImage(t, stored, input)
}

func TestInsertImageDeduplicatesByPath(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "exif.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := database.Close(); closeErr != nil {
			t.Fatalf("Close returned error: %v", closeErr)
		}
	})

	makeValue := "Canon"
	modelValue := "EOS 80D"
	statusValue := meta.EXIFStatusOK
	dateTimeValue := time.Date(2024, time.January, 1, 12, 30, 45, 0, time.UTC)

	first := meta.ImageMeta{
		FilePath:    "/tmp/photo.jpg",
		FileHash:    "hash-one",
		CameraMake:  &makeValue,
		CameraModel: &modelValue,
		DateTime:    &dateTimeValue,
		EXIFStatus:  statusValue,
		RawEXIF: map[string]any{
			"Make":  makeValue,
			"Model": modelValue,
		},
	}

	second := first
	second.FileHash = "hash-two"

	if err := InsertImage(database, first); err != nil {
		t.Fatalf("InsertImage first returned error: %v", err)
	}
	if err := InsertImage(database, second); err != nil {
		t.Fatalf("InsertImage second returned error: %v", err)
	}

	assertImageCount(t, database, 1)
	stored := readStoredImage(t, database, first.FilePath)
	if stored.fileHash != second.FileHash {
		t.Fatalf("file_hash after second insert: got %q, want %q", stored.fileHash, second.FileHash)
	}
}

type storedImage struct {
	fileHash     string
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
	createdAt    string
	updatedAt    string
}

func readStoredImage(t *testing.T, database *sql.DB, filePath string) storedImage {
	t.Helper()

	var got storedImage
	err := database.QueryRow(`
SELECT
	file_hash,
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
	last_error,
	created_at,
	updated_at
FROM images
WHERE file_path = ?
`, filePath).Scan(
		&got.fileHash,
		&got.cameraMake,
		&got.cameraModel,
		&got.dateTime,
		&got.gpsLatitude,
		&got.gpsLongitude,
		&got.imageWidth,
		&got.imageHeight,
		&got.orientation,
		&got.exposureTime,
		&got.fNumber,
		&got.isoSpeed,
		&got.focalLength,
		&got.exifJSON,
		&got.exifStatus,
		&got.lastError,
		&got.createdAt,
		&got.updatedAt,
	)
	if err != nil {
		t.Fatalf("QueryRow stored image: %v", err)
	}

	return got
}

func assertStoredImage(t *testing.T, got storedImage, want meta.ImageMeta) {
	t.Helper()

	if got.fileHash != want.FileHash {
		t.Fatalf("file_hash: got %q, want %q", got.fileHash, want.FileHash)
	}
	if !got.cameraMake.Valid || got.cameraMake.String != *want.CameraMake {
		t.Fatalf("camera_make: got %v, want %q", got.cameraMake, *want.CameraMake)
	}
	if !got.cameraModel.Valid || got.cameraModel.String != *want.CameraModel {
		t.Fatalf("camera_model: got %v, want %q", got.cameraModel, *want.CameraModel)
	}
	if !got.dateTime.Valid {
		t.Fatal("date_time should be valid")
	}
	if !got.gpsLatitude.Valid || got.gpsLatitude.Float64 != *want.GPSLatitude {
		t.Fatalf("gps_latitude: got %v, want %v", got.gpsLatitude, *want.GPSLatitude)
	}
	if !got.gpsLongitude.Valid || got.gpsLongitude.Float64 != *want.GPSLongitude {
		t.Fatalf("gps_longitude: got %v, want %v", got.gpsLongitude, *want.GPSLongitude)
	}
	if !got.imageWidth.Valid || got.imageWidth.Int64 != int64(*want.ImageWidth) {
		t.Fatalf("image_width: got %v, want %v", got.imageWidth, *want.ImageWidth)
	}
	if !got.imageHeight.Valid || got.imageHeight.Int64 != int64(*want.ImageHeight) {
		t.Fatalf("image_height: got %v, want %v", got.imageHeight, *want.ImageHeight)
	}
	if !got.orientation.Valid || got.orientation.Int64 != int64(*want.Orientation) {
		t.Fatalf("orientation: got %v, want %v", got.orientation, *want.Orientation)
	}
	if !got.exposureTime.Valid || got.exposureTime.String != *want.ExposureTime {
		t.Fatalf("exposure_time: got %v, want %v", got.exposureTime, *want.ExposureTime)
	}
	if !got.fNumber.Valid || got.fNumber.String != *want.FNumber {
		t.Fatalf("f_number: got %v, want %v", got.fNumber, *want.FNumber)
	}
	if !got.isoSpeed.Valid || got.isoSpeed.Int64 != int64(*want.ISOSpeed) {
		t.Fatalf("iso_speed: got %v, want %v", got.isoSpeed, *want.ISOSpeed)
	}
	if !got.focalLength.Valid || got.focalLength.String != *want.FocalLength {
		t.Fatalf("focal_length: got %v, want %v", got.focalLength, *want.FocalLength)
	}
	if got.exifStatus != want.EXIFStatus {
		t.Fatalf("exif_status: got %q, want %q", got.exifStatus, want.EXIFStatus)
	}
	if got.lastError.Valid {
		t.Fatalf("last_error: got %v, want NULL", got.lastError)
	}
	if got.createdAt == "" {
		t.Fatal("created_at should not be empty")
	}
	if got.updatedAt == "" {
		t.Fatal("updated_at should not be empty")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(got.exifJSON), &parsed); err != nil {
		t.Fatalf("exif_json unmarshal: %v", err)
	}
	if parsed["Make"] != (*want.CameraMake) {
		t.Fatalf("exif_json Make: got %v, want %v", parsed["Make"], *want.CameraMake)
	}
}

func assertImageCount(t *testing.T, database *sql.DB, want int) {
	t.Helper()

	var got int
	err := database.QueryRow(`SELECT COUNT(*) FROM images`).Scan(&got)
	if err != nil {
		t.Fatalf("QueryRow image count: %v", err)
	}
	if got != want {
		t.Fatalf("image count: got %d, want %d", got, want)
	}
}
