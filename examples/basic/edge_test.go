//go:build (amd64 || arm64) && (windows || linux)

package main

import (
	"strings"
	"testing"
	"unsafe"
)

func TestOwnStringRejectsInvalidSpan(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil || !strings.Contains(r.(string), "invalid String buffer state") {
			t.Fatalf("panic = %v, want invalid String buffer state", r)
		}
	}()
	_ = _go2zigOwnString(&_go2zigRuntime{}, _go2zigString{Ptr: nil, Len: 1})
}

func TestOwnScoreListRejectsInvalidSpan(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil || !strings.Contains(r.(string), "invalid ScoreList buffer state") {
			t.Fatalf("panic = %v, want invalid ScoreList buffer state", r)
		}
	}()
	_ = _go2zigOwnScoreList(&_go2zigRuntime{}, _go2zigScoreList{ptr: nil, len: 1})
}

func TestOwnStringAllowsZeroLengthWithNonNilPointer(t *testing.T) {
	t.Parallel()

	value := _go2zigOwnString(&_go2zigRuntime{}, _go2zigString{Ptr: unsafe.Pointer(new(byte)), Len: 0})
	if value != "" {
		t.Fatalf("_go2zigOwnString() = %q, want empty string", value)
	}
}

func TestOwnBytesAllowsZeroLengthWithNonNilPointer(t *testing.T) {
	t.Parallel()

	value := _go2zigOwnBytes(&_go2zigRuntime{}, _go2zigBytes{Ptr: unsafe.Pointer(new(byte)), Len: 0})
	if len(value) != 0 {
		t.Fatalf("_go2zigOwnBytes() = %v, want empty slice", value)
	}
}

func TestRuntimeFreeSkipsZeroLengthPointer(t *testing.T) {
	t.Parallel()

	rt := &_go2zigRuntime{}
	rt.free(unsafe.Pointer(new(byte)), 0)
}
