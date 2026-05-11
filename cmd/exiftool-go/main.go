package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lbe/exiftool-go/internal/input"
	"github.com/lbe/exiftool-go/internal/logging"
	"github.com/lbe/exiftool-go/internal/pipeline"
	"github.com/lbe/exiftool-go/internal/xdg"
)

const defaultLogLevel = "INFO"

// cliConfig stores command-line options for the executable.
type cliConfig struct {
	logLevel string
	workers  int
	dirs     []string
}

func main() {
	if err := run(os.Args[1:], os.Stdin); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin *os.File) error {
	slog.Debug("run enter", "args_count", len(args))

	cfg, err := parseFlags(args)
	if err != nil {
		return err
	}
	if setupErr := logging.Setup(cfg.logLevel); setupErr != nil {
		return setupErr
	}

	dirs, resolveErr := input.ResolveDirs(cfg.dirs, stdin)
	if resolveErr != nil {
		return resolveErr
	}

	dataDir, dataDirErr := xdg.DataDir()
	if dataDirErr != nil {
		return dataDirErr
	}
	dbPath := filepath.Join(dataDir, "exif.db")

	slog.Info("application started", "dirs", dirs, "log_level", cfg.logLevel, "workers", cfg.workers)

	err = pipeline.Run(context.Background(), dirs, dbPath, cfg.workers)
	if err != nil {
		return err
	}

	slog.Debug("run exit", "error", nil)
	return nil
}

func parseFlags(args []string) (cliConfig, error) {
	slog.Debug("parseFlags enter", "args_count", len(args))

	workersDefault := runtime.NumCPU()
	if workersDefault < 1 {
		workersDefault = 1
	}

	flagSet := flag.NewFlagSet("exiftool-go", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	shortLevel := flagSet.String("l", defaultLogLevel, "log level")
	longLevel := flagSet.String("log-level", "", "log level")
	shortWorkers := flagSet.Int("w", workersDefault, "worker count")
	longWorkers := flagSet.Int("workers", 0, "worker count")
	flagSet.Usage = func() {
		_, _ = fmt.Fprintln(flagSet.Output(), usageMessage(workersDefault))
	}

	if err := flagSet.Parse(args); err != nil {
		return cliConfig{}, err
	}

	level := normalizeLevel(*shortLevel)
	if long := normalizeLevel(*longLevel); long != "" {
		level = long
	}
	if _, err := logging.ParseLevel(level); err != nil {
		return cliConfig{}, fmt.Errorf("invalid log level: %w", err)
	}

	workers := *shortWorkers
	if *longWorkers != 0 {
		workers = *longWorkers
	}
	if workers < 1 {
		return cliConfig{}, fmt.Errorf("invalid worker count %d", workers)
	}

	cfg := cliConfig{logLevel: level, workers: workers, dirs: flagSet.Args()}
	slog.Debug("parseFlags exit", "workers", cfg.workers, "dirs_count", len(cfg.dirs), "log_level", cfg.logLevel)
	return cfg, nil
}

func normalizeLevel(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func usageMessage(defaultWorkers int) string {
	return fmt.Sprintf("usage: exiftool-go [-l LEVEL|--log-level LEVEL] [-w N|--workers N] [DIR ...]\n  LEVEL: DEBUG, INFO, WARN, ERROR (default: %s)\n  N: worker count >= 1 (default: %d)", defaultLogLevel, defaultWorkers)
}
