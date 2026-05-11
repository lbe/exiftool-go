package meta

import (
	"encoding/json"
	"testing"
	"time"
)

func TestImageMetaJSONRoundTrip(t *testing.T) {
	t.Helper()

	makeValue := "Canon"
	modelValue := "EOS 80D"
	errorValue := "parse failed"
	latitudeValue := 40.7128
	longitudeValue := -74.0060
	widthValue := 6000
	heightValue := 4000
	orientationValue := 1
	exposureValue := "1/125"
	fNumberValue := "5.6"
	isoValue := 100
	focalLengthValue := "35.0 mm"
	dateTimeValue := time.Date(2022, time.January, 1, 12, 34, 56, 0, time.UTC)

	original := ImageMeta{
		FilePath:     "/photos/sample.jpg",
		FileHash:     "abc123",
		CameraMake:   &makeValue,
		CameraModel:  &modelValue,
		DateTime:     &dateTimeValue,
		GPSLatitude:  &latitudeValue,
		GPSLongitude: &longitudeValue,
		ImageWidth:   &widthValue,
		ImageHeight:  &heightValue,
		Orientation:  &orientationValue,
		ExposureTime: &exposureValue,
		FNumber:      &fNumberValue,
		ISOSpeed:     &isoValue,
		FocalLength:  &focalLengthValue,
		EXIFStatus:   EXIFStatusParseError,
		LastError:    &errorValue,
		RawEXIF: map[string]any{
			"Make":             makeValue,
			"CameraModelName":  modelValue,
			"DateTimeOriginal": "2022:01:01 12:34:56",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ImageMeta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.FilePath != original.FilePath {
		t.Fatalf("FilePath: got %q, want %q", decoded.FilePath, original.FilePath)
	}
	if decoded.FileHash != original.FileHash {
		t.Fatalf("FileHash: got %q, want %q", decoded.FileHash, original.FileHash)
	}
	if decoded.CameraMake == nil || *decoded.CameraMake != *original.CameraMake {
		t.Fatalf("CameraMake: got %v, want %q", decoded.CameraMake, *original.CameraMake)
	}
	if decoded.CameraModel == nil || *decoded.CameraModel != *original.CameraModel {
		t.Fatalf("CameraModel: got %v, want %q", decoded.CameraModel, *original.CameraModel)
	}
	if decoded.DateTime == nil || !decoded.DateTime.Equal(*original.DateTime) {
		t.Fatalf("DateTime: got %v, want %v", decoded.DateTime, original.DateTime)
	}
	if decoded.GPSLatitude == nil || *decoded.GPSLatitude != *original.GPSLatitude {
		t.Fatalf("GPSLatitude: got %v, want %v", decoded.GPSLatitude, *original.GPSLatitude)
	}
	if decoded.GPSLongitude == nil || *decoded.GPSLongitude != *original.GPSLongitude {
		t.Fatalf("GPSLongitude: got %v, want %v", decoded.GPSLongitude, *original.GPSLongitude)
	}
	if decoded.ImageWidth == nil || *decoded.ImageWidth != *original.ImageWidth {
		t.Fatalf("ImageWidth: got %v, want %v", decoded.ImageWidth, *original.ImageWidth)
	}
	if decoded.ImageHeight == nil || *decoded.ImageHeight != *original.ImageHeight {
		t.Fatalf("ImageHeight: got %v, want %v", decoded.ImageHeight, *original.ImageHeight)
	}
	if decoded.Orientation == nil || *decoded.Orientation != *original.Orientation {
		t.Fatalf("Orientation: got %v, want %v", decoded.Orientation, *original.Orientation)
	}
	if decoded.ExposureTime == nil || *decoded.ExposureTime != *original.ExposureTime {
		t.Fatalf("ExposureTime: got %v, want %v", decoded.ExposureTime, *original.ExposureTime)
	}
	if decoded.FNumber == nil || *decoded.FNumber != *original.FNumber {
		t.Fatalf("FNumber: got %v, want %v", decoded.FNumber, *original.FNumber)
	}
	if decoded.ISOSpeed == nil || *decoded.ISOSpeed != *original.ISOSpeed {
		t.Fatalf("ISOSpeed: got %v, want %v", decoded.ISOSpeed, *original.ISOSpeed)
	}
	if decoded.FocalLength == nil || *decoded.FocalLength != *original.FocalLength {
		t.Fatalf("FocalLength: got %v, want %v", decoded.FocalLength, *original.FocalLength)
	}
	if decoded.EXIFStatus != original.EXIFStatus {
		t.Fatalf("EXIFStatus: got %q, want %q", decoded.EXIFStatus, original.EXIFStatus)
	}
	if decoded.LastError == nil || *decoded.LastError != *original.LastError {
		t.Fatalf("LastError: got %v, want %v", decoded.LastError, *original.LastError)
	}
	if got, want := decoded.RawEXIF["Make"], original.RawEXIF["Make"]; got != want {
		t.Fatalf("RawEXIF Make: got %v, want %v", got, want)
	}
	if got, want := decoded.RawEXIF["DateTimeOriginal"], original.RawEXIF["DateTimeOriginal"]; got != want {
		t.Fatalf("RawEXIF DateTimeOriginal: got %v, want %v", got, want)
	}
}

func TestImageMetaZeroValueJSON(t *testing.T) {
	t.Helper()

	data, err := json.Marshal(ImageMeta{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if got, want := string(data), "{}"; got != want {
		t.Fatalf("Marshal zero value: got %q, want %q", got, want)
	}

	var decoded ImageMeta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.FilePath != "" {
		t.Fatalf("FilePath: got %q, want empty", decoded.FilePath)
	}
	if decoded.RawEXIF != nil {
		t.Fatalf("RawEXIF: got %v, want nil", decoded.RawEXIF)
	}
	if decoded.DateTime != nil {
		t.Fatalf("DateTime: got %v, want nil", decoded.DateTime)
	}
	if decoded.LastError != nil {
		t.Fatalf("LastError: got %v, want nil", decoded.LastError)
	}
}
