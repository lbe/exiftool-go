# exiftool-go

Go CLI and libraries around **ExifTool**: **ingest** (scan images → EXIF in SQLite) and **argv-first passthrough** to ExifTool with a **unified runtime** (native subprocess + in-process wasm in one binary).

## Features

- **`internal/backend`**: single **`Driver`** API implemented by **`NativeDriver`** (ncruces), **`WasmDriver`** (go-exiftool-wasm), and **`DispatchDriver`** (runtime `BackendKind`).
- **CLI** (`cmd/exiftool-go`): default passthrough backend **`wasm`**; **`--backend=native`** or **`--backend=wasm`**; **`pipeline`** subcommand; legacy ingest argv.
- **Goldens** (`internal/golden`): dual-backend parity vs **`testdata/golden/`**; Pod-style matrices and corpus helpers.

## Quick start

```bash
make build          # ./bin/exiftool-go
make test           # go test ./... (native tests need exiftool on PATH or EXIFTOOL_GO_EXIFTOOL)
```

See **`ARCHITECTURE.md`** for diagrams and module layout. Planning notes may exist locally where your clone tracks them.

## Repository

| Path | Notes |
|------|--------|
| `cmd/exiftool-go` | Binary: ingest + passthrough |
| `internal/backend` | ExifTool driver adapters only |
| `internal/pipeline` | Ingest orchestration |
| `internal/exifutil` | JSON extraction helpers |
| `internal/golden` | CLI integration tests |
| `scripts/exiftoolgen` | Regenerate goldens (pinned ExifTool) |
| `testdata/golden` | Expected CLI outputs |

Module: **`github.com/lbe/exiftool-go`**.
