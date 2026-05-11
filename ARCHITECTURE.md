# Architecture: cmd/exiftool-go

This document describes the standalone CLI ingest app in `cmd/exiftool-go`.

## Goals

- Accept directories from CLI args or stdin pipe.
- Scan images, extract metadata, and store results in SQLite.
- Provide deterministic behavior under no-input and invalid-input cases.

## Module dependency baseline

This app is in the root module and uses dependency versions from the top-level go.mod.

Current direct module dependencies are:

- github.com/lbe/cfsread v0.1.0
- github.com/phsym/console-slog v0.3.1
- github.com/pierrec/lz4/v4 v4.1.26
- modernc.org/sqlite v1.50.0

## Execution flow

1. Parse flags (`-l/--log-level`, `-w/--workers`) in `main.go`.
2. Resolve directories from args or stdin in `internal/input`.
3. Resolve XDG data dir and open `exif.db` via `internal/db`.
4. Run concurrent scan/hash/extract/write orchestration in `internal/pipeline`.

## Data model

- Table: `images` (defined in `internal/db/db.go`).
- Upsert key: `file_path`.
- EXIF status values: `ok`, `no_exif`, `parse_error`, `hash_error`, `unsupported`.
- `exif_json` stores JSON object payloads; when EXIF is absent or cannot be parsed the persisted value is `{}`.
- A shared `ImageMeta` contract includes 12 structured EXIF fields plus `RawEXIF`/`exif_json` and status metadata.
- The app validates `RawEXIF` consistency before persisting: if raw EXIF contains a supported key, the corresponding structured `ImageMeta` field must be populated.

The primary E2E evidence path is the rich fixture `testdata/Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif.jpg` and its manifest text file. Tests assert both structured image metadata columns and raw `exif_json` values for the contract fields, with manifest-derived values as the source of truth.

## Verification coverage

- Unit tests: flags, input resolution, scanner/hash/db behavior, and pipeline logic.
- E2E tests: CLI args, stdin pipe, no-input TTY, empty pipe, missing directory, no-EXIF fixtures, corrupt fixtures, and `-l`/`-w` flags.
- Gates: `golangci-lint run ./...` and `go test ./...`.

## Relevant files

| File                          | Responsibility                                 |
| ----------------------------- | ---------------------------------------------- |
| cmd/exiftool-go/main.go       | CLI entrypoint, flags, app wiring              |
| cmd/exiftool-go/main_test.go  | Unit tests for parsing and entrypoint behavior |
| cmd/exiftool-go/e2e_test.go   | Binary-level end-to-end tests                  |
| internal/input/input.go       | Directory input resolution from args/stdin     |
| internal/pipeline/pipeline.go | Concurrent orchestration and status mapping    |
| internal/db/db.go             | SQLite schema and upsert persistence           |
