//go:build !((amd64 || arm64) && (windows || linux))

package main

import "log"

func main() {
	log.Fatal("example currently requires windows/linux on amd64 or arm64 for the no-cgo asm runtime")
}
