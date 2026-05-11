package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lbe/exiftool-go/internal/meta"

	_ "modernc.org/sqlite"
)

const (
	sqlitePragmaJournalMode = "PRAGMA journal_mode=WAL;"
	sqlitePragmaBusyTimeout = "PRAGMA busy_timeout=5000;"
	sqlitePragmaSynchronous = "PRAGMA synchronous=NORMAL;"
	sqlitePragmaCacheSize   = "PRAGMA cache_size=-64000;"
	sqlitePragmaForeignKeys = "PRAGMA foreign_keys=ON;"
)

var sqlitePragmas = []string{
	sqlitePragmaJournalMode,
	sqlitePragmaBusyTimeout,
	sqlitePragmaSynchronous,
	sqlitePragmaCacheSize,
	sqlitePragmaForeignKeys,
}

const imagesSchemaSQL = `
CREATE TABLE IF NOT EXISTS images (
	id INTEGER PRIMARY KEY,
	file_path TEXT NOT NULL UNIQUE,
	file_hash TEXT NOT NULL,
	camera_make TEXT,
	camera_model TEXT,
	date_time TEXT,
	gps_latitude REAL,
	gps_longitude REAL,
	image_width INTEGER,
	image_height INTEGER,
	orientation INTEGER,
	exposure_time TEXT,
	f_number TEXT,
	iso_speed INTEGER,
	focal_length TEXT,
	exif_json TEXT CHECK(json_valid(exif_json)),
	exif_status TEXT NOT NULL CHECK(exif_status IN ('ok', 'no_exif', 'parse_error', 'hash_error', 'unsupported')),
	last_error TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	CHECK(last_error IS NULL OR exif_status != 'ok')
);
CREATE INDEX IF NOT EXISTS idx_images_camera_make ON images(camera_make);
CREATE INDEX IF NOT EXISTS idx_images_camera_model ON images(camera_model);
CREATE INDEX IF NOT EXISTS idx_images_date_time ON images(date_time);
CREATE INDEX IF NOT EXISTS idx_images_gps_lat_lon ON images(gps_latitude, gps_longitude);
CREATE INDEX IF NOT EXISTS idx_images_iso_speed ON images(iso_speed);
CREATE INDEX IF NOT EXISTS idx_images_exif_status ON images(exif_status);
`

const insertImageSQL = `
INSERT INTO images (
	file_path,
	file_hash,
	camera_make,
	camera_model,
	date_time,
	gps_latitude,
	gps_longitude,
	image_width,
	image_height,
	orientation,
	exposure_time,
	f_number,
	iso_speed,
	focal_length,
	exif_json,
	exif_status,
	last_error,
	created_at,
	updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(file_path) DO UPDATE SET
	file_hash = excluded.file_hash,
	camera_make = excluded.camera_make,
	camera_model = excluded.camera_model,
	date_time = excluded.date_time,
	gps_latitude = excluded.gps_latitude,
	gps_longitude = excluded.gps_longitude,
	image_width = excluded.image_width,
	image_height = excluded.image_height,
	orientation = excluded.orientation,
	exposure_time = excluded.exposure_time,
	f_number = excluded.f_number,
	iso_speed = excluded.iso_speed,
	focal_length = excluded.focal_length,
	exif_json = excluded.exif_json,
	exif_status = excluded.exif_status,
	last_error = excluded.last_error,
	updated_at = excluded.updated_at
`

// Open opens the SQLite database, applies the required pragmas, and ensures the schema exists.
func Open(path string) (*sql.DB, error) {
	slog.Debug("Open enter", "path", path)

	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", path, err)
	}

	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if err := applyPragmas(database); err != nil {
		_ = database.Close()
		return nil, err
	}
	if err := ensureSchema(database); err != nil {
		_ = database.Close()
		return nil, err
	}

	slog.Debug("Open exit", "path", path)
	return database, nil
}

func applyPragmas(database *sql.DB) error {
	slog.Debug("applyPragmas enter")

	for _, statement := range sqlitePragmas {
		if _, err := database.Exec(statement); err != nil {
			return fmt.Errorf("apply pragma %q: %w", statement, err)
		}
	}

	slog.Debug("applyPragmas exit")
	return nil
}

