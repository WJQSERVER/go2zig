//go:build ((linux && (amd64 || arm64)) || (darwin && arm64)) && !cgo

package rtld

import "unsafe"

//go:linkname runtime_cgocall runtime.cgocall
func runtime_cgocall(fn uintptr, arg unsafe.Pointer) int32
