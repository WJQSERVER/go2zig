//go:build ((linux && amd64) || (darwin && arm64)) && !cgo

package rtld

import "unsafe"

func dlerrorString() string {
	ptr, _, _ := SyscallN(dlerrorABI0)
	if ptr == 0 {
		return "dynamic loader error"
	}
	return cString(ptr)
}

func cString(ptr uintptr) string {
	p := (*byte)(unsafe.Pointer(ptr))
	var n uintptr
	for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(p, n))
}
