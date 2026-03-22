//go:build !(windows && amd64)

package main

import "log"

func main() {
	log.Fatal("example currently requires windows/amd64 for the no-cgo asm runtime")
}
