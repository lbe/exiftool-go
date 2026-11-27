package pipeline

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/lbe/exiftool-go/internal/db"
	"github.com/lbe/exiftool-go/internal/meta"
	"github.com/lbe/exiftool-go/internal/testsupport"
)

func TestC104MetadataFromExtractMapsAllStructuredFields(t *testing.T) {
	t.Helper()

	rawEXIF := c104FullRawEXIFSample()
	m := metadataFromExtract("/tmp/c104-rich.jpg", "hash-c104", rawEXIF)

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
	assertNotNilField(t, "orientation", m.Orientation)
	assertNotNilField(t, "exposure_time", m.ExposureTime)
	assertNotNilField(t, "f_number", m.FNumber)
	assertNotNilField(t, "iso_speed", m.ISOSpeed)
	assertNotNilField(t, "focal_length", m.FocalLength)
}

func TestC104MetadataFromExtractRawEXIFJSONAndStructuredFieldsAreConsistent(t *testing.T) {
	t.Helper()

	rawEXIF := c104FullRawEXIFSample()
	m := metadataFromExtract("/tmp/c104-rich.jpg", "hash-c104", rawEXIF)
	parsed := testsupport.AssertRawEXIFJSONParsable(t, m.RawEXIF)

	testsupport.AssertRawJSONHasAnyKey(t, parsed, "Make")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "Model", "CameraModelName")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "DateTimeOriginal")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "GPSLatitude")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "GPSLongitude")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "ImageWidth", "ExifImageWidth")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "ImageHeight", "ExifImageHeight")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "Orientation")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "ExposureTime")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "FNumber")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "ISO", "ISOSpeedRatings")
	testsupport.AssertRawJSONHasAnyKey(t, parsed, "FocalLength")

	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"Make"}, "camera_make", m.CameraMake)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"Model", "CameraModelName"}, "camera_model", m.CameraModel)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"DateTimeOriginal"}, "date_time", m.DateTime)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"GPSLatitude"}, "gps_latitude", m.GPSLatitude)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"GPSLongitude"}, "gps_longitude", m.GPSLongitude)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"ImageWidth", "ExifImageWidth"}, "image_width", m.ImageWidth)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"ImageHeight", "ExifImageHeight"}, "image_height", m.ImageHeight)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"Orientation"}, "orientation", m.Orientation)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"ExposureTime"}, "exposure_time", m.ExposureTime)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"FNumber"}, "f_number", m.FNumber)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"ISO", "ISOSpeedRatings"}, "iso_speed", m.ISOSpeed)
	testsupport.AssertRawPresenceImpliesStructuredNonNil(t, parsed, []string{"FocalLength"}, "focal_length", m.FocalLength)
}

func TestC107InsertImageRejectsJSONStructuredInconsistency(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "c107-guard.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	defer database.Close()

	now := time.Now().UTC()
	inconsistent := meta.ImageMeta{
		FilePath:    "/tmp/inconsistent-rich.jpg",
		FileHash:    "hash-c107-guard",
		DateTime:    &now,
		EXIFStatus:  meta.EXIFStatusOK,
		RawEXIF:     c104FullRawEXIFSample(),
		CameraMake:  nil,
		CameraModel: nil,
	}

	err = db.InsertImage(database, inconsistent)
	if err == nil {
		t.Fatal("expected insert to fail when mapped EXIF keys exist in raw JSON but structured columns are nil")
	}
}

func assertNotNilField(t *testing.T, name string, value any) { //nolint:govet
	t.Helper()

	if value == nil {
		t.Fatalf("%s: got nil, want non-nil", name)
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.IsNil() { //nolint:govet
		t.Fatalf("%s: got nil pointer, want non-nil", name)
	}
}

func c104FullRawEXIFSample() map[string]any {
	return map[string]any{
		"Make":             "samsung",
		"CameraModelName":  "SM-G930P",
		"DateTimeOriginal": "2017:05:29 11:11:16",
		"GPSLatitude":      "26 deg 34' 57.06\" N",
		"GPSLongitude":     "80 deg 12' 0.84\" W",
		"ImageWidth":       float64(650),
		"ImageHeight":      float64(488),
		"Orientation":      "Rotate 90 CW",
		"ExposureTime":     "1/1600",
		"FNumber":          float64(1.7),
		"ISO":              float64(40),
		"FocalLength":      "4.2 mm",
	}
}
