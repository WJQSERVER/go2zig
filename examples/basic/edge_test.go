//go:build amd64 && (windows || linux)

package main

import (
	"strings"
	"testing"
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
