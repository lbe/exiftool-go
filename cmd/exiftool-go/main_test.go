package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/lbe/exiftool-go/internal/input"
)

func TestParseFlags(t *testing.T) {
	t.Helper()

	cases := []struct {
		name    string
		args    []string
		want    cliConfig
		wantErr string
	}{
		{
			name: "short_log_level",
			args: []string{"-l", "DEBUG"},
			want: cliConfig{logLevel: "DEBUG", workers: workersDefault()},
		},
		{
			name: "long_log_level",
			args: []string{"--log-level", "WARN"},
			want: cliConfig{logLevel: "WARN", workers: workersDefault()},
		},
		{
			name: "worker_count",
			args: []string{"-w", "4"},
			want: cliConfig{logLevel: defaultLogLevel, workers: 4},
		},
		{
			name:    "invalid_log_level",
			args:    []string{"--log-level", "TRACE"},
			wantErr: "invalid log level",
		},
		{
			name:    "invalid_worker_count",
			args:    []string{"-w", "0"},
			wantErr: "invalid worker",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			got, err := parseFlags(tc.args)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("parseFlags: expected error, got nil")
				}
				if !strings.Contains(strings.ToLower(err.Error()), tc.wantErr) {
					t.Fatalf("parseFlags error: got %q", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFlags: %v", err)
			}
			if got.logLevel != tc.want.logLevel {
				t.Fatalf("logLevel: got %q, want %q", got.logLevel, tc.want.logLevel)
			}
			if got.workers != tc.want.workers {
				t.Fatalf("workers: got %d, want %d", got.workers, tc.want.workers)
			}
		})
	}
}

func TestRunIngestNoInputEmptyPipe(t *testing.T) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close write end: %v", err)
	}
	defer r.Close()

	err = run([]string{"pipeline"}, r)
	if !errors.Is(err, input.ErrNoDirectories) {
		t.Fatalf("run without input: got error %v, want %v", err, input.ErrNoDirectories)
	}
}
