//go:build ((windows || linux) && (amd64 || arm64)) || (darwin && arm64)

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type partialWriteCloser struct {
	buf       bytes.Buffer
	maxChunk  int
	closeErr  error
	closeCall int
}

func (w *partialWriteCloser) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p)
	if w.maxChunk > 0 && n > w.maxChunk {
		n = w.maxChunk
	}
	return w.buf.Write(p[:n])
}

func (w *partialWriteCloser) Close() error {
	w.closeCall++
	return w.closeErr
}

type closeErrWriteCloser struct {
	buf       bytes.Buffer
	closeErr  error
	closeCall int
}

func (w *closeErrWriteCloser) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *closeErrWriteCloser) Close() error {
	w.closeCall++
	return w.closeErr
}

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

func TestEnsureStreamLoadedKeepsLoadedRuntime(t *testing.T) {
	ensureStreamLoaded(t)

	if Default == nil || Default.rt == nil {
		t.Fatal("Default runtime is nil after ensureStreamLoaded")
	}
	if Default.rt.procCopyStream == 0 {
		t.Fatal("procCopyStream = 0 after ensureStreamLoaded")
	}

	beforeClient := Default
	beforeRuntime := Default.rt

	ensureStreamLoaded(t)

	if Default != beforeClient {
		t.Fatal("ensureStreamLoaded replaced the default client")
	}
	if Default.rt != beforeRuntime {
		t.Fatal("ensureStreamLoaded replaced the loaded runtime")
	}
	if Default.rt.procCopyStream == 0 {
		t.Fatal("procCopyStream = 0 after repeated ensureStreamLoaded")
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

func TestCopyStreamHandlesPartialWrites(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	content, err := os.ReadFile(filepath.Join(filepath.Dir(file), "lib.zig"))
	if err != nil {
		t.Fatalf("ReadFile(lib.zig) error = %v", err)
	}
	text := string(content)
	for _, check := range []string{
		"var pending = buf[0..n];",
		"while (pending.len > 0)",
		"if (written == 0) @panic(\"stream write returned zero bytes\");",
		"pending = pending[written..];",
	} {
		if !strings.Contains(text, check) {
			t.Fatalf("lib.zig missing partial-write handling %q\n%s", check, text)
		}
	}
}

func TestCopyStreamDoesNotPanicOnWriterCloseError(t *testing.T) {
	t.Parallel()
	ensureStreamLoaded(t)

	reader, err := NewGoReader(strings.NewReader("close-error"))
	if err != nil {
		t.Fatalf("NewGoReader() error = %v", err)
	}
	closeErr := fmt.Errorf("close failed")
	writerTarget := &closeErrWriteCloser{closeErr: closeErr}
	writer, err := NewGoWriteCloser(writerTarget)
	if err != nil {
		t.Fatalf("NewGoWriteCloser() error = %v", err)
	}

	didPanic := false
	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()
		if n := CopyStream(reader, writer); n != uint64(len("close-error")) {
			t.Fatalf("CopyStream() = %d, want %d", n, len("close-error"))
		}
	}()
	if didPanic {
		t.Fatal("CopyStream() panicked on writer close error")
	}
	if err := reader.Err(); err != nil {
		t.Fatalf("reader.Err() error = %v", err)
	}
	if err := writer.Err(); err == nil {
		t.Fatal("writer.Err() = nil, want close error")
	} else if !strings.Contains(err.Error(), closeErr.Error()) {
		t.Fatalf("writer.Err() = %v, want close error %q", err, closeErr)
	}
	if got, want := writerTarget.buf.String(), "close-error"; got != want {
		t.Fatalf("partial writer payload = %q, want %q", got, want)
	}
	if writerTarget.closeCall != 1 {
		t.Fatalf("writer close calls = %d, want 1", writerTarget.closeCall)
	}
}
