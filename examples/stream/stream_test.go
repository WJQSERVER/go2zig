//go:build ((windows || linux) && (amd64 || arm64)) || (darwin && arm64)

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFileHandleEncodingUsesUntaggedValues(t *testing.T) {
	t.Parallel()

	file, err := os.OpenFile(filepath.Join(t.TempDir(), "handle.bin"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	reader, err := NewGoReader(file)
	if err != nil {
		t.Fatalf("NewGoReader() error = %v", err)
	}
	writer, err := NewGoWriter(file)
	if err != nil {
		t.Fatalf("NewGoWriter() error = %v", err)
	}

	if got := reader.handle(); got == 0 {
		t.Fatal("reader.handle() = 0")
	} else if got&_go2zigStreamHandleMask != 0 {
		t.Fatalf("reader.handle() tag = %d, want 0", got&_go2zigStreamHandleMask)
	}
	if got := writer.handle(); got == 0 {
		t.Fatal("writer.handle() = 0")
	} else if got&_go2zigStreamHandleMask != 0 {
		t.Fatalf("writer.handle() tag = %d, want 0", got&_go2zigStreamHandleMask)
	}
	if reader.handle() != writer.handle() {
		t.Fatalf("handle mismatch: %d vs %d", reader.handle(), writer.handle())
	}
}

func TestCopyStreamFileHandles(t *testing.T) {
	t.Parallel()
	ensureStreamLoaded(t)

	dir := t.TempDir()
	srcPath := filepath.Join(dir, "in.bin")
	dstPath := filepath.Join(dir, "out.bin")
	payload := bytes.Repeat([]byte("go2zig-stream-copy-"), 1024)
	if err := os.WriteFile(srcPath, payload, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", srcPath, err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		t.Fatalf("Open(%s) error = %v", srcPath, err)
	}
	defer func() { _ = src.Close() }()
	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("OpenFile(%s) error = %v", dstPath, err)
	}
	defer func() { _ = dst.Close() }()

	reader, err := NewGoReader(src)
	if err != nil {
		t.Fatalf("NewGoReader() error = %v", err)
	}
	writer, err := NewGoWriter(dst)
	if err != nil {
		t.Fatalf("NewGoWriter() error = %v", err)
	}

	if n := CopyStream(reader, writer); n != uint64(len(payload)) {
		t.Fatalf("CopyStream() = %d, want %d", n, len(payload))
	}
	if err := reader.Err(); err != nil {
		t.Fatalf("reader.Err() error = %v", err)
	}
	if err := writer.Err(); err != nil {
		t.Fatalf("writer.Err() error = %v", err)
	}
	if _, err := src.Seek(0, 0); err != nil {
		t.Fatalf("src.Seek() error = %v", err)
	}
	if _, err := dst.Seek(0, 0); err != nil {
		t.Fatalf("dst.Seek() error = %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", dstPath, err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("copied payload mismatch: got %d bytes, want %d", len(got), len(payload))
	}
}
