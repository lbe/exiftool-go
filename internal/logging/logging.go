package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

const handlerTimeFormat = time.RFC3339

// ParseLevel converts a CLI log level string into the corresponding slog level.
func ParseLevel(level string) (slog.Level, error) {
	slog.Debug("ParseLevel enter", "level", level)

	normalizedLevel := strings.ToUpper(strings.TrimSpace(level))

	var parsed slog.Level
	switch normalizedLevel {
	case "DEBUG":
		parsed = slog.LevelDebug
	case "INFO", "":
		parsed = slog.LevelInfo
	case "WARN", "WARNING":
		parsed = slog.LevelWarn
	case "ERROR":
		parsed = slog.LevelError
	default:
		return 0, fmt.Errorf("invalid log level %q", level)
	}

	slog.Debug("ParseLevel exit", "level", level, "parsed", parsed)
	return parsed, nil
}

// Setup configures the process default logger with a text handler on stderr.
func Setup(level string) error {
	slog.Debug("Setup enter", "level", level)

	parsedLevel, err := ParseLevel(level)
	if err != nil {
		return err
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: parsedLevel,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, attr.Value.Time().Format(handlerTimeFormat))
			}
			return attr
		},
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Debug("Setup exit", "level", level, "parsed", parsedLevel)
	return nil
}
