//go:build ((windows || linux) && (amd64 || arm64)) || (darwin && arm64)

package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

func main() {
	if err := Default.Load(); err != nil {
		log.Fatal(err)
	}

	const payload = "go2zig stream example\n"
	reader, err := NewGoReader(strings.NewReader(payload))
	if err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	writer, err := NewGoWriter(&out)
	if err != nil {
		log.Fatal(err)
	}
	n, err := CopyStream(reader, writer)
	if err != nil {
		log.Fatal(err)
	}
	if n != uint64(len(payload)) {
		log.Fatalf("CopyStream() = %d, want %d", n, len(payload))
	}
	if err := reader.Err(); err != nil {
		log.Fatal(err)
	}
	if err := writer.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Print(out.String())
}
