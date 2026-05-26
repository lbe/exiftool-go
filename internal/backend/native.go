package backend

import (
	"context"
	"io"
	"os"
	"strings"

	native "github.com/ncruces/go-exiftool"
)

// NativeDriver forwards Driver methods to github.com/ncruces/go-exiftool.
type NativeDriver struct{}

func (NativeDriver) Command(stdin io.Reader, arg ...string) ([]byte, error) {
	configureNativeExecFromEnv()
	return native.Command(stdin, arg...)
}

func (NativeDriver) CommandContext(ctx context.Context, stdin io.Reader, arg ...string) ([]byte, error) {
	configureNativeExecFromEnv()
	return native.CommandContext(ctx, stdin, arg...)
}

func (NativeDriver) Unmarshal(data []byte, m map[string][]byte) error {
	return native.Unmarshal(data, m)
}

func (NativeDriver) NewServer(commonArg ...string) (Server, error) {
	configureNativeExecFromEnv()
	return native.NewServer(commonArg...)
}

func configureNativeExecFromEnv() {
	if p := strings.TrimSpace(os.Getenv("EXIFTOOL_GO_EXIFTOOL")); p != "" {
		native.Exec = p
		return
	}
	native.Exec = "exiftool"
}
