# Golden CLI outputs

Goldens are produced with **ExifTool 13.55** (see `testdata/golden/version.stdout`). Use the same major tool version when regenerating.

**Driver path:** use `EXIFTOOL_GO_EXIFTOOL` or `exiftool` on `PATH`. Do not point scripts or tests at personal planning-only corpus directories that are intentionally excluded from version control (see `.gitignore`).

- Go tests: `EXIFTOOL_GO_EXIFTOOL` pointing at the perl `exiftool` script, or `exiftool` on `PATH` (`internal/golden.DriverExiftool`).
- Regeneration script: `./scripts/exiftoolgen` uses `EXIFTOOL`, then `EXIFTOOL_GO_EXIFTOOL`, then `command -v exiftool`.

Environment for stable bytes: `TZ=UTC`, `LC_ALL=C` (documented per slice).

## `version.stdout`

From the repository root:

```sh
./scripts/exiftoolgen version
```

Manual equivalent (after choosing a driver path):

```sh
TZ=UTC LC_ALL=C /path/to/exiftool -ver > testdata/golden/version.stdout
```

## `json_samsung_sm_g930p.json`

Golden for `exiftool -json -q testdata/images/jpeg/samsung_sm_g930p.jpg` (run with `cmd.Dir` = repo root so `SourceFile` is stable).

Regenerate:

```sh
./scripts/exiftoolgen json-samsung
```

## JSON normalization (`internal/golden/normalize.go`)

Goldens store **normalized** `-json` output (see `NormalizeExiftoolJSONGolden`).

Dropped keys (volatile across checkouts / OS):

| Tag | Reason |
|-----|--------|
| `FileModifyDate` | git checkout / build changes mtime |
| `FileAccessDate` | read-time dependent |
| `FileInodeChangeDate` | inode metadata |
| `FilePermissions` | umask / platform differs |

After dropping, JSON is re-indented with two spaces (sorted map keys) for stable byte comparison.

### Plain-text tag lists (`NormalizeExiftoolTagListText`)

For `-All` and similar line-oriented listings, the same volatile filesystem fields are removed by matching line prefixes (`File Modification Date/Time`, `File Access Date/Time`, `File Inode Change Date/Time`, `File Permissions`).

### XML (`NormalizeExiftoolXMLGolden`)

For `-X` RDF output, one-line elements containing `System:FileModifyDate`, `System:FileAccessDate`, `System:FileInodeChangeDate`, or `System:FilePermissions` are dropped.

### List prefix (`NormalizeExiftoolHeadBytes`)

`-listw` / full `-listx` output is huge; prefix goldens compare only the first *n* logical lines after normalization.

## Pod matrix goldens (remediation Phase 1)

Regenerate from repo root with `TZ=UTC` and `LC_ALL=C`. Driver: `EXIFTOOL_GO_EXIFTOOL` or `exiftool` on `PATH` (13.55).

