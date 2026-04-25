//go:build linux

package asmcall_test

import (
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/WJQSERVER/go2zig/asmcall"
)

func BenchmarkDynlibDLSym(b *testing.B) {
	if os.Getenv("GO2ZIG_RUN_LINUX_RUNTIME_TESTS") != "1" {
		b.Skip("linux runtime execution benchmarks are disabled by default")
	}
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
	_ = slot
}

func BenchmarkAsmStoreSumLinux(b *testing.B) {
	if os.Getenv("GO2ZIG_RUN_LINUX_RUNTIME_TESTS") != "1" {
		b.Skip("linux runtime execution benchmarks are disabled by default")
	}
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
	_ = slot
}

func BenchmarkDynlibLookupStoreSumLinux(b *testing.B) {
	if os.Getenv("GO2ZIG_RUN_LINUX_RUNTIME_TESTS") != "1" {
		b.Skip("linux runtime execution benchmarks are disabled by default")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	addr := buildAndLookupStoreSumLinux(b)
	if addr == 0 {
		b.Fatal("buildAndLookupStoreSumLinux returned 0")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr
	}
}
