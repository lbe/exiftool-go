package backend

import (
	"context"
	"io"

	wasm "github.com/lbe/go-exiftool-wasm"
)

// WasmDriver forwards Driver methods to github.com/lbe/go-exiftool-wasm.
// A failed command returns the library's opaque error; it carries no exit code.
type WasmDriver struct{}

func (WasmDriver) Command(stdin io.Reader, arg ...string) ([]byte, error) {
	return wasm.Command(stdin, arg...)
}

func (WasmDriver) CommandContext(ctx context.Context, stdin io.Reader, arg ...string) ([]byte, error) {
	return wasm.CommandContext(ctx, stdin, arg...)
}

func (WasmDriver) Unmarshal(data []byte, m map[string][]byte) error {
	return wasm.Unmarshal(data, m)
}

func (WasmDriver) NewServer(commonArg ...string) (Server, error) {
	return wasm.NewServer(commonArg...)
}
