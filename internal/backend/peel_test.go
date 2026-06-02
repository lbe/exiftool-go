package backend

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseBackendLiteral(t *testing.T) {
	t.Helper()

	valid := []struct {
		in   string
		want BackendKind
	}{
		{"native", BackendNative},
		{"wasm", BackendWasm},
		{"NATIVE", BackendNative},
	}
	for _, tc := range valid {
		got, err := ParseBackendLiteral(tc.in)
		if err != nil {
			t.Fatalf("ParseBackendLiteral(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseBackendLiteral(%q): got %v, want %v", tc.in, got, tc.want)
		}
	}

	_, err := ParseBackendLiteral("nope")
	if err == nil {
		t.Fatal("expected error for invalid backend")
	}
	if !strings.Contains(err.Error(), "invalid --backend value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStripLeadingBackendFlagsNoFlagDefaultWasm(t *testing.T) {
	t.Helper()
	rest, kind, err := StripLeadingBackendFlags([]string{"-json", "a.jpg"})
	if err != nil {
		t.Fatalf("StripLeadingBackendFlags: %v", err)
	}
	if kind != BackendWasm {
		t.Fatalf("default backend kind: got %v, want BackendWasm", kind)
	}
	want := []string{"-json", "a.jpg"}
	if !reflect.DeepEqual(rest, want) {
		t.Fatalf("rest: got %#v, want %#v", rest, want)
	}
}

func TestStripLeadingBackendFlagsExplicitNative(t *testing.T) {
	t.Helper()
	rest, kind, err := StripLeadingBackendFlags([]string{"--backend=native", "-ver"})
	if err != nil {
		t.Fatalf("StripLeadingBackendFlags: %v", err)
	}
	if kind != BackendNative {
		t.Fatalf("kind: got %v, want BackendNative", kind)
	}
	if !reflect.DeepEqual(rest, []string{"-ver"}) {
		t.Fatalf("rest: got %#v", rest)
	}
}

func TestStripLeadingBackendFlagsExplicitWasmTwoToken(t *testing.T) {
	t.Helper()
	rest, kind, err := StripLeadingBackendFlags([]string{"--backend", "wasm", "-q", "-ver"})
	if err != nil {
		t.Fatalf("StripLeadingBackendFlags: %v", err)
	}
	if kind != BackendWasm {
		t.Fatalf("kind: got %v, want BackendWasm", kind)
	}
	if !reflect.DeepEqual(rest, []string{"-q", "-ver"}) {
		t.Fatalf("rest: got %#v", rest)
	}
}

func TestStripLeadingBackendFlagsMissingValue(t *testing.T) {
	t.Helper()
	_, _, err := StripLeadingBackendFlags([]string{"--backend"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--backend requires a value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStripLeadingBackendFlagsStripsOnlyLeadingReserved(t *testing.T) {
	t.Helper()

	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "single_eq_then_rest",
			in:   []string{"--backend=native", "-json", "a.jpg", "b.jpg"},
			want: []string{"-json", "a.jpg", "b.jpg"},
		},
		{
			name: "two_tokens_then_rest",
			in:   []string{"--backend", "wasm", "pipeline", "-l", "INFO"},
			want: []string{"pipeline", "-l", "INFO"},
		},
		{
			name: "double_leading_backend",
			in:   []string{"--backend=native", "--backend=wasm", "-w", "3"},
			want: []string{"-w", "3"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got, _, err := StripLeadingBackendFlags(tc.in)
			if err != nil {
				t.Fatalf("StripLeadingBackendFlags: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}