| File | Command |
|------|---------|
| `pod_sss_model_one_argv.stdout` | `exiftool -q -sss -Model testdata/images/jpeg/samsung_sm_g930p.jpg` |
| `pod_foo_unknown.stdout` | `exiftool -q -Foo testdata/tree/alpha/a.jpg` (empty stdout) |
| `pod_json_endopts_alpha.json` | `exiftool -json -q -- testdata/tree/alpha/a.jpg \| go run ./internal/golden/cmd/regenjson testdata/golden/pod_json_endopts_alpha.json` |
| `pod_read_model_short.stdout` | `exiftool -q -s -s -s -Model testdata/images/jpeg/samsung_sm_g930p.jpg` |
| `pod_all_alpha_text.stdout` | `exiftool -q -All testdata/tree/alpha/a.jpg` then apply the same line drops as `NormalizeExiftoolTagListText` (or rely on normalization-only compare) |
| `pod_xml_alpha.stdout` | `exiftool -X -q testdata/tree/alpha/a.jpg` then drop volatile `System:*` lines matching `NormalizeExiftoolXMLGolden` |
| `pod_table_model.stdout` | `exiftool -T -Model testdata/images/jpeg/samsung_sm_g930p.jpg` |
| `pod_short_capital_S_model.stdout` | `exiftool -S -Model testdata/images/jpeg/samsung_sm_g930p.jpg` |
| `pod_ext_jpg_tree.stdout` | `exiftool -r -ext jpg -T -Filename testdata/tree` |
| `pod_ext_plus_jpg_tree.stdout` | `exiftool -r -ext+ JPG -T -Filename testdata/tree \| sort` (`FilenameLinesStable` in tests) |
| `pod_if_imagewidth.stdout` | `exiftool -q -if '$imagewidth' -p '$filename' testdata/tree/alpha/a.jpg` |
| `pod_listw_head.stdout` | `exiftool -listw \| head -n 35` |
| `pod_listx_head.stdout` | `exiftool -listx \| head -n 28` |
| `pod_tagsfromfile_model.stdout` | line content `SM-G930P` after `-tagsFromFile` from a temp copy (see `TestExiftoolGoCLI_podMatrix_tagsFromFileModel`) |

Reuse: `json_samsung_sm_g930p.json` is shared by `-json`, `-JSON`, and `FILE -json -q` ordering cases.

## JPEG corpus (`testdata/images/jpeg/`)

- `samsung_sm_g930p.jpg` — copy of `testdata/Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif.jpg` (Samsung SM-G930P); used for JSON golden above.
- `samsung_sm_g930p_altpath.jpg` — same JPEG bytes as `samsung_sm_g930p.jpg`; reserved for additional goldens under a second relative path.

## `tree_r_filename_sorted.stdout`

Stable listing for `exiftool -r -T -Filename testdata/tree` (run with `cmd.Dir` = repo root).

Filenames are **sorted** in the golden and in tests (`FilenameLinesStable` in `internal/golden/stable_lines.go`) so traversal order differences between platforms do not break CI.

Regenerate:

```sh
./scripts/exiftoolgen tree-recurse
```

## `api_userparam_ver.stdout`

Golden for passthrough flags: `exiftool -ver -api RequestAll=3 -userParam Foo=Bar`.

Manual regenerate (after choosing a driver path):

```sh
TZ=UTC LC_ALL=C /path/to/exiftool -ver -api RequestAll=3 -userParam Foo=Bar > testdata/golden/api_userparam_ver.stdout
```

## `execute_sequence_ver.stdout`

Golden for execute-sequence parity: `exiftool -ver -execute -ver`.

Manual regenerate (after choosing a driver path):

```sh
TZ=UTC LC_ALL=C /path/to/exiftool -ver -execute -ver > testdata/golden/execute_sequence_ver.stdout
```

## `phase_g_r8_argfile.stdout`

Golden for mixed-format argfile passthrough coverage:

`exiftool -@ testdata/cli/argfiles/r8_mixed_formats.args -json -q`

Output is normalized through `regenjson` so JSON formatting/key-order expectations stay stable across platforms.

Manual regenerate (after choosing a driver path):

```sh
TZ=UTC LC_ALL=C /path/to/exiftool -@ testdata/cli/argfiles/r8_mixed_formats.args -json -q | go run ./internal/golden/cmd/regenjson testdata/golden/phase_g_r8_argfile.stdout
```

## `phase_g_t3_write_comment.stdout`

Golden for T3 write/original utility coverage: write a comment to a **temp copy** and assert stdout while leaving tracked fixtures unchanged.

Manual regenerate (after choosing a driver path):

```sh
tmp="$(mktemp -d)" && cp testdata/images/jpeg/samsung_sm_g930p.jpg "$tmp/write_original_input.jpg" && TZ=UTC LC_ALL=C /path/to/exiftool -Comment=phase-g-t3-red "$tmp/write_original_input.jpg" > testdata/golden/phase_g_t3_write_comment.stdout
```
