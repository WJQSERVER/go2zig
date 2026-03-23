//go:build !((amd64 || arm64) && (windows || linux || darwin))

package main

import "log"

func main() {
	log.Fatal("example currently requires windows/linux/darwin on amd64 or arm64 for the no-cgo asm runtime")
}
