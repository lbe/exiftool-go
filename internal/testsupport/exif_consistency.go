package testsupport

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

func AssertRawEXIFJSONParsable(t *testing.T, raw map[string]any) map[string]any {
	t.Helper()

	blob, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal raw_exif: %v", err)
	}

	parsed := map[string]any{}
	if err := json.Unmarshal(blob, &parsed); err != nil {
		t.Fatalf("unmarshal raw_exif: %v", err)
	}

	return parsed
}

func AssertRawJSONHasAnyKey(t *testing.T, parsed map[string]any, keys ...string) {
	t.Helper()

	for _, key := range keys {
		if _, ok := parsed[key]; ok {
			return
		}
	}
	t.Fatalf("raw_exif JSON missing all candidate keys: %v", keys)
}

func AssertRawPresenceImpliesStructuredNonNil(t *testing.T, parsed map[string]any, rawKeys []string, structuredName string, value any) { //nolint:govet
	t.Helper()

	present := false
	for _, key := range rawKeys {
		if _, ok := parsed[key]; ok {
			present = true
			break
		}
	}
	if !present {
		t.Fatalf("raw_exif JSON missing keys for %s: %v", structuredName, rawKeys)
	}

	if value == nil {
		t.Fatalf("%s: got nil, want non-nil", structuredName)
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.IsNil() { //nolint:govet
		t.Fatalf("%s: got nil pointer, want non-nil", structuredName)
	}
}

func AssertRawPresenceImpliesDBNotNull(t *testing.T, parsed map[string]any, rawKeys []string, dbColumn string, dbValid bool) {
	t.Helper()

	for _, key := range rawKeys {
		if _, ok := parsed[key]; ok {
			if !dbValid {
				t.Fatalf("%s is NULL while exif_json contains %q", dbColumn, key)
			}
			return
		}
	}
}

func ParseEXIFJSON(t *testing.T, exifJSON string) map[string]any {
	t.Helper()

	parsed := map[string]any{}
	if err := json.Unmarshal([]byte(exifJSON), &parsed); err != nil {
		t.Fatalf("unmarshal exif_json: %v", err)
	}
	return parsed
}

func Float64ToString(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
