//go:build darwin && arm64 && !cgo

package rtld

import (
	"errors"
	"runtime"
	"unsafe"

	_ "go2zig/internal/rtld/fakecgo"
)

func Available() bool {
	return true
}

func Dlopen(path string, mode int) (uintptr, error) {
	pathBytes := append([]byte(path), 0)
	handle, _, _ := SyscallN(dlopenABI0, uintptr(unsafe.Pointer(unsafe.SliceData(pathBytes))), uintptr(mode))
	runtime.KeepAlive(pathBytes)
	if handle == 0 {
		return 0, errors.New(dlerrorString())
	}
	return handle, nil
}

func Dlsym(handle uintptr, name string) (uintptr, error) {
	nameBytes := append([]byte(name), 0)
	addr, _, _ := SyscallN(dlsymABI0, handle, uintptr(unsafe.Pointer(unsafe.SliceData(nameBytes))))
	runtime.KeepAlive(nameBytes)
	if addr == 0 {
		return 0, errors.New(dlerrorString())
	}
	return addr, nil
}

func Dlclose(handle uintptr) error {
	ret, _, _ := SyscallN(dlcloseABI0, handle)
	if ret != 0 {
		return errors.New(dlerrorString())
	}
	return nil
}
