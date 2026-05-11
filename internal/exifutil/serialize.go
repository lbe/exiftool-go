package exifutil

import (
	"encoding/json"
	"fmt"
)

// SerializeExif serializes EXIF metadata to compact JSON bytes.
//
// The function marshals the provided map of EXIF field names to values into
// JSON format using the standard encoding/json package. This is suitable for
// storage in the database as raw EXIF data.
//
// The returned JSON is compact (no indentation). For display purposes,
// consider unmarshaling and re-marshaling with an encoder that adds indentation.
//
// Parameters:
//   - exifData: A map of EXIF field names to their values (typically extracted via Extract)
//
// Returns:
//   - []byte: Compact JSON representation of the EXIF data
//   - error: If the data cannot be marshaled (rare for maps of interface{})
//
// Example:
//
//	exifData := map[string]interface{}{"Make": "Canon", "ISO": 100}
//	jsonBytes, err := SerializeExif(exifData)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// jsonBytes now contains: {"ISO":100,"Make":"Canon"}
func SerializeExif(exifData map[string]interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(exifData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal EXIF data to JSON: %w", err)
	}
	return jsonBytes, nil
}