func ensureSchema(database *sql.DB) error {
	slog.Debug("ensureSchema enter")

	if _, err := database.Exec(imagesSchemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	slog.Debug("ensureSchema exit")
	return nil
}

// InsertImage stores a single image metadata record in the database.
func InsertImage(database *sql.DB, metadata meta.ImageMeta) error {
	slog.Debug("InsertImage enter", "file_path", metadata.FilePath)

	if err := validateJSONStructuredConsistency(metadata); err != nil {
		return err
	}

	args, err := buildInsertImageArgs(metadata)
	if err != nil {
		return err
	}

	if _, err := database.Exec(insertImageSQL, args...); err != nil {
		return fmt.Errorf("insert image %q: %w", metadata.FilePath, err)
	}

	slog.Debug("InsertImage exit", "file_path", metadata.FilePath)
	return nil
}

func validateJSONStructuredConsistency(metadata meta.ImageMeta) error {
	if metadata.RawEXIF == nil {
		return nil
	}

	if rawContainsAny(metadata.RawEXIF, "Make") && metadata.CameraMake == nil {
		return fmt.Errorf("raw_exif contains Make but CameraMake is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "Model", "CameraModelName") && metadata.CameraModel == nil {
		return fmt.Errorf("raw_exif contains Model/CameraModelName but CameraModel is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "DateTimeOriginal") && metadata.DateTime == nil {
		return fmt.Errorf("raw_exif contains DateTimeOriginal but DateTime is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "GPSLatitude") && metadata.GPSLatitude == nil {
		return fmt.Errorf("raw_exif contains GPSLatitude but GPSLatitude is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "GPSLongitude") && metadata.GPSLongitude == nil {
		return fmt.Errorf("raw_exif contains GPSLongitude but GPSLongitude is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "ImageWidth", "ExifImageWidth") && metadata.ImageWidth == nil {
		return fmt.Errorf("raw_exif contains ImageWidth/ExifImageWidth but ImageWidth is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "ImageHeight", "ExifImageHeight") && metadata.ImageHeight == nil {
		return fmt.Errorf("raw_exif contains ImageHeight/ExifImageHeight but ImageHeight is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "Orientation") && metadata.Orientation == nil {
		return fmt.Errorf("raw_exif contains Orientation but Orientation is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "ExposureTime") && metadata.ExposureTime == nil {
		return fmt.Errorf("raw_exif contains ExposureTime but ExposureTime is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "FNumber") && metadata.FNumber == nil {
		return fmt.Errorf("raw_exif contains FNumber but FNumber is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "ISO", "ISOSpeedRatings") && metadata.ISOSpeed == nil {
		return fmt.Errorf("raw_exif contains ISO/ISOSpeedRatings but ISOSpeed is nil")
	}
	if rawContainsAny(metadata.RawEXIF, "FocalLength") && metadata.FocalLength == nil {
		return fmt.Errorf("raw_exif contains FocalLength but FocalLength is nil")
	}

	return nil
}

func rawContainsAny(raw map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := raw[key]; ok {
			return true
		}
	}
	return false
}

func buildInsertImageArgs(metadata meta.ImageMeta) ([]any, error) {
	slog.Debug("buildInsertImageArgs enter", "file_path", metadata.FilePath)

	exifJSON, err := serializeRawEXIF(metadata)
	if err != nil {
		return nil, err
	}

	nowUTC := time.Now().UTC().Format(time.RFC3339)

	args := []any{
		metadata.FilePath,
		metadata.FileHash,
		nullableString(metadata.CameraMake),
		nullableString(metadata.CameraModel),
		nullableTime(metadata.DateTime),
		nullableFloat64(metadata.GPSLatitude),
		nullableFloat64(metadata.GPSLongitude),
		nullableInt(metadata.ImageWidth),
		nullableInt(metadata.ImageHeight),
		nullableInt(metadata.Orientation),
		nullableString(metadata.ExposureTime),
		nullableString(metadata.FNumber),
		nullableInt(metadata.ISOSpeed),
		nullableString(metadata.FocalLength),
		exifJSON,
		normalizeEXIFStatus(metadata.EXIFStatus),
		nullableString(metadata.LastError),
		nowUTC,
		nowUTC,
	}

	slog.Debug("buildInsertImageArgs exit", "file_path", metadata.FilePath)
	return args, nil
}

func serializeRawEXIF(metadata meta.ImageMeta) (string, error) {
	slog.Debug("serializeRawEXIF enter", "file_path", metadata.FilePath)

	exifJSON, err := json.Marshal(metadata.RawEXIF)
	if err != nil {
		return "", fmt.Errorf("marshal exif json for %q: %w", metadata.FilePath, err)
	}
	if len(exifJSON) == 0 {
		exifJSON = []byte("{}")
	}

	slog.Debug("serializeRawEXIF exit", "file_path", metadata.FilePath)
	return string(exifJSON), nil
}

func normalizeEXIFStatus(status string) string {
	slog.Debug("normalizeEXIFStatus enter", "status", status)

	if status == "" {
		slog.Debug("normalizeEXIFStatus exit", "status", meta.EXIFStatusUnsupported)
		return meta.EXIFStatusUnsupported
	}

	slog.Debug("normalizeEXIFStatus exit", "status", status)
	return status
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableFloat64(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}
