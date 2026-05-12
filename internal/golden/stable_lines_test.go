package golden

import "testing"

func TestFilenameLinesStable(t *testing.T) {
	const in = "c\n  a  \n\nb\n"
	const want = "a\nb\nc\n"
	if got := FilenameLinesStable(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
