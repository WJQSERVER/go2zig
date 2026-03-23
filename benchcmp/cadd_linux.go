//go:build cgo && linux && (amd64 || arm64)

package benchcmp

/*
#include <stdint.h>
#include <stddef.h>
static uint64_t bench_add_u64(uint64_t a, uint64_t b) { return a + b; }
static uintptr_t bench_add_u64_addr(void) { return (uintptr_t)(bench_add_u64); }
*/
import "C"

func CAddU64(a, b uint64) uint64 {
	return uint64(C.bench_add_u64(C.uint64_t(a), C.uint64_t(b)))
}

func CAddU64Addr() uintptr {
	return uintptr(C.bench_add_u64_addr())
}
