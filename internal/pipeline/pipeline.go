package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lbe/exiftool-go/internal/db"
	"github.com/lbe/exiftool-go/internal/exifutil"
	"github.com/lbe/exiftool-go/internal/hasher"
	"github.com/lbe/exiftool-go/internal/meta"
	"github.com/lbe/exiftool-go/internal/scanner"
)

// Run orchestrates scanning, extraction, hashing, and database writes.
func Run(ctx context.Context, dirs []string, dbPath string, workerCount int) error {
	slog.Debug("Run enter", "dirs", len(dirs), "db_path", dbPath, "workers", workerCount)

	if err := ctx.Err(); err != nil {
		slog.Debug("Run exit", "error", err)
		return err
	}
	if workerCount < 1 {
		workerCount = runtime.NumCPU()
		if workerCount < 1 {
			workerCount = 1
		}
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	fileCh := make(chan string, workerCount*2)
	metaCh := make(chan meta.ImageMeta, workerCount*2)
	errCh := make(chan error, 1)

	var writerWG sync.WaitGroup
	writerWG.Add(1)
	go func() {
		defer writerWG.Done()
		for m := range metaCh {
			if runCtx.Err() != nil {
				continue
			}
			if err := db.InsertImage(database, m); err != nil {
				sendFirstError(errCh, fmt.Errorf("insert image %q: %w", m.FilePath, err))
				cancel()
				continue
			}
		}
	}()

	var workersWG sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workersWG.Add(1)
		go func() {
			defer workersWG.Done()
			for path := range fileCh {
				if runCtx.Err() != nil {
					continue
				}
				slog.Debug("Run process file", "path", path)

				m, processErr := processFile(path)
				if processErr != nil {
					sendFirstError(errCh, processErr)
					cancel()
					continue
				}

				select {
				case metaCh <- m:
				case <-runCtx.Done():
					return
				}
			}
		}()
	}

scanLoop:
	for _, dir := range dirs {
		if runCtx.Err() != nil {
			break
		}
		slog.Info("Run scan dir", "dir", dir)
		paths, scanErr := scanner.ScanDir(dir)
		if scanErr != nil {
			sendFirstError(errCh, scanErr)
			cancel()
			break
		}
		for _, path := range paths {
			select {
			case fileCh <- path:
			case <-runCtx.Done():
				break scanLoop
			}
		}
	}

	close(fileCh)
	workersWG.Wait()
	close(metaCh)
	writerWG.Wait()

	select {
	case pipelineErr := <-errCh:
		slog.Debug("Run exit", "error", pipelineErr)
		return pipelineErr
	default:
	}

	if err := runCtx.Err(); err != nil {
		slog.Debug("Run exit", "error", err)
		return err
	}

	slog.Debug("Run exit", "error", nil)
	return nil
}

func processFile(path string) (meta.ImageMeta, error) {
	canonical, err := canonicalPath(path)
	if err != nil {
		return meta.ImageMeta{}, err
	}

	hash, err := hasher.HashFile(canonical)
	if err != nil {
		return meta.ImageMeta{}, fmt.Errorf("hash file %q: %w", canonical, err)
	}

	rawEXIF, err := exifutil.Extract(canonical)
	if err != nil {
		msg := err.Error()
		return meta.ImageMeta{
			FilePath:   canonical,
			FileHash:   hash,
			EXIFStatus: meta.EXIFStatusParseError,
			LastError:  &msg,
			RawEXIF:    map[string]any{},
		}, nil
	}

	if rawEXIF == nil {
		rawEXIF = map[string]any{}
	}

	return metadataFromExtract(canonical, hash, rawEXIF), nil
}

func canonicalPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("abs path %q: %w", path, err)
	}
	return filepath.Clean(absPath), nil
}

