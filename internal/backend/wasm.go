package backend

import (
	"context"
	"io"
	"strings"

	wasm "github.com/lbe/go-exiftool-wasm"
)

// WasmDriver forwards Driver methods to github.com/lbe/go-exiftool-wasm,
// including stderr normalization on success paths to match native exec.Output
// semantics for warning-only guest stderr.
type WasmDriver struct{}

func (WasmDriver) Command(stdin io.Reader, arg ...string) ([]byte, error) {
	out, err := wasm.Command(stdin, arg...)
	return normalizeWasmSuccessStderr(out, err)
}

func (WasmDriver) CommandContext(ctx context.Context, stdin io.Reader, arg ...string) ([]byte, error) {
	out, err := wasm.CommandContext(ctx, stdin, arg...)
	return normalizeWasmSuccessStderr(out, err)
}

func (WasmDriver) Unmarshal(data []byte, m map[string][]byte) error {
	return wasm.Unmarshal(data, m)
}

func (WasmDriver) NewServer(commonArg ...string) (Server, error) {
	return wasm.NewServer(commonArg...)
}

const exiftoolWasmStderrErrPrefix = "exiftool: "

func normalizeWasmSuccessStderr(out []byte, err error) ([]byte, error) {
	if err == nil {
		return out, nil
	}
	if isExiftoolWasmWarningsOnlyStderr(err) {
		return out, nil
	}
	return out, err
}

func isExiftoolWasmWarningsOnlyStderr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, exiftoolWasmStderrErrPrefix) {
		return false
	}
	body := strings.TrimSpace(strings.TrimPrefix(msg, exiftoolWasmStderrErrPrefix))
	if body == "" {
		return false
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "Warning:") {
			return false
		}
	}
	return true
}
