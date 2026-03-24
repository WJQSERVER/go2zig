//go:build ((linux && (amd64 || arm64)) || (darwin && arm64)) && !cgo

package rtld

const maxArgs = 15

type syscall15Args struct {
	fn, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 uintptr
	f1, f2, f3, f4, f5, f6, f7, f8                                       uintptr
	arm64_r8                                                             uintptr
}

func (s *syscall15Args) set(fn uintptr, ints []uintptr) {
	s.fn = fn
	s.a1 = ints[0]
	s.a2 = ints[1]
	s.a3 = ints[2]
	s.a4 = ints[3]
	s.a5 = ints[4]
	s.a6 = ints[5]
	s.a7 = ints[6]
	s.a8 = ints[7]
	s.a9 = ints[8]
	s.a10 = ints[9]
	s.a11 = ints[10]
	s.a12 = ints[11]
	s.a13 = ints[12]
	s.a14 = ints[13]
	s.a15 = ints[14]
	s.f1 = ints[0]
	s.f2 = ints[1]
	s.f3 = ints[2]
	s.f4 = ints[3]
	s.f5 = ints[4]
	s.f6 = ints[5]
	s.f7 = ints[6]
	s.f8 = ints[7]
}

//go:uintptrescapes
func SyscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	if fn == 0 {
		panic("rtld: fn is nil")
	}
	if len(args) > maxArgs {
		panic("rtld: too many arguments to SyscallN")
	}
	var tmp [maxArgs]uintptr
	copy(tmp[:], args)
	return syscall_syscall15X(fn, tmp[0], tmp[1], tmp[2], tmp[3], tmp[4], tmp[5], tmp[6], tmp[7], tmp[8], tmp[9], tmp[10], tmp[11], tmp[12], tmp[13], tmp[14])
}
