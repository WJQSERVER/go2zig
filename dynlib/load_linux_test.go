//go:build linux

package dynlib_test

import (
	"testing"

	"go2zig/dynlib"
)

func TestLoadAndLookupLinux(t *testing.T) {
	t.Parallel()

	lib, err := dynlib.Load("libdl.so.2")
	if err != nil {
		t.Fatalf("Load(libdl.so.2) error = %v", err)
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
