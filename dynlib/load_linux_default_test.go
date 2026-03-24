//go:build linux

package dynlib_test

import (
	"testing"

	"go2zig/dynlib"
)

func TestLinuxRTLDConstants(t *testing.T) {
	t.Parallel()

	if dynlib.RTLDDefault != 0x00000 {
		t.Fatalf("RTLDDefault = %#x, want %#x", dynlib.RTLDDefault, 0x00000)
	}
	if dynlib.RTLDLazy != 0x00001 {
		t.Fatalf("RTLDLazy = %#x, want %#x", dynlib.RTLDLazy, 0x00001)
	}
	if dynlib.RTLDNow != 0x00002 {
		t.Fatalf("RTLDNow = %#x, want %#x", dynlib.RTLDNow, 0x00002)
	}
	if dynlib.RTLDGlobal != 0x00100 {
		t.Fatalf("RTLDGlobal = %#x, want %#x", dynlib.RTLDGlobal, 0x00100)
	}
}
