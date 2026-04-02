//go:build windows

package dynlib

import "syscall"

type Library struct {
	dll *syscall.DLL
}

func Load(path string) (*Library, error) {
	dll, err := syscall.LoadDLL(path)
	if err != nil {
		return nil, err
	}
	return &Library{dll: dll}, nil
}

func (l *Library) Lookup(name string) (uintptr, error) {
	proc, err := l.dll.FindProc(name)
	if err != nil {
		return 0, err
	}
	return proc.Addr(), nil
}

func (l *Library) Close() error {
	if l == nil || l.dll == nil {
		return nil
	}
	return l.dll.Release()
}
