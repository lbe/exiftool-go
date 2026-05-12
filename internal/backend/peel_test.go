package backend

import (
	"reflect"
	"testing"
)

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
