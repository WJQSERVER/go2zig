//go:build amd64 && !windows

package asmcall

import (
	"unsafe"
	_ "unsafe"
)

var _ = CallFuncG0P2StoreR1

//go:linkname runtimeSystemstack runtime.systemstack
func runtimeSystemstack(fn func())

func CallFuncG0P0(fn unsafe.Pointer) {
	runtimeSystemstack(func() {
		CallFuncP0(fn)
	})
}

func CallFuncG0P0R1(fn unsafe.Pointer) uintptr {
	var ret uintptr
	runtimeSystemstack(func() {
		ret = CallFuncP0R1(fn)
	})
	return ret
}

func CallFuncG0P1(fn, arg0 unsafe.Pointer) {
	runtimeSystemstack(func() {
		CallFuncP1(fn, arg0)
	})
}

func CallFuncG0P1R1(fn, arg0 unsafe.Pointer) uintptr {
	var ret uintptr
	runtimeSystemstack(func() {
		ret = CallFuncP1R1(fn, arg0)
	})
	return ret
}

func CallFuncG0P2(fn, arg0, arg1 unsafe.Pointer) {
	runtimeSystemstack(func() {
		CallFuncP2(fn, arg0, arg1)
	})
}

func CallFuncG0P2StoreR1(fn, arg0, arg1, out unsafe.Pointer) {
	var ret uintptr
	runtimeSystemstack(func() {
		ret = CallFuncP2R1(fn, arg0, arg1)
	})
	if out != nil {
		*(*uintptr)(out) = ret
	}
}

func CallFuncG0P2R1(fn, arg0, arg1 unsafe.Pointer) uintptr {
	var ret uintptr
	runtimeSystemstack(func() {
		ret = CallFuncP2R1(fn, arg0, arg1)
	})
	return ret
}

func CallFuncG0P3(fn, arg0, arg1, arg2 unsafe.Pointer) {
	runtimeSystemstack(func() {
		CallFuncP3(fn, arg0, arg1, arg2)
	})
}

func CallFuncG0P3R1(fn, arg0, arg1, arg2 unsafe.Pointer) uintptr {
	var ret uintptr
	runtimeSystemstack(func() {
		ret = CallFuncP3R1(fn, arg0, arg1, arg2)
	})
	return ret
}

//go:noescape
//go:nosplit
func CallFuncP0(fn unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncP0R1(fn unsafe.Pointer) uintptr

//go:noescape
//go:nosplit
func CallFuncP1(fn, arg0 unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncP1R1(fn, arg0 unsafe.Pointer) uintptr

//go:noescape
//go:nosplit
func CallFuncP2(fn, arg0, arg1 unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncP2R1(fn, arg0, arg1 unsafe.Pointer) uintptr

//go:noescape
//go:nosplit
func CallFuncP3(fn, arg0, arg1, arg2 unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncP3R1(fn, arg0, arg1, arg2 unsafe.Pointer) uintptr
