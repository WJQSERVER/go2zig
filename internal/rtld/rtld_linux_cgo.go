//go:build linux && amd64 && cgo

package rtld

/*
#cgo LDFLAGS: -ldl

#include <dlfcn.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

func Available() bool {
	return true
}

func Dlopen(path string, mode int) (uintptr, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	handle := C.dlopen(cpath, C.int(mode))
	if handle == nil {
		return 0, errors.New(C.GoString(C.dlerror()))
	}
	return uintptr(handle), nil
}

func Dlsym(handle uintptr, name string) (uintptr, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	addr := C.dlsym(*(*unsafe.Pointer)(unsafe.Pointer(&handle)), cname)
	if addr == nil {
		return 0, errors.New(C.GoString(C.dlerror()))
	}
	return uintptr(addr), nil
}

func Dlclose(handle uintptr) error {
	ret := C.dlclose(*(*unsafe.Pointer)(unsafe.Pointer(&handle)))
	if ret != 0 {
		return errors.New(C.GoString(C.dlerror()))
	}
	return nil
}
