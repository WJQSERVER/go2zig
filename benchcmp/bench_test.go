//go:build cgo && (amd64 || arm64)

package benchcmp

import (
	"testing"
	"unsafe"

	"go2zig/asmcall"
)

func BenchmarkCgoAddU64(b *testing.B) {
	var out uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = CAddU64(uint64(i), uint64(i+1))
	}
	_ = out
}

func BenchmarkAsmCallCAddU64(b *testing.B) {
	addr := CAddU64Addr()
	var out uintptr
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		asmcall.CallFuncG0P2StoreR1(
			unsafe.Pointer(addr),
			unsafe.Pointer(uintptr(i)),
			unsafe.Pointer(uintptr(i+1)),
			unsafe.Pointer(&out),
		)
	}
	_ = out
}
