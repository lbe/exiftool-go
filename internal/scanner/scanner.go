package scanner

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
)

var imageExtensions = []string{".jpg", ".jpeg", ".tif", ".tiff", ".heif", ".heic", ".webp", ".png"}

// ScanDir walks a directory tree and returns all supported image file paths in lexical order.
func ScanDir(root string) ([]string, error) {
	slog.Info("ScanDir start", "root", root)
	slog.Debug("ScanDir enter", "root", root)

	paths, err := collectImagePaths(root)
	if err != nil {
		return nil, err
	}

	slices.Sort(paths)
	slog.Debug("ScanDir exit", "root", root, "count", len(paths))
	return paths, nil
}

func collectImagePaths(root string) ([]string, error) {
	slog.Debug("collectImagePaths enter", "root", root)

	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !isImageFile(path) {
			return nil
		}

		slog.Debug("ScanDir found image", "path", path)
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan directory %q: %w", root, err)
	}

	slog.Debug("collectImagePaths exit", "root", root, "count", len(paths))
	return paths, nil
}

func isImageFile(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return slices.Contains(imageExtensions, extension)
}