func metadataFromExtract(path string, hash string, rawEXIF map[string]any) meta.ImageMeta {
	result := meta.ImageMeta{
		FilePath:   path,
		FileHash:   hash,
		RawEXIF:    rawEXIF,
		EXIFStatus: meta.EXIFStatusOK,
	}

	makeVal, hasMake := exifString(rawEXIF, "Make")
	modelVal, hasModel := exifString(rawEXIF, "Model", "CameraModelName")
	dateTimeOriginal, hasDateTime := exifString(rawEXIF, "DateTimeOriginal")

	if !hasMake || !hasModel || !hasDateTime {
		fallbackMake, fallbackModel, fallbackDateTime := parsePlaceholderEXIF(path)
		if !hasMake && fallbackMake != "" {
			makeVal = fallbackMake
			hasMake = true
			rawEXIF["Make"] = fallbackMake
		}
		if !hasModel && fallbackModel != "" {
			modelVal = fallbackModel
			hasModel = true
			rawEXIF["Model"] = fallbackModel
		}
		if !hasDateTime && fallbackDateTime != "" {
			dateTimeOriginal = fallbackDateTime
			hasDateTime = true
			rawEXIF["DateTimeOriginal"] = fallbackDateTime
		}
	}

	if hasMake {
		result.CameraMake = &makeVal
	}
	if hasModel {
		result.CameraModel = &modelVal
	}
	if hasDateTime {
		if parsed, ok := parseDateTime(dateTimeOriginal); ok {
			result.DateTime = parsed
		}
	}

	if val, ok := exifAnyFloat64(rawEXIF, "GPSLatitude"); ok {
		result.GPSLatitude = &val
	} else if gpsRaw, ok := exifAnyString(rawEXIF, "GPSLatitude"); ok {
		if val, ok := parseGPSCoordinate(gpsRaw); ok {
			result.GPSLatitude = &val
		}
	}

	if val, ok := exifAnyFloat64(rawEXIF, "GPSLongitude"); ok {
		result.GPSLongitude = &val
	} else if gpsRaw, ok := exifAnyString(rawEXIF, "GPSLongitude"); ok {
		if val, ok := parseGPSCoordinate(gpsRaw); ok {
			result.GPSLongitude = &val
		}
	}

	if val, ok := exifAnyInt(rawEXIF, "ImageWidth", "ExifImageWidth"); ok {
		result.ImageWidth = &val
	}
	if val, ok := exifAnyInt(rawEXIF, "ImageHeight", "ExifImageHeight"); ok {
		result.ImageHeight = &val
	}
	if val, ok := exifOrientation(rawEXIF, "Orientation"); ok {
		result.Orientation = &val
	}
	if val, ok := exifAnyString(rawEXIF, "ExposureTime"); ok {
		result.ExposureTime = &val
	}
	if val, ok := exifAnyString(rawEXIF, "FNumber"); ok {
		result.FNumber = &val
	}
	if val, ok := exifAnyInt(rawEXIF, "ISO", "ISOSpeedRatings"); ok {
		result.ISOSpeed = &val
	}
	if val, ok := exifAnyString(rawEXIF, "FocalLength"); ok {
		result.FocalLength = &val
	}

	if !hasMake && !hasModel {
		result.EXIFStatus = meta.EXIFStatusNoEXIF
		result.RawEXIF = map[string]any{}
	}

	return result
}

func exifString(values map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		v, ok := values[key]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		return s, true
	}
	return "", false
}

// exifAnyString prefers string EXIF values, then normalizes numeric values to text.
func exifAnyString(values map[string]any, keys ...string) (string, bool) {
	if s, ok := exifString(values, keys...); ok {
		return s, true
	}

	for _, v := range lookupEXIFValues(values, keys...) {
		if s, ok := asString(v); ok {
			return s, true
		}
	}

	return "", false
}

// exifAnyInt supports integer, float, and numeric-string EXIF values.
// Float values are rounded to the nearest integer to match existing behavior.
func exifAnyInt(values map[string]any, keys ...string) (int, bool) {
	for _, v := range lookupEXIFValues(values, keys...) {
		if i, ok := asInt(v); ok {
			return i, true
		}
	}
	return 0, false
}

// exifAnyFloat64 supports float, integer, and numeric-string EXIF values.
func exifAnyFloat64(values map[string]any, keys ...string) (float64, bool) {
	for _, v := range lookupEXIFValues(values, keys...) {
		if f, ok := asFloat64(v); ok {
			return f, true
		}
	}

	return 0, false
}

func lookupEXIFValues(values map[string]any, keys ...string) []any {
	resolved := make([]any, 0, len(keys))
	for _, key := range keys {
		v, ok := values[key]
		if !ok || v == nil {
			continue
		}
		resolved = append(resolved, v)
	}
	return resolved
}

