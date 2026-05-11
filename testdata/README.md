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

## Secondary EXIF Fixture

### with_exif/nested/another.jpg

This is a narrower EXIF-positive fixture retained for recursion and nested-path coverage.

Expected subset:

- Camera Make: Nikon
- Camera Model: D3500
- DateTimeOriginal: 2021-06-15T08:00:00

This fixture is not authoritative for GPS, dimensions, exposure, ISO, focal length, or orientation.

## Legacy Placeholder Fixture

### with_exif/photo.jpg

This file remains a placeholder fixture with inline pseudo-EXIF tags. It is no longer the primary evidence for EXIF completeness.

Expected subset:

- Camera Make: Canon
- Camera Model: EOS 80D
- DateTimeOriginal: 2022-01-01T12:34:56

## Non-EXIF and Error Fixtures

### no_exif/blank.jpg

- No EXIF data

### not_image/readme.txt

- Not an image file and should be skipped

### corrupt/broken.jpg

- Corrupt image and should produce parse_error or no_exif depending on extractor behavior
