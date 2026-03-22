//go:build windows && amd64

package dynlib

import _ "go2zig/asmcall"
import _ "unsafe"
import "unsafe"

//go:linkname callFuncG0P2R1 go2zig/asmcall.CallFuncG0P2R1
func callFuncG0P2R1(fn, arg0, arg1 unsafe.Pointer) uintptr
