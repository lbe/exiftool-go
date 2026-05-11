package exifutil

import (
	"encoding/json"
	"testing"
)

// TestSerializeExifToJSON tests JSON serialization of EXIF data.
func TestSerializeExifToJSON(t *testing.T) {
	exifData := map[string]interface{}{
		"Make":             "Canon",
		"Model":            "Canon EOS 5D Mark IV",
		"DateTimeOriginal": "2020:06:15 14:30:00",
		"ISO":              100,
		"GPSLatitude":      "40.7128 N",
		"GPSLongitude":     "74.0060 W",
		"LensModel":        "EF24-70mm f/2.8L II USM",
		"ExposureTime":     "1/125",
		"FNumber":          "f/4.0",
		"FocalLength":      "50.0 mm",
		"WhiteBalance":     "Auto",
		"MeteringMode":     "Pattern",
		"Flash":            "Off, Did not fire",
		"ExifToolVersion":  "13.55",
	}

	jsonBytes, err := SerializeExif(exifData)
	if err != nil {
		t.Fatalf("SerializeExif failed: %v", err)
	}

	// Verify the output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Assert that key fields are present in the serialized output
	if len(parsed) == 0 {
		t.Errorf("SerializeExif returned empty JSON object")
	}

	// Check that we can retrieve some of the fields
	if v, ok := parsed["Make"]; !ok {
		t.Errorf("Make field not found in serialized JSON")
	} else if v != "Canon" {
		t.Errorf("Make field value mismatch: got %v, want Canon", v)
	}
}

// TestSerializeExifEmptyMap tests serialization of empty EXIF data.
func TestSerializeExifEmptyMap(t *testing.T) {
	exifData := map[string]interface{}{}

	jsonBytes, err := SerializeExif(exifData)
	if err != nil {
		t.Fatalf("SerializeExif failed on empty map: %v", err)
	}

	// Should produce valid JSON representing an empty object
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("Expected empty JSON object, got %d fields", len(parsed))
	}
}

// TestSerializeExifVariousTypes tests serialization of various data types in EXIF.
func TestSerializeExifVariousTypes(t *testing.T) {
	exifData := map[string]interface{}{
		"StringField": "value",
		"IntField":    42,
		"FloatField":  3.14,
		"BoolField":   true,
		"NilField":    nil,
	}

	jsonBytes, err := SerializeExif(exifData)
	if err != nil {
		t.Fatalf("SerializeExif failed: %v", err)
	}

	// Verify it's valid JSON and contains all fields
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 5 {
		t.Errorf("Expected 5 fields, got %d", len(parsed))
	}
}
