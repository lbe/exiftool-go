package backend

import (
	"context"
	"io"

	native "github.com/ncruces/go-exiftool"
)

// NativeDriver forwards Driver methods to github.com/ncruces/go-exiftool.
type NativeDriver struct{}

func (NativeDriver) Command(stdin io.Reader, arg ...string) ([]byte, error) {
	return native.Command(stdin, arg...)
}

func (NativeDriver) CommandContext(ctx context.Context, stdin io.Reader, arg ...string) ([]byte, error) {
	return native.CommandContext(ctx, stdin, arg...)
}

func (NativeDriver) Unmarshal(data []byte, m map[string][]byte) error {
	return native.Unmarshal(data, m)
}

func (NativeDriver) NewServer(commonArg ...string) (Server, error) {
	return native.NewServer(commonArg...)
}
