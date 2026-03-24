//go:build linux && amd64 && !cgo

package rtld

import "unsafe"

//go:cgo_import_dynamic go2zig_rtld_dlopen dlopen "libdl.so.2"
//go:cgo_import_dynamic go2zig_rtld_dlsym dlsym "libdl.so.2"
//go:cgo_import_dynamic go2zig_rtld_dlerror dlerror "libdl.so.2"
//go:cgo_import_dynamic go2zig_rtld_dlclose dlclose "libdl.so.2"
//go:cgo_import_dynamic _ _ "libdl.so.2"

//go:linkname dlopen dlopen
var dlopen byte

var dlopenABI0 = uintptr(unsafe.Pointer(&dlopen))

//go:linkname dlsym dlsym
var dlsym byte

var dlsymABI0 = uintptr(unsafe.Pointer(&dlsym))

//go:linkname dlerror dlerror
var dlerror byte

var dlerrorABI0 = uintptr(unsafe.Pointer(&dlerror))

//go:linkname dlclose dlclose
var dlclose byte

var dlcloseABI0 = uintptr(unsafe.Pointer(&dlclose))
