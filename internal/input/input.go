// Package input resolves the directories to process from CLI arguments or stdin.
package input

import (
	"bufio"
	"errors"
	"log/slog"
	"os"
	"strings"
)

// ErrNoDirectories is returned when no directory paths are available from args or stdin.
var ErrNoDirectories = errors.New("no directories provided")

// terminalChecker is the function used to detect whether a file is a TTY.
// It is a variable so tests can replace it with a mock.
var terminalChecker = func(f *os.File) (bool, error) {
	slog.Debug("terminalChecker enter")
	stat, err := f.Stat()
	if err != nil {
		slog.Debug("terminalChecker exit", "error", err)
		return false, err
	}
	isTerminal := (stat.Mode() & os.ModeCharDevice) != 0
	slog.Debug("terminalChecker exit", "is_terminal", isTerminal)
	return isTerminal, nil
}

// ResolveDirs returns the CLI args when they are present. When no args are
// given it calls terminalChecker and, when stdin is a pipe, reads
// newline-delimited directories from it. When stdin is a terminal it returns
// errNoDirectories without blocking.
func ResolveDirs(args []string, stdin *os.File) ([]string, error) {
	slog.Debug("ResolveDirs enter", "args_count", len(args))
	if len(args) > 0 {
		slog.Debug("ResolveDirs exit", "source", "args", "count", len(args))
		return args, nil
	}

	dirs, err := resolveFromStdin(stdin)
	if err != nil {
		slog.Debug("ResolveDirs exit", "source", "stdin", "error", err)
		return nil, err
	}

	slog.Debug("ResolveDirs exit", "source", "pipe", "count", len(dirs))
	return dirs, nil
}

func resolveFromStdin(stdin *os.File) ([]string, error) {
	isTerminal, err := terminalChecker(stdin)
	if err != nil {
		return nil, err
	}
	if isTerminal {
		return nil, ErrNoDirectories
	}

	return readDirs(stdin)
}

func readDirs(stdin *os.File) ([]string, error) {
	dirs := make([]string, 0)
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		dir := strings.TrimSpace(scanner.Text())
		if dir == "" {
			continue
		}
		dirs = append(dirs, dir)
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return nil, scanErr
	}
	if len(dirs) == 0 {
		return nil, ErrNoDirectories
	}
	return dirs, nil
}
