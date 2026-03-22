//go:build linux

package dynlib_test

import (
	"os"
	"testing"

	"go2zig/dynlib"
)

func TestLoadAndLookupLinux(t *testing.T) {
	if os.Getenv("GO2ZIG_RUN_LINUX_RUNTIME_TESTS") != "1" {
		t.Skip("linux runtime execution tests are disabled by default")
	}
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
