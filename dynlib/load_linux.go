//go:build linux

package dynlib

import (
	"fmt"
	"go2zig/asmcall"
	"unsafe"
	_ "unsafe"
)

//go:cgo_import_dynamic libdl_dlopen dlopen "libdl.so.2"
//go:cgo_import_dynamic libdl_dlsym dlsym "libdl.so.2"
//go:cgo_import_dynamic libdl_dlclose dlclose "libdl.so.2"
//go:cgo_import_dynamic _ _ "libdl.so.2"

var libdl_dlopen uintptr
var libdl_dlsym uintptr
var libdl_dlclose uintptr

const (
	rtldLazy   = 0x00001
	rtldLocal  = 0
	rtldNoLoad = 0x00004
)

type Library struct {
	handle uintptr
	paths  [][]byte
}

const (
	RTLDDefault = 0x00000
	RTLDLazy    = 0x00001
	RTLDNow     = 0x00002
	RTLDLocal   = 0x00000
	RTLDGlobal  = 0x00100
)

func Load(path string) (*Library, error) {
	pathBytes := append([]byte(path), 0)
	handle := libdlOpen(pathBytes, RTLDLazy|RTLDLocal)
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
	addr := libdlSym(l.handle, nameBytes)
	if addr == 0 {
		return 0, fmt.Errorf("dlsym %s failed", name)
	}
	return addr, nil
}

func (l *Library) Close() error {
	if l == nil || l.handle == 0 {
		return nil
	}
	ret := libdlClose(l.handle)
	l.handle = 0
	l.paths = nil
	if ret != 0 {
		return fmt.Errorf("dlclose failed")
	}
	return nil
}

func libdlOpen(path []byte, mode uintptr) uintptr {
	var out uintptr
	asmcall.CallFuncG0P2StoreR1(unsafe.Pointer(libdl_dlopen), unsafe.Pointer(unsafe.SliceData(path)), unsafe.Pointer(mode), unsafe.Pointer(&out))
	return out
}

func libdlSym(handle uintptr, name []byte) uintptr {
	var out uintptr
	asmcall.CallFuncG0P2StoreR1(unsafe.Pointer(libdl_dlsym), unsafe.Pointer(handle), unsafe.Pointer(unsafe.SliceData(name)), unsafe.Pointer(&out))
	return out
}

func libdlClose(handle uintptr) uintptr {
	var out uintptr
	asmcall.CallFuncG0P2StoreR1(unsafe.Pointer(libdl_dlclose), unsafe.Pointer(handle), nil, unsafe.Pointer(&out))
	return out
}
