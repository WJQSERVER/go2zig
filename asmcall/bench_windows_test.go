//go:build windows

package asmcall_test

import (
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"github.com/WJQSERVER/go2zig/asmcall"
)

func BenchmarkSyscallStoreSum(b *testing.B) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	addr := buildAndLookupStoreSum(b)
	var slot uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = syscall.SyscallN(addr, uintptr(unsafe.Pointer(&slot)), uintptr(i), uintptr(i+1))
	}
}

func BenchmarkAsmStoreSum(b *testing.B) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	addr := buildAndLookupStoreSum(b)
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
