package meta

import "time"

const (
	// EXIFStatusOK marks metadata extracted successfully.
	EXIFStatusOK = "ok"
	// EXIFStatusNoEXIF marks files that are valid images without EXIF payloads.
	EXIFStatusNoEXIF = "no_exif"
	// EXIFStatusParseError marks files whose EXIF payload could not be parsed.
	EXIFStatusParseError = "parse_error"
	// EXIFStatusHashError marks files that could not be hashed.
	EXIFStatusHashError = "hash_error"
	// EXIFStatusUnsupported marks files that are not supported for EXIF extraction.
	EXIFStatusUnsupported = "unsupported"
)

// ImageMeta is the shared metadata contract passed between scanner, EXIF,
// hashing, and database packages.
type ImageMeta struct {
	FilePath     string         `json:"file_path,omitempty"`
	FileHash     string         `json:"file_hash,omitempty"`
	CameraMake   *string        `json:"camera_make,omitempty"`
	CameraModel  *string        `json:"camera_model,omitempty"`
	DateTime     *time.Time     `json:"date_time,omitempty"`
	GPSLatitude  *float64       `json:"gps_latitude,omitempty"`
	GPSLongitude *float64       `json:"gps_longitude,omitempty"`
	ImageWidth   *int           `json:"image_width,omitempty"`
	ImageHeight  *int           `json:"image_height,omitempty"`
	Orientation  *int           `json:"orientation,omitempty"`
	ExposureTime *string        `json:"exposure_time,omitempty"`
	FNumber      *string        `json:"f_number,omitempty"`
	ISOSpeed     *int           `json:"iso_speed,omitempty"`
	FocalLength  *string        `json:"focal_length,omitempty"`
	EXIFStatus   string         `json:"exif_status,omitempty"`
	LastError    *string        `json:"last_error,omitempty"`
	RawEXIF      map[string]any `json:"raw_exif,omitempty"`
}
