package backend

import (
	"context"
	"io"
)

// DispatchDriver implements Driver by delegating to NativeDriver or WasmDriver
// according to BackendKind.
type DispatchDriver struct {
	kind   BackendKind
	native NativeDriver
	wasm   WasmDriver
}

// NewDispatchDriver returns a Driver that uses the given backend kind for every
// operation.
func NewDispatchDriver(kind BackendKind) *DispatchDriver {
	return &DispatchDriver{kind: kind}
}

func (d *DispatchDriver) active() Driver {
	if d.kind == BackendWasm {
		return &d.wasm
	}
	return &d.native
}

func (d *DispatchDriver) Command(stdin io.Reader, arg ...string) ([]byte, error) {
	return d.active().Command(stdin, arg...)
}

func (d *DispatchDriver) CommandContext(ctx context.Context, stdin io.Reader, arg ...string) ([]byte, error) {
	return d.active().CommandContext(ctx, stdin, arg...)
}

func (d *DispatchDriver) Unmarshal(data []byte, m map[string][]byte) error {
	return d.active().Unmarshal(data, m)
}

func (d *DispatchDriver) NewServer(commonArg ...string) (Server, error) {
	return d.active().NewServer(commonArg...)
}
