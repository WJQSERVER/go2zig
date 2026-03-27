//go:build ((windows || linux) && (amd64 || arm64)) || (darwin && arm64)

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"
)

var streamBenchPayload = bytes.Repeat([]byte("go2zig-stream-payload-"), 2048)

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

func benchmarkReadDirect(state *_go2zigDirectReaderState, scratch []byte) int {
	if state == nil || len(scratch) == 0 {
		return 0
	}
	total := 0
	for state.pos < state.len {
		remaining := int(state.len - state.pos)
		n := len(scratch)
		if n > remaining {
			n = remaining
		}
		src := unsafe.Slice((*byte)(unsafe.Add(state.ptr, state.pos)), n)
		copy(scratch[:n], src)
		state.pos += uintptr(n)
		total += n
	}
	return total
}

func benchmarkWriteDirect(state *_go2zigDirectWriterState, payload []byte) int {
	if state == nil || len(payload) == 0 {
		return 0
	}
	dst := unsafe.Slice((*byte)(state.ptr), int(state.cap))
	start := int(state.len)
	n := copy(dst[start:], payload)
	state.len += uintptr(n)
	return n
}

func BenchmarkStreamReaderDirectBytes(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(streamBenchPayload)))
	scratch := make([]byte, 64<<10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, err := NewGoReader(bytes.NewReader(streamBenchPayload))
		if err != nil {
			b.Fatalf("NewGoReader() error = %v", err)
		}
		if reader.state.directReader == nil {
			b.Fatal("reader direct fast path unavailable")
		}
		if n := benchmarkReadDirect(reader.state.directReader, scratch); n != len(streamBenchPayload) {
			b.Fatalf("read bytes = %d, want %d", n, len(streamBenchPayload))
		}
	}
}

func BenchmarkStreamReaderPipeFallback(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(streamBenchPayload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pr, pw := io.Pipe()
		go func() {
			_, _ = pw.Write(streamBenchPayload)
			_ = pw.Close()
		}()
		reader, err := NewGoReadCloser(pr)
		if err != nil {
			b.Fatalf("NewGoReadCloser() error = %v", err)
		}
		if _, err := io.Copy(io.Discard, reader.state.file); err != nil {
			b.Fatalf("io.Copy() error = %v", err)
		}
		if err := reader.Close(); err != nil {
			b.Fatalf("reader.Close() error = %v", err)
		}
		if err := reader.Err(); err != nil {
			b.Fatalf("reader.Err() error = %v", err)
		}
	}
}

func BenchmarkStreamWriterDirectBuffer(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(streamBenchPayload)))
	var out bytes.Buffer
	out.Grow(len(streamBenchPayload) + 64<<10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out.Reset()
		writer, err := NewGoWriter(&out)
		if err != nil {
			b.Fatalf("NewGoWriter() error = %v", err)
		}
		if writer.state.directWriter == nil {
			b.Fatal("writer direct fast path unavailable")
		}
		if n := benchmarkWriteDirect(writer.state.directWriter, streamBenchPayload); n != len(streamBenchPayload) {
			b.Fatalf("written bytes = %d, want %d", n, len(streamBenchPayload))
		}
		if err := writer.Close(); err != nil {
			b.Fatalf("writer.Close() error = %v", err)
		}
		if out.Len() != len(streamBenchPayload) {
			b.Fatalf("out.Len() = %d, want %d", out.Len(), len(streamBenchPayload))
		}
	}
}

func BenchmarkStreamWriterPipeWriteCloser(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(streamBenchPayload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out bytes.Buffer
		writer, err := NewGoWriteCloser(nopWriteCloser{Writer: &out})
		if err != nil {
			b.Fatalf("NewGoWriteCloser() error = %v", err)
		}
		if _, err := writer.state.file.Write(streamBenchPayload); err != nil {
			b.Fatalf("writer.state.file.Write() error = %v", err)
		}
		if err := writer.Close(); err != nil {
			b.Fatalf("writer.Close() error = %v", err)
		}
		if out.Len() != len(streamBenchPayload) {
			b.Fatalf("out.Len() = %d, want %d", out.Len(), len(streamBenchPayload))
		}
	}
}

func BenchmarkStreamCopyFileHandles(b *testing.B) {
	b.ReportAllocs()
	if runtime.GOOS == "windows" {
		b.Skip("windows file-handle end-to-end stream benchmark is currently unstable")
	}
	ensureStreamLoaded(b)
	dir := b.TempDir()
	srcPath := filepath.Join(dir, "in.bin")
	dstPath := filepath.Join(dir, "out.bin")
	if err := os.WriteFile(srcPath, streamBenchPayload, 0o644); err != nil {
		b.Fatalf("WriteFile(%s) error = %v", srcPath, err)
	}
	src, err := os.Open(srcPath)
	if err != nil {
		b.Fatalf("Open(%s) error = %v", srcPath, err)
	}
	b.Cleanup(func() { _ = src.Close() })
	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		b.Fatalf("OpenFile(%s) error = %v", dstPath, err)
	}
	b.Cleanup(func() { _ = dst.Close() })

	b.SetBytes(int64(len(streamBenchPayload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := src.Seek(0, 0); err != nil {
			b.Fatalf("src.Seek() error = %v", err)
		}
		if err := dst.Truncate(0); err != nil {
			b.Fatalf("dst.Truncate() error = %v", err)
		}
		if _, err := dst.Seek(0, 0); err != nil {
			b.Fatalf("dst.Seek() error = %v", err)
		}
		reader, err := NewGoReader(src)
		if err != nil {
			b.Fatalf("NewGoReader() error = %v", err)
		}
		writer, err := NewGoWriter(dst)
		if err != nil {
			b.Fatalf("NewGoWriter() error = %v", err)
		}
		n, err := CopyStream(reader, writer)
		if err != nil {
			b.Fatalf("CopyStream() error = %v", err)
		}
		if n != uint64(len(streamBenchPayload)) {
			b.Fatalf("CopyStream() = %d, want %d", n, len(streamBenchPayload))
		}
		if got, err := dst.Seek(0, io.SeekCurrent); err != nil {
			b.Fatalf("dst.SeekCurrent() error = %v", err)
		} else if got != int64(len(streamBenchPayload)) {
			b.Fatalf("dst size = %d, want %d", got, len(streamBenchPayload))
		}
	}
}

func BenchmarkStreamHandleAccess(b *testing.B) {
	b.ReportAllocs()
	file, err := os.OpenFile(filepath.Join(b.TempDir(), "handle.bin"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		b.Fatalf("OpenFile() error = %v", err)
	}
	b.Cleanup(func() { _ = file.Close() })
	reader, err := NewGoReader(file)
	if err != nil {
		b.Fatalf("NewGoReader() error = %v", err)
	}
	writer, err := NewGoWriter(file)
	if err != nil {
		b.Fatalf("NewGoWriter() error = %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := reader.handle(); got == 0 {
			b.Fatal("reader.handle() = 0")
		}
		if got := writer.handle(); got == 0 {
			b.Fatal("writer.handle() = 0")
		}
	}
	b.StopTimer()
	if reader.handle() != writer.handle() {
		b.Fatalf("handle mismatch: %d vs %d", reader.handle(), writer.handle())
	}
}
