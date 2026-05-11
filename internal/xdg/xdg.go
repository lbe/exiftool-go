package xdg

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const appDirName = "exiftool-go"

// DataDir resolves and creates the application's XDG data directory.
func DataDir() (string, error) {
	slog.Debug("DataDir enter")

	baseDir, err := resolveBaseDir()
	if err != nil {
		return "", err
	}

	dataDir := filepath.Join(baseDir, appDirName)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("create data directory %q: %w", dataDir, err)
	}

	slog.Debug("DataDir exit", "path", dataDir)
	return dataDir, nil
}

func resolveBaseDir() (string, error) {
	slog.Debug("resolveBaseDir enter")

	baseDir := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if baseDir != "" {
		slog.Debug("resolveBaseDir exit", "base_dir", baseDir, "source", "xdg_data_home")
		return baseDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	baseDir = filepath.Join(homeDir, ".local", "share")
	slog.Debug("resolveBaseDir exit", "base_dir", baseDir, "source", "home_fallback")
	return baseDir, nil
}
