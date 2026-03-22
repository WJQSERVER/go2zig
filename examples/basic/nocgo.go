//go:build !cgo

package main

import "log"

func main() {
	log.Fatal("example requires cgo; run with CGO_ENABLED=1 and a working C compiler such as `CC=zig cc`")
}
