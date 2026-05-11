package input

import (
	"bufio"
	"os"
	"reflect"
	"testing"
)

// TestResolveDirsReturnsArgsWhenProvided verifies that when CLI arguments are
// present, ResolveDirs returns them directly without consulting stdin.
func TestResolveDirsReturnsArgsWhenProvided(t *testing.T) {
	args := []string{"/photos/one", "/photos/two"}

	// stdin state is irrelevant when args are present; pass os.Stdin directly.
	got, err := ResolveDirs(args, os.Stdin)
	if err != nil {
		t.Fatalf("ResolveDirs returned unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, args) {
		t.Fatalf("ResolveDirs: got %v, want %v", got, args)
	}
}

// TestResolveDirsPipeReadsLines verifies that when no CLI args are given and
// stdin is a pipe, ResolveDirs reads newline-delimited directory paths from it.
func TestResolveDirsPipeReadsLines(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	want := []string{"/photos/from-pipe", "/photos/also-pipe"}

	// Write lines into the write end and close so the reader sees EOF.
	go func() {
		defer w.Close()
		for _, dir := range want {
			if _, werr := w.WriteString(dir + "\n"); werr != nil {
				t.Errorf("write to pipe: %v", werr)
			}
		}
	}()

	got, err := ResolveDirs(nil, r)
	if err != nil {
		t.Fatalf("ResolveDirs returned unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveDirs: got %v, want %v", got, want)
	}
}

// TestResolveDirsTerminalNoArgsReturnsError verifies that when no CLI args are
// given and stdin is a terminal (TTY), ResolveDirs returns an error rather than
// blocking on stdin. It also verifies that terminalChecker is actually invoked,
// so this test fails (RED) until the GREEN implementation wires up the call.
func TestResolveDirsTerminalNoArgsReturnsError(t *testing.T) {
	// Inject a fake terminalChecker that records whether it was called.
	orig := terminalChecker
	defer func() { terminalChecker = orig }()
	called := false
	terminalChecker = func(_ *os.File) (bool, error) {
		called = true
		return true, nil
	}

	_, err := ResolveDirs(nil, os.Stdin)
	if err == nil {
		t.Fatal("ResolveDirs: expected error when stdin is a terminal with no args, got nil")
	}
	if !called {
		t.Fatal("ResolveDirs: terminalChecker was not called, but it must be when no args are given")
	}
}

// TestResolveDirsArgsWinOverPipe verifies that CLI args take precedence even
// when stdin is a pipe with data, and stdin is not consumed in that case.
func TestResolveDirsArgsWinOverPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()

	if _, werr := w.WriteString("/from/pipe\n"); werr != nil {
		t.Fatalf("write to pipe: %v", werr)
	}
	if cerr := w.Close(); cerr != nil {
		t.Fatalf("close write end: %v", cerr)
	}

	args := []string{"/from/args"}

	orig := terminalChecker
	defer func() { terminalChecker = orig }()
	called := false
	terminalChecker = func(_ *os.File) (bool, error) {
		called = true
		return false, nil
	}

	got, err := ResolveDirs(args, r)
	if err != nil {
		t.Fatalf("ResolveDirs returned unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, args) {
		t.Fatalf("ResolveDirs: got %v, want %v", got, args)
	}
	if called {
		t.Fatal("ResolveDirs: terminalChecker should not be called when args are provided")
	}

	// Args path should not consume piped stdin.
	s := bufio.NewScanner(r)
	if !s.Scan() {
		t.Fatal("expected pipe data to remain unread when args are provided")
	}
	if gotLine := s.Text(); gotLine != "/from/pipe" {
		t.Fatalf("pipe line got %q, want %q", gotLine, "/from/pipe")
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scan pipe after ResolveDirs: %v", err)
	}
}
