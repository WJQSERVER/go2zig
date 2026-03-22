//go:build cgo && amd64

package asmcall_test

/*
#include <stdint.h>
#include <stddef.h>
static uint64_t bench_add_u64(uint64_t a, uint64_t b) { return a + b; }
static uintptr_t bench_add_u64_addr(void) { return (uintptr_t)(bench_add_u64); }
*/
import "C"

import (
	"testing"
	"unsafe"

	"go2zig/asmcall"
)

func BenchmarkCgoAddU64(b *testing.B) {
	var out uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = uint64(C.bench_add_u64(C.uint64_t(i), C.uint64_t(i+1)))
	}
	_ = out
}

func BenchmarkAsmCallCAddU64(b *testing.B) {
	addr := uintptr(C.bench_add_u64_addr())
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
