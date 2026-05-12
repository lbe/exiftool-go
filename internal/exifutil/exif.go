// Package exifutil provides EXIF data extraction using the unified
// internal/backend driver (default wasm).
package exifutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lbe/exiftool-go/internal/backend"
)

// Extract extracts EXIF metadata from an image file using the wasm backend by
// default (same default as omitting --backend on the CLI).
func Extract(imagePath string) (map[string]interface{}, error) {
	ctx := context.Background()

	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	if _, statErr := os.Stat(absPath); statErr != nil {
		return nil, fmt.Errorf("file not found: %w", statErr)
	}

	d := backend.NewDispatchDriver(backend.BackendWasm)
	out, err := d.CommandContext(ctx, nil, "-json", absPath)
	if err != nil {
		return nil, fmt.Errorf("exiftool extraction failed: %w", err)
	}

	return parseExiftoolJSONArray(out)
}

// ExtractServer runs JSON extraction using a long-lived stay_open session.
func ExtractServer(srv backend.Server, imagePath string) (map[string]interface{}, error) {
	if srv == nil {
		return nil, fmt.Errorf("exiftool server is nil")
	}

	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	if _, statErr := os.Stat(absPath); statErr != nil {
		return nil, fmt.Errorf("file not found: %w", statErr)
	}

	out, err := srv.Command("-json", absPath)
	if err != nil {
		return nil, fmt.Errorf("exiftool extraction failed: %w", err)
	}

	return parseExiftoolJSONArray(out)
}

func parseExiftoolJSONArray(out []byte) (map[string]interface{}, error) {
	var result []map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse exiftool JSON: %w", err)
	}

	if len(result) == 0 {
		return map[string]interface{}{}, nil
	}

	return result[0], nil
}
