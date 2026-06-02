package input

import (
	"bufio"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestResolveDirsReturnsArgsWhenProvided(t *testing.T) {
	args := []string{"/photos/one", "/photos/two"}

	got, err := ResolveDirs(args, os.Stdin)
	if err != nil {
		t.Fatalf("ResolveDirs returned unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, args) {
		t.Fatalf("ResolveDirs: got %v, want %v", got, args)
	}
}

func TestResolveDirsPipeReadsLines(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	want := []string{"/photos/from-pipe", "/photos/also-pipe"}

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

func TestResolveDirsPipeSkipsBlankLines(t *testing.T) {
	stdin := strings.NewReader("\n  /photos/one  \n\n/photos/two\n")

	got, err := ResolveDirs(nil, fileFromReader(t, stdin))
	if err != nil {
		t.Fatalf("ResolveDirs returned unexpected error: %v", err)
	}
	want := []string{"/photos/one", "/photos/two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveDirs: got %v, want %v", got, want)
	}
}

func TestResolveDirsEmptyPipeReturnsErrNoDirectories(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close write end: %v", err)
	}

	_, err = ResolveDirs(nil, r)
	if !errors.Is(err, ErrNoDirectories) {
		t.Fatalf("ResolveDirs: got error %v, want ErrNoDirectories", err)
	}
}

func TestResolveDirsWhitespaceOnlyPipeReturnsErrNoDirectories(t *testing.T) {
	stdin := strings.NewReader("\n   \n")

	_, err := ResolveDirs(nil, fileFromReader(t, stdin))
	if !errors.Is(err, ErrNoDirectories) {
		t.Fatalf("ResolveDirs: got error %v, want ErrNoDirectories", err)
	}
}

func TestResolveDirsTerminalNoArgsReturnsErrNoDirectories(t *testing.T) {
	orig := terminalChecker
	defer func() { terminalChecker = orig }()
	called := false
	terminalChecker = func(_ *os.File) (bool, error) {
		called = true
		return true, nil
	}

	_, err := ResolveDirs(nil, os.Stdin)
	if !errors.Is(err, ErrNoDirectories) {
		t.Fatalf("ResolveDirs: got error %v, want ErrNoDirectories", err)
	}
	if !called {
		t.Fatal("ResolveDirs: terminalChecker was not called when no args are given")
	}
}

func TestResolveDirsTerminalCheckerError(t *testing.T) {
	orig := terminalChecker
	defer func() { terminalChecker = orig }()
	checkErr := errors.New("stat failed")
	terminalChecker = func(_ *os.File) (bool, error) {
		return false, checkErr
	}

	_, err := ResolveDirs(nil, os.Stdin)
	if !errors.Is(err, checkErr) {
		t.Fatalf("ResolveDirs: got error %v, want %v", err, checkErr)
	}
}

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

func fileFromReader(t *testing.T, r io.Reader) *os.File {
	t.Helper()

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(pipeW, r)
		closeErr := pipeW.Close()
		if copyErr != nil {
			done <- copyErr
			return
		}
		done <- closeErr
	}()

	if err := <-done; err != nil {
		t.Fatalf("populate pipe: %v", err)
	}

	return pipeR
}
