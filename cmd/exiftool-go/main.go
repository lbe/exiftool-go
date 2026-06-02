package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lbe/exiftool-go/internal/backend"
	"github.com/lbe/exiftool-go/internal/input"
	"github.com/lbe/exiftool-go/internal/logging"
	"github.com/lbe/exiftool-go/internal/pipeline"
	"github.com/lbe/exiftool-go/internal/xdg"
	wasmexif "github.com/lbe/go-exiftool-wasm"
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
		var exitErr exitError
		if errors.As(err, &exitErr) {
			if len(exitErr.stderr) > 0 {
				_, _ = os.Stderr.Write(exitErr.stderr)
			}
			os.Exit(exitErr.code)
		}
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// exitError carries a subprocess exit code and optional stderr from passthrough mode.
type exitError struct {
	code   int
	stderr []byte
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func run(args []string, stdin *os.File) error {
	slog.Debug("run enter", "args_count", len(args))

	if len(args) == 1 && (args[0] == "--help" || args[0] == "-help") {
		_, _ = fmt.Fprint(os.Stdout, globalUsage(workersDefault()))
		return nil
	}
	if len(args) == 1 && args[0] == "--version" {
		_, _ = fmt.Fprintln(os.Stdout, "exiftool-go (development build)")
		return nil
	}

	rest, backendKind, err := backend.StripLeadingBackendFlags(args)
	if err != nil {
		return err
	}

	if len(rest) >= 1 && rest[0] == "pipeline" {
		cfg, err := parseFlags(rest[1:])
		if err != nil {
			return err
		}
		return runIngest(cfg, stdin)
	}

	return runForward(context.Background(), stdin, rest, backendKind)
}

func workersDefault() int {
	n := runtime.NumCPU()
	if n < 1 {
		return 1
	}
	return n
}

func runForward(ctx context.Context, stdin io.Reader, args []string, kind backend.BackendKind) error {
	d := backend.NewDispatchDriver(kind)
	out, err := d.CommandContext(ctx, stdin, args...)
	if _, werr := os.Stdout.Write(out); werr != nil {
		return werr
	}
	if err == nil {
		return nil
	}
	var exitErr *wasmexif.ExitError
	if errors.As(err, &exitErr) {
		return exitError{code: exitErr.Code}
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return exitError{code: ee.ExitCode(), stderr: ee.Stderr}
	}
	return err
}

func runIngest(cfg cliConfig, stdin *os.File) error {
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

	if err := pipeline.Run(context.Background(), dirs, dbPath, cfg.workers); err != nil {
		return err
	}

	slog.Debug("run exit", "error", nil)
	return nil
}

func parseFlags(args []string) (cliConfig, error) {
	slog.Debug("parseFlags enter", "args_count", len(args))

	wd := workersDefault()

	flagSet := flag.NewFlagSet("exiftool-go", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	shortLevel := flagSet.String("l", defaultLogLevel, "log level")
	longLevel := flagSet.String("log-level", "", "log level")
	shortWorkers := flagSet.Int("w", wd, "worker count")
	longWorkers := flagSet.Int("workers", 0, "worker count")
	flagSet.Usage = func() {
		_, _ = fmt.Fprintln(flagSet.Output(), usageMessage(wd))
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
	return fmt.Sprintf("usage: exiftool-go pipeline [-l LEVEL|--log-level LEVEL] [-w N|--workers N] [DIR ...]\n  LEVEL: DEBUG, INFO, WARN, ERROR (default: %s)\n  N: worker count >= 1 (default: %d)", defaultLogLevel, defaultWorkers)
}

func globalUsage(defaultWorkers int) string {
	return usageMessage(defaultWorkers) + `

Default behavior forwards argv to ExifTool (wasm backend unless --backend=native):
  exiftool-go [EXIFTOOL_ARGS...]

Ingest (scan directories into SQLite) requires the pipeline subcommand:
  exiftool-go pipeline [ingest flags] [DIR ...]

Reserved before forward: pipeline subcommand, --backend=..., --help/--version.
`
}
