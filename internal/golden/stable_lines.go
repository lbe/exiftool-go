package golden

import (
	"sort"
	"strings"
)

// FilenameLinesStable returns newline-terminated sorted lines from s, trimming
// each line and dropping empties. Used for goldens where directory walk order
// may vary between hosts (e.g. exiftool -r -T -Filename).
func FilenameLinesStable(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	sort.Strings(out)
	return strings.Join(out, "\n") + "\n"
}
