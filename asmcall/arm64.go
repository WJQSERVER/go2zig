//go:build arm64

package asmcall

import "unsafe"

var _ = CallFuncG0P2StoreR1

//go:noescape
//go:nosplit
func CallFuncG0P0(fn unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncG0P0R1(fn unsafe.Pointer) uintptr

//go:noescape
//go:nosplit
func CallFuncG0P1(fn, arg0 unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncG0P1R1(fn, arg0 unsafe.Pointer) uintptr

//go:noescape
//go:nosplit
func CallFuncG0P2(fn, arg0, arg1 unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncG0P2StoreR1(fn, arg0, arg1, out unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncG0P2R1(fn, arg0, arg1 unsafe.Pointer) uintptr

//go:noescape
//go:nosplit
func CallFuncG0P3(fn, arg0, arg1, arg2 unsafe.Pointer)

//go:noescape
//go:nosplit
func CallFuncG0P3R1(fn, arg0, arg1, arg2 unsafe.Pointer) uintptr

func CallFuncP0(fn unsafe.Pointer) {
	CallFuncG0P0(fn)
}

func CallFuncP0R1(fn unsafe.Pointer) uintptr {
	return CallFuncG0P0R1(fn)
}

func CallFuncP1(fn, arg0 unsafe.Pointer) {
	CallFuncG0P1(fn, arg0)
}

func CallFuncP1R1(fn, arg0 unsafe.Pointer) uintptr {
	return CallFuncG0P1R1(fn, arg0)
}

func CallFuncP2(fn, arg0, arg1 unsafe.Pointer) {
	CallFuncG0P2(fn, arg0, arg1)
}

func CallFuncP2R1(fn, arg0, arg1 unsafe.Pointer) uintptr {
	return CallFuncG0P2R1(fn, arg0, arg1)
}

func CallFuncP3(fn, arg0, arg1, arg2 unsafe.Pointer) {
	CallFuncG0P3(fn, arg0, arg1, arg2)
}

func CallFuncP3R1(fn, arg0, arg1, arg2 unsafe.Pointer) uintptr {
	return CallFuncG0P3R1(fn, arg0, arg1, arg2)
}
