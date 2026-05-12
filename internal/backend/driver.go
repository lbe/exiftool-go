package backend

import (
	"context"
	"io"
)

// Server is the stay_open ExifTool session API shared by the native and wasm
// drivers (github.com/ncruces/go-exiftool and github.com/lbe/go-exiftool-wasm).
type Server interface {
	Close() error
	Command(arg ...string) ([]byte, error)
	Shutdown() error
}

// Driver is the unified ExifTool driver API for this repository: the same
// operations and signatures as github.com/ncruces/go-exiftool and
// github.com/lbe/go-exiftool-wasm.
type Driver interface {
	Command(stdin io.Reader, arg ...string) ([]byte, error)
	CommandContext(ctx context.Context, stdin io.Reader, arg ...string) ([]byte, error)
	Unmarshal(data []byte, m map[string][]byte) error
	NewServer(commonArg ...string) (Server, error)
}