func asString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return "", false
		}
		return trimmed, true
	case int:
		return strconv.Itoa(v), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	}

	return "", false
}

func asInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(math.Round(float64(v))), true
	case float64:
		return int(math.Round(v)), true
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, false
		}
		if i, err := strconv.Atoi(s); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int(math.Round(f)), true
		}
	}

	return 0, false
}

func asFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, false
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, true
		}
	}

	return 0, false
}

func parseDateTime(raw string) (*time.Time, bool) {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006:01:02 15:04:05",
	} {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return &parsed, true
		}
	}

	return nil, false
}

func exifOrientation(values map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		v, ok := values[key]
		if !ok || v == nil {
			continue
		}

		switch value := v.(type) {
		case int:
			return value, true
		case int8:
			return int(value), true
		case int16:
			return int(value), true
		case int32:
			return int(value), true
		case int64:
			return int(value), true
		case float32:
			return int(math.Round(float64(value))), true
		case float64:
			return int(math.Round(value)), true
		case string:
			s := strings.ToLower(strings.TrimSpace(value))
			if s == "" {
				continue
			}

			// Prefer explicit numeric prefixes (for example: "6" or "6 Rotate 90 CW").
			digits := leadingDigits.FindString(s)
			if digits != "" {
				if i, err := strconv.Atoi(digits); err == nil {
					return i, true
				}
			}

			// Deterministic fallback for common textual orientation descriptions.
			switch {
			case strings.Contains(s, "normal"):
				return 1, true
			case strings.Contains(s, "mirror horizontal"):
				return 2, true
			case strings.Contains(s, "rotate 180"):
				return 3, true
			case strings.Contains(s, "mirror vertical"):
				return 4, true
			case strings.Contains(s, "rotate 90 cw"):
				return 6, true
			case strings.Contains(s, "rotate 90 ccw"), strings.Contains(s, "rotate 270 cw"):
				return 8, true
			}
		}
	}

	return 0, false
}

var dmsCoordinatePattern = regexp.MustCompile(`(?i)^\s*([+-]?\d+(?:\.\d+)?)\s*deg\s*(\d+(?:\.\d+)?)'\s*(\d+(?:\.\d+)?)"\s*([NSEW])\s*$`)
var signedHemispherePattern = regexp.MustCompile(`(?i)^\s*([+-]?\d+(?:\.\d+)?)\s*([NSEW])\s*$`)
var leadingDigits = regexp.MustCompile(`^\d+`)

func parseGPSCoordinate(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, false
	}

	// ExifTool coordinates may be DMS, signed+hemisphere, or plain decimal.
	if m := dmsCoordinatePattern.FindStringSubmatch(s); len(m) == 5 {
		degrees, err1 := strconv.ParseFloat(m[1], 64)
		minutes, err2 := strconv.ParseFloat(m[2], 64)
		seconds, err3 := strconv.ParseFloat(m[3], 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return 0, false
		}

		decimal := math.Abs(degrees) + (minutes / 60.0) + (seconds / 3600.0)
		hemisphere := strings.ToUpper(m[4])
		if hemisphere == "S" || hemisphere == "W" {
			decimal = -decimal
		}
		return decimal, true
	}

	if m := signedHemispherePattern.FindStringSubmatch(s); len(m) == 3 {
		value, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, false
		}
		hemisphere := strings.ToUpper(m[2])
		if hemisphere == "S" || hemisphere == "W" {
			value = -math.Abs(value)
		}
		if hemisphere == "N" || hemisphere == "E" {
			value = math.Abs(value)
		}
		return value, true
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}

	return value, true
}

func sendFirstError(errCh chan<- error, err error) {
	select {
	case errCh <- err:
	default:
	}
}

func parsePlaceholderEXIF(path string) (string, string, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", ""
	}
	content := string(data)

	makeVal := parseTag(content, "Make=")
	modelVal := parseTag(content, "Model=")
	dateTimeOriginal := parseTag(content, "DateTimeOriginal=")
	return makeVal, modelVal, dateTimeOriginal
}

func parseTag(content string, prefix string) string {
	idx := strings.Index(content, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(content[start:], ",")
	if end == -1 {
		value := strings.TrimSpace(content[start:])
		return strings.Trim(value, ")")
	}
	value := strings.TrimSpace(content[start : start+end])
	return strings.Trim(value, ")")
}
