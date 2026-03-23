//go:build darwin

package dynlib

import (
	"fmt"
	"go2zig/asmcall"
	"unsafe"
	_ "unsafe"
)

//go:cgo_import_dynamic libsystem_dlopen dlopen "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libsystem_dlsym dlsym "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libsystem_dlclose dlclose "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic _ _ "/usr/lib/libSystem.B.dylib"

var libsystem_dlopen uintptr
var libsystem_dlsym uintptr
var libsystem_dlclose uintptr

const (
	rtldLazy   = 0x1
	rtldLocal  = 0x4
	rtldNoLoad = 0x10
)

type Library struct {
	handle uintptr
	paths  [][]byte
}

func Load(path string) (*Library, error) {
	pathBytes := append([]byte(path), 0)
	handle := libSystemOpen(pathBytes, rtldLazy|rtldLocal)
	if handle == 0 {
		return nil, fmt.Errorf("dlopen %s failed", path)
	}
	return &Library{handle: handle, paths: [][]byte{pathBytes}}, nil
}

func (l *Library) Lookup(name string) (uintptr, error) {
	if l == nil || l.handle == 0 {
		return 0, fmt.Errorf("library is not loaded")
	}
	nameBytes := append([]byte(name), 0)
	l.paths = append(l.paths, nameBytes)
	addr := libSystemSym(l.handle, nameBytes)
	if addr == 0 {
		return 0, fmt.Errorf("dlsym %s failed", name)
	}
	return addr, nil
}

func (l *Library) Close() error {
	if l == nil || l.handle == 0 {
		return nil
	}
	ret := libSystemClose(l.handle)
	l.handle = 0
	l.paths = nil
	if ret != 0 {
		return fmt.Errorf("dlclose failed")
	}
	return nil
}

func libSystemOpen(path []byte, mode uintptr) uintptr {
	var out uintptr
	asmcall.CallFuncG0P2StoreR1(unsafe.Pointer(libsystem_dlopen), unsafe.Pointer(unsafe.SliceData(path)), unsafe.Pointer(mode), unsafe.Pointer(&out))
	return out
}

func libSystemSym(handle uintptr, name []byte) uintptr {
	var out uintptr
	asmcall.CallFuncG0P2StoreR1(unsafe.Pointer(libsystem_dlsym), unsafe.Pointer(handle), unsafe.Pointer(unsafe.SliceData(name)), unsafe.Pointer(&out))
	return out
}

func libSystemClose(handle uintptr) uintptr {
	return asmcall.CallFuncG0P1R1(unsafe.Pointer(libsystem_dlclose), unsafe.Pointer(handle))
}
