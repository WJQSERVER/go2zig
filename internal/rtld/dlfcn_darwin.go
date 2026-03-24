//go:build darwin && arm64 && !cgo

package rtld

import "unsafe"

//go:cgo_import_dynamic go2zig_rtld_dlopen dlopen "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic go2zig_rtld_dlsym dlsym "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic go2zig_rtld_dlerror dlerror "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic go2zig_rtld_dlclose dlclose "/usr/lib/libSystem.B.dylib"

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
