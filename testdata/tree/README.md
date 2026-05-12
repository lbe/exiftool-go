# `testdata/tree`

Nested directories with **mixed extensions** for recurse / `-ext` scenarios (plan §5.5, §6.3).

| Path | Role |
|------|------|
| `alpha/a.jpg` | JPEG without camera EXIF (same bytes as `testdata/sample_noexif.jpg` / `testdata/no_exif/blank.jpg`) |
| `alpha/nested/b.jpg` | JPEG (copied from `testdata/sample.jpg`) |
| `bravo/c.jpg` | JPEG (same as `b.jpg`) |
| `bravo/notes.txt` | Non-image text file |

Goldens: `testdata/golden/tree_r_filename_sorted.stdout` from `./scripts/exiftoolgen tree-recurse`.
