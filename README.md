# exiftool-go

[![Go Reference](https://pkg.go.dev/badge/github.com/lbe/exiftool-go.svg)](https://pkg.go.dev/github.com/lbe/exiftool-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)](https://go.dev/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/lbe/exiftool-go)](https://goreportcard.com/report/github.com/lbe/exiftool-go)
[![CI](https://github.com/lbe/exiftool-go/actions/workflows/go.yml/badge.svg)](https://github.com/lbe/exiftool-go/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/lbe/exiftool-go/branch/main/graph/badge.svg)](https://codecov.io/gh/lbe/exiftool-go)

Go CLI and libraries around **ExifTool**: **ingest** (scan images → EXIF in SQLite) and **argv-first passthrough** to ExifTool with a **unified runtime** (native Perl subprocess + in-process wasm in one binary).

## Motivation

[ExifTool](https://exiftool.org/) is the reference tool for reading and writing image metadata, but the traditional deployment path assumes a Perl runtime and an `exiftool` script on `PATH`. That is fine for ad hoc use, yet awkward when you want a single portable binary, repeatable CI, or a Go program that calls ExifTool without shelling out to Perl by default.

**exiftool-go** addresses that by linking two drivers behind one **`internal/backend`** API:

- **Wasm (default):** in-process ExifTool via **`github.com/lbe/go-exiftool-wasm`** — no external Perl install required for passthrough or ingest.
- **Native:** full Perl subprocess parity via **`github.com/ncruces/go-exiftool`** when you need host plugins, `-use MODULE`, or other Tier-5 Perl-only behavior.

The same binary also runs an **ingest pipeline** that walks directories, hashes files, extracts JSON metadata, and upserts structured rows into SQLite — useful for building a local catalog or feeding downstream tools.

Dual-backend **golden tests** keep wasm and native CLI output aligned against pinned ExifTool **13.55** fixtures.

## Features

- **`internal/backend`**: single **`Driver`** API implemented by **`NativeDriver`** (ncruces), **`WasmDriver`** (go-exiftool-wasm), and **`DispatchDriver`** (runtime **`BackendKind`**).
- **CLI** (`cmd/exiftool-go`):
  - **Passthrough (default):** argv forwarded to ExifTool; default backend **`wasm`**; **`--backend=native`** or **`--backend=wasm`**.
  - **Ingest:** explicit **`pipeline`** subcommand only; always uses wasm internally; writes **`exif.db`** under XDG data dir.
- **Ingest libraries:** concurrent workers, SHA-256 hashing, JSON extraction, structured **`meta.ImageMeta`**, SQLite upsert with raw/structured consistency checks.
- **Goldens** (`internal/golden`): dual-backend parity vs **`testdata/golden/`**; Pod-style matrices and corpus helpers.

## Quick start

```bash
git clone https://github.com/lbe/exiftool-go.git
cd exiftool-go

make build          # ./bin/exiftool-go
make test           # unit tests only (fast)
make test-all       # unit + integration + e2e (needs EXIFTOOL_GO_EXIFTOOL or exiftool on PATH)
make cover          # coverage.out + coverage.html
```

### Passthrough (ExifTool argv)

```bash
# Default wasm backend — no exiftool on PATH required
./bin/exiftool-go -json -q testdata/images/jpeg/samsung_sm_g930p.jpg

# Native Perl driver (EXIFTOOL_GO_EXIFTOOL or exiftool on PATH)
./bin/exiftool-go --backend=native -ver
```

### Ingest (scan → SQLite)

```bash
# Explicit subcommand
./bin/exiftool-go pipeline -w 4 /path/to/photos
```

Database path: **`$XDG_DATA_HOME/exiftool-go/exif.db`** (or **`~/.local/share/exiftool-go/exif.db`**).

### Environment

| Variable | Purpose |
| -------- | ------- |
| `EXIFTOOL_GO_EXIFTOOL` | Path to Perl **`exiftool`** for **`--backend=native`** and native golden rows |
| `XDG_DATA_HOME` | Base directory for ingest database (see above) |

See **`ARCHITECTURE.md`** for diagrams, module layout, and data model. Planning notes may exist locally under **`/docs`** (gitignored in this repository).

## Repository

| Path | Notes |
|------|--------|
| `cmd/exiftool-go` | Binary: ingest + passthrough |
| `internal/backend` | ExifTool driver adapters only |
| `internal/pipeline` | Ingest orchestration |
| `internal/exifutil` | JSON extraction helpers |
| `internal/meta` | Shared metadata contract |
| `internal/db` | SQLite schema and upsert |
| `internal/golden` | CLI integration tests |
| `scripts/exiftoolgen` | Regenerate goldens (pinned ExifTool) |
| `testdata/golden` | Expected CLI outputs |

Module: **`github.com/lbe/exiftool-go`**.

## Development

```bash
make build    # ./bin/exiftool-go
make test             # unit tests only
make test-integration # -tags integration
make test-e2e         # -tags e2e
make test-all         # all tiers (needs exiftool for native/integration rows)
make cover            # coverage.out, coverage.html, summary line
make lint     # golangci-lint (local)
make fmt      # go fmt ./...
make check    # fmt, vet, test
```

Golden regeneration (requires ExifTool **13.55** on `PATH` or via `EXIFTOOL` / `EXIFTOOL_GO_EXIFTOOL`):

```bash
make golden-version
make golden-json
make golden-tree
```
