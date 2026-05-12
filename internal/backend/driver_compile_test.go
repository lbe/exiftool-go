package backend

import "testing"

func TestDriverInterfaceImplementations(t *testing.T) {
	t.Helper()
	var (
		_ Driver = (*NativeDriver)(nil)
		_ Driver = (*WasmDriver)(nil)
		_ Driver = (*DispatchDriver)(nil)
	)
}
