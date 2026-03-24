//go:build linux && amd64 && !cgo

#include "textflag.h"

TEXT dlopen(SB), NOSPLIT|NOFRAME, $0-0
	JMP go2zig_rtld_dlopen(SB)

TEXT dlsym(SB), NOSPLIT|NOFRAME, $0-0
	JMP go2zig_rtld_dlsym(SB)

TEXT dlerror(SB), NOSPLIT|NOFRAME, $0-0
	JMP go2zig_rtld_dlerror(SB)

TEXT dlclose(SB), NOSPLIT|NOFRAME, $0-0
	JMP go2zig_rtld_dlclose(SB)
