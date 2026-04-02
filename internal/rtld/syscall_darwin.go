//go:build darwin && arm64 && !cgo

package rtld

import (
	"sync"
	"unsafe"
)

var syscall15Pool = sync.Pool{New: func() any { return new(syscall15Args) }}

var syscall15XABI0 uintptr

func syscall_syscall15X(fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 uintptr) (r1, r2, err uintptr) {
	args := syscall15Pool.Get().(*syscall15Args)
	defer syscall15Pool.Put(args)

	*args = syscall15Args{}
	args.set(fn, []uintptr{a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15})
	runtime_cgocall(syscall15XABI0, unsafe.Pointer(args))
	return args.a1, args.a2, args.a3
}
