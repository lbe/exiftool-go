package logging

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	t.Helper()

	tests := []struct {
		name    string
		input   string
		want    slog.Level
		wantErr bool
	}{
		{name: "debug", input: "DEBUG", want: slog.LevelDebug},
		{name: "info", input: "INFO", want: slog.LevelInfo},
		{name: "warn", input: "WARN", want: slog.LevelWarn},
		{name: "error", input: "ERROR", want: slog.LevelError},
		{name: "case insensitive", input: "debug", want: slog.LevelDebug},
		{name: "invalid", input: "trace", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := ParseLevel(testCase.input)
			if testCase.wantErr {
				if err == nil {
					t.Fatal("ParseLevel returned nil error for invalid level")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseLevel returned error: %v", err)
			}
			if got != testCase.want {
				t.Fatalf("ParseLevel(%q): got %v, want %v", testCase.input, got, testCase.want)
			}
		})
	}
}
