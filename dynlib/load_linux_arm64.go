//go:build linux && arm64

package dynlib

import (
	"fmt"

	"go2zig/internal/rtld"
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
	if !rtld.Available() {
		return nil, fmt.Errorf("linux dynamic loader symbols are unavailable")
	}
	pathBytes := append([]byte(path), 0)
	handle, err := rtld.Dlopen(path, int(RTLDLazy|RTLDLocal))
	if err != nil {
		return nil, fmt.Errorf("dlopen %s failed: %w", path, err)
	}
	return &Library{handle: handle, paths: [][]byte{pathBytes}}, nil
}

func (l *Library) Lookup(name string) (uintptr, error) {
	if l == nil || l.handle == 0 {
		return 0, fmt.Errorf("library is not loaded")
	}
	nameBytes := append([]byte(name), 0)
	l.paths = append(l.paths, nameBytes)
	addr, err := rtld.Dlsym(l.handle, name)
	if err != nil {
		return 0, fmt.Errorf("dlsym %s failed: %w", name, err)
	}
	return addr, nil
}

func (l *Library) Close() error {
	if l == nil || l.handle == 0 {
		return nil
	}
	err := rtld.Dlclose(l.handle)
	l.handle = 0
	l.paths = nil
	if err != nil {
		return fmt.Errorf("dlclose failed: %w", err)
	}
	return nil
}
