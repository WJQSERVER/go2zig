//go:build linux

package asmcall_test

import (
	"runtime"
	"testing"
	"unsafe"

	"go2zig/asmcall"
	"go2zig/dynlib"
)

func BenchmarkDynlibDLSym(b *testing.B) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	lib, err := dynlib.Load("libdl.so.2")
	if err != nil {
		b.Fatal(err)
	}
	defer lib.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := lib.Lookup("dlopen")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAsmStoreSumLinux(b *testing.B) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	addr := buildAndLookupStoreSumLinux(b)
	var slot uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		asmcall.CallFuncG0P3(
			unsafe.Pointer(addr),
			unsafe.Pointer(&slot),
			unsafe.Pointer(uintptr(i)),
			unsafe.Pointer(uintptr(i+1)),
		)
	}
}
