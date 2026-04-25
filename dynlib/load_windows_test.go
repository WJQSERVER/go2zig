//go:build windows

package dynlib_test

import (
	"testing"

	"github.com/WJQSERVER/go2zig/dynlib"
)

func TestLoadAndLookup(t *testing.T) {
	t.Parallel()

	lib, err := dynlib.Load("kernel32.dll")
	if err != nil {
		t.Fatalf("Load(kernel32.dll) error = %v", err)
	}
	defer lib.Close()

	addr, err := lib.Lookup("SetLastError")
	if err != nil {
		t.Fatalf("Lookup(SetLastError) error = %v", err)
	}
	if addr == 0 {
		t.Fatal("Lookup(SetLastError) returned 0")
	}
}
