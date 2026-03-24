//go:build !linux && !(darwin && arm64)

package dynlib

const (
	RTLDDefault = uintptr(0)
	RTLDLazy    = uintptr(0)
	RTLDNow     = uintptr(0)
	RTLDLocal   = uintptr(0)
	RTLDGlobal  = uintptr(0)
)
