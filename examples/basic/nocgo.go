//go:build !(((windows || linux) && (amd64 || arm64)) || (darwin && arm64))

package main

import "log"

func main() {
	log.Fatal("example currently requires windows/linux on amd64 or arm64, or darwin on arm64, for the no-cgo asm runtime")
}
