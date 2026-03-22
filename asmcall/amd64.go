//go:build amd64

package asmcall

import (
	_ "runtime"
	"unsafe"
)

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
