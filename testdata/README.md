# testdata/README.md

This directory contains test fixtures for EXIF extraction, pipeline validation, and CLI end-to-end coverage.

## Primary Rich EXIF Fixture

### Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif.jpg

This is the primary full-field EXIF-positive fixture.

All corrected tests should default to this fixture for EXIF-positive full-field validation.

Structured-field expectations sourced from Exif values in Metadata_test_file_-_includes_data_in_IIM,_XMP,_and_Exif-exiftool.txt:

- Camera Make: samsung
- Camera Model: SM-G930P
- DateTimeOriginal: 2017:05:29 11:11:16
- GPS Latitude: 26 deg 34' 57.06" N
- GPS Longitude: 80 deg 12' 0.84" W
- Image Width: 650
- Image Height: 488
- Orientation: not present in this fixture
- Exposure Time: 1/1600
- F Number: 1.7
- ISO: 40
- Focal Length: 4.2 mm

Notes:

- Tests must prefer Exif-derived values for structured columns when XMP or IPTC/IIM values differ.
- Known Exif-vs-XMP/IPTC mismatches in this fixture:
	- Caption/description: Exif uses Image Description, XMP uses Description, IPTC/IIM uses Caption-Abstract, and values differ.
	- Creator/author: Exif uses Artist, XMP uses Creator, IPTC/IIM uses By-line, and values differ.
	- Copyright: Exif uses Copyright, XMP uses Rights, IPTC/IIM uses Copyright Notice, and values differ.
- This fixture is the source of truth for full structured-field assertions except Orientation, which is covered by synthetic mapper tests because no bundled image fixture currently exposes it.

## Secondary EXIF Fixtures (path diversity)

### with_exif/nested/another.jpg

Byte-for-byte copy of `testdata/images/jpeg/samsung_sm_g930p.jpg` (same Samsung SM-G930P metadata as the primary Commons fixture). Kept under a nested directory so scanner and pipeline tests exercise recursion and non-root paths.

### with_exif/photo.jpg

Same bytes as `samsung_sm_g930p.jpg` / `another.jpg` (Samsung SM-G930P). Used as a second shallow `with_exif/` path without implying a different camera body.

### images/jpeg/samsung_sm_g930p_altpath.jpg

Same bytes as `samsung_sm_g930p.jpg`; lives under `testdata/images/jpeg/` for alternate relative-path coverage (replaces a former placeholder named `nikon_d3500.jpg`).

## Phase G R8 mixed-format corpus fixtures

### images/png/sample.png

Byte-for-byte copy of tracked `testdata/sample.png`, added under `testdata/images/png/` for format/path corpus coverage.

### images/tiff/sample.tiff

Byte-for-byte copy of tracked `testdata/sample.tiff`, added under `testdata/images/tiff/` for format/path corpus coverage.

### images/webp/sample.webp

WebP conversion derived from tracked `testdata/sample.png` (same image content, transcoded for cross-format corpus coverage).

### cli/argfiles/r8_mixed_formats.args

Argfile fixture listing the three mixed-format corpus paths above for `-@` CLI scenario coverage.

### Minimum JPEG camera diversity anchor

`testdata/sample.jpg` provides a second camera signature (`TestMaker` / `TestModel`) distinct from Samsung SM-G930P fixtures and is used by R8 corpus checks to keep the minimum JPEG corpus diversity requirement explicit.

## Non-EXIF and Error Fixtures

### no_exif/blank.jpg

Minimal JPEG (`testdata/sample_noexif.jpg` copy) with no camera EXIF payload. Pipeline classifies as `no_exif` (no Make/Model in ExifTool JSON).

### not_image/readme.txt

- Not an image file and should be skipped

### corrupt/broken.jpg

- Corrupt image and should produce parse_error or no_exif depending on extractor behavior
