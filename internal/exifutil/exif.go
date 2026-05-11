// Package exifutil provides EXIF data extraction from image files using
// the go-exiftool-wasm library. It handles various image formats including
// JPEG, TIFF, WebP, HEIF, and PNG.
package exifutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	exiftool "github.com/lbe/go-exiftool-wasm"
)

// Extract extracts EXIF metadata from an image file.
//
// The imagePath can be either absolute or relative to the current working directory.
// It returns a map of EXIF field names to their values, or an error if:
//   - The file does not exist
//   - The file cannot be read
//   - ExifTool fails to process the file
//   - The JSON output cannot be parsed
//
// The returned map may be empty if the image contains no EXIF data, but the
// error will still be nil in that case.
func Extract(imagePath string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Resolve the absolute path and verify the file exists
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	if _, statErr := os.Stat(absPath); statErr != nil {
		return nil, fmt.Errorf("file not found: %w", statErr)
	}

	// Separate directory and filename for WASI sandbox mounting
	dir := filepath.Dir(absPath)
	filename := filepath.Base(absPath)

	// Mount the directory containing the image at /work inside the sandbox
	workDir := os.DirFS(dir)

	// Construct the guest path (exiftool mounts workDir at /work)
	guestPath := "/work/" + filename

	// Run exiftool with JSON output format
	out, err := exiftool.Run(ctx, workDir, "-json", guestPath)
	if err != nil {
		return nil, fmt.Errorf("exiftool extraction failed: %w", err)
	}

	// Parse the JSON output into a slice of maps
	var result []map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse exiftool JSON: %w", err)
	}

	// Return the first result (single image); empty map if no results
	if len(result) == 0 {
		return map[string]interface{}{}, nil
	}

	return result[0], nil
}
