//go:build !windows && !linux && !(darwin && arm64)

package dynlib

import "fmt"

type Library struct{}

func Load(path string) (*Library, error) {
	return nil, fmt.Errorf("dynamic loading is not implemented on this platform: %s", path)
}

func (l *Library) Lookup(name string) (uintptr, error) {
	return 0, fmt.Errorf("dynamic loading is not implemented on this platform: %s", name)
}

func (l *Library) Close() error {
	return nil
}
