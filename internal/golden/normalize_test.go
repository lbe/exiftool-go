package golden

import (
	"strings"
	"testing"
)

func TestNormalizeExiftoolJSONGolden_dropsVolatileSystemTags(t *testing.T) {
	const in = `[
  {
    "SourceFile": "testdata/images/jpeg/x.jpg",
    "FileModifyDate": "2099:01:01 00:00:00+00:00",
    "FileAccessDate": "2099:01:01 00:00:00+00:00",
    "FileInodeChangeDate": "2099:01:01 00:00:00+00:00",
    "FilePermissions": "----------",
    "Make": "samsung",
    "Model": "SM-G930P"
  }
]`
	out, err := NormalizeExiftoolJSONGolden([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, key := range []string{"FileModifyDate", "FileAccessDate", "FileInodeChangeDate", "FilePermissions"} {
		if strings.Contains(s, key) {
			t.Errorf("expected %q removed from normalized JSON:\n%s", key, s)
		}
	}
	if !strings.Contains(s, `"Make": "samsung"`) {
		t.Fatalf("expected stable tag preserved:\n%s", s)
	}
}

func TestNormalizeExiftoolTagListText_dropsVolatileLines(t *testing.T) {
	const in = `File Name                       : a.jpg
File Modification Date/Time     : 2099:01:01 00:00:00+00:00
Image Width                     : 100
`
	out, err := NormalizeExiftoolTagListText([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "File Modification Date/Time") {
		t.Fatalf("expected volatile line removed:\n%s", s)
	}
	if !strings.Contains(s, "Image Width") {
		t.Fatalf("expected stable line kept:\n%s", s)
	}
}

func TestNormalizeExiftoolHeadBytes_truncates(t *testing.T) {
	in := "a\nb\nc\nd\n"
	out, err := NormalizeExiftoolHeadBytes([]byte(in), 2)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "a\nb\n" {
		t.Fatalf("got %q", string(out))
	}
}
