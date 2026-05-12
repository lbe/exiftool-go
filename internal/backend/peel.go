package backend

import (
	"errors"
	"fmt"
	"strings"
)

// BackendKind selects the ExifTool runtime (native subprocess vs in-process wasm).
type BackendKind int

const (
	BackendNative BackendKind = iota
	BackendWasm
)

var (
	errBackendMissingValue = fmt.Errorf("--backend requires a value")
	errInvalidBackend      = errors.New("invalid --backend value")
)

// StripLeadingBackendFlags removes leading reserved --backend / --backend=…
// tokens from argv and returns the effective backend kind. When no such token
// is present, the kind is BackendWasm (CLI default).
func StripLeadingBackendFlags(args []string) (rest []string, kind BackendKind, err error) {
	kind = BackendWasm
	rest = args
	for len(rest) > 0 {
		a := rest[0]
		if strings.HasPrefix(a, "--backend=") {
			v := strings.TrimPrefix(a, "--backend=")
			k, perr := ParseBackendLiteral(v)
			if perr != nil {
				return nil, kind, perr
			}
			kind = k
			rest = rest[1:]
			continue
		}
		if a == "--backend" {
			if len(rest) < 2 {
				return nil, kind, errBackendMissingValue
			}
			k, perr := ParseBackendLiteral(rest[1])
			if perr != nil {
				return nil, kind, perr
			}
			kind = k
			rest = rest[2:]
			continue
		}
		break
	}
	return rest, kind, nil
}

func ParseBackendLiteral(v string) (BackendKind, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "native":
		return BackendNative, nil
	case "wasm":
		return BackendWasm, nil
	default:
		return BackendNative, fmt.Errorf("%w %q (supported: native, wasm)", errInvalidBackend, v)
	}
}
