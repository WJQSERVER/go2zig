//go:build !(amd64 && (windows || linux))

package main

import "log"

func main() {
	log.Fatal("example currently requires windows/amd64 or linux/amd64 for the no-cgo asm runtime")
}
