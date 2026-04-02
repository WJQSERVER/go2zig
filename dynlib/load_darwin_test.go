//go:build darwin && arm64

package dynlib_test

import (
	"testing"

	"go2zig/dynlib"
)

func TestLoadAndLookupDarwin(t *testing.T) {
	t.Parallel()

	lib, err := dynlib.Load("/usr/lib/libSystem.B.dylib")
	if err != nil {
		t.Fatalf("Load(/usr/lib/libSystem.B.dylib) error = %v", err)
	}
	defer lib.Close()

	addr, err := lib.Lookup("dlopen")
	if err != nil {
		t.Fatalf("Lookup(dlopen) error = %v", err)
	}
	if addr == 0 {
		t.Fatal("Lookup(dlopen) returned 0")
	}
}
