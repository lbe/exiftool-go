package golden

import (
	"encoding/json"
	"strings"
)

// NormalizeExiftoolJSONGolden rewrites exiftool -json stdout to a stable canonical form.
//
// Rules (aligned with plan §6.3 “fuzzy compare” / upstream TestLib nearEnough intent):
//   - Parse the JSON array exiftool emits.
//   - Remove tags that depend on filesystem metadata or process access time:
//     FileModifyDate, FileAccessDate, FileInodeChangeDate, FilePermissions.
//   - Re-encode with json.MarshalIndent so object key order is sorted (Go json), giving
//     deterministic bytes for golden comparison.
//
// Add new drop keys only when a field is shown to vary across machines or checkouts;
// document each addition here and in testdata/golden/README.md.
func NormalizeExiftoolJSONGolden(in []byte) ([]byte, error) {
	var v []map[string]any
	if err := json.Unmarshal(in, &v); err != nil {
		return nil, err
	}
	drop := map[string]struct{}{
		"FileModifyDate":      {},
		"FileAccessDate":      {},
		"FileInodeChangeDate": {},
		"FilePermissions":     {},
	}
	for _, obj := range v {
		for k := range drop {
			delete(obj, k)
		}
	}
	return json.MarshalIndent(v, "", "  ")
}

// NormalizeExiftoolTagListText rewrites exiftool plain-text tag listings (e.g. `-All`)
// to a stable form by dropping lines for filesystem-volatile fields. Rules mirror
// `NormalizeExiftoolJSONGolden`: same field names, line-oriented Pod tag list format.
func NormalizeExiftoolTagListText(in []byte) ([]byte, error) {
	s := strings.TrimSpace(string(in))
	if s == "" {
		return []byte{}, nil
	}
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		trim := strings.TrimSpace(line)
		if isVolatileExiftoolTagListLine(trim) {
			continue
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n") + "\n"), nil
}

func isVolatileExiftoolTagListLine(trim string) bool {
	for _, prefix := range []string{
		"File Modification Date/Time",
		"File Access Date/Time",
		"File Inode Change Date/Time",
		"File Permissions",
	} {
		if strings.HasPrefix(trim, prefix) {
			return true
		}
	}
	return false
}

// NormalizeExiftoolXMLGolden rewrites exiftool `-X` XML by removing one-line elements that
// embed unstable filesystem metadata (same semantics as tag-list / JSON goldens).
func NormalizeExiftoolXMLGolden(in []byte) ([]byte, error) {
	lines := strings.Split(string(in), "\n")
	var b strings.Builder
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.Contains(line, "System:FileModifyDate") ||
			strings.Contains(line, "System:FileAccessDate") ||
			strings.Contains(line, "System:FileInodeChangeDate") ||
			strings.Contains(line, "System:FilePermissions") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return []byte{}, nil
	}
	return []byte(out + "\n"), nil
}

// NormalizeExiftoolHeadBytes keeps the first n lines (newline-separated) and trims trailing
// space for `-listw`/`-listx` prefix goldens. n counts logical lines after trimming the final
// trailing newline from the input chunk.
func NormalizeExiftoolHeadBytes(in []byte, n int) ([]byte, error) {
	if n <= 0 {
		return nil, nil
	}
	s := strings.TrimSuffix(string(in), "\n")
	if s == "" {
		return []byte{}, nil
	}
	parts := strings.Split(s, "\n")
	if len(parts) > n {
		parts = parts[:n]
	}
	return []byte(strings.Join(parts, "\n") + "\n"), nil
}
