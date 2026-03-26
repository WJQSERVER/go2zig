//go:build ((windows || linux) && (amd64 || arm64)) || (darwin && arm64)

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var (
	prepareRuntimeOnce sync.Once
	prepareRuntimePath string
	prepareRuntimeSkip string
	prepareRuntimeErr  error
)

func prepareStreamRuntime(tb testing.TB) {
	tb.Helper()
	prepareRuntimeOnce.Do(func() {
		zigPath, err := exec.LookPath("zig")
		if err != nil {
			prepareRuntimeSkip = "zig not available in PATH"
			return
		}

		_, file, _, ok := runtime.Caller(0)
		if !ok {
			prepareRuntimeErr = fmt.Errorf("runtime.Caller failed")
			return
		}
		dir := filepath.Dir(file)
		libPath := filepath.Join(dir, _go2zigDynamicLibraryName())
		root := filepath.Join(dir, "go2zig_build_root.zig")

		cmd := exec.Command(zigPath, "build-lib", "-dynamic", "-O", "ReleaseSafe", "-femit-bin="+libPath, root)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			prepareRuntimeErr = fmt.Errorf("zig build-lib failed: %v\n%s", err, out)
			return
		}

		prepareRuntimePath = libPath
		Default = NewGo2ZigClient(prepareRuntimePath)
	})
	if prepareRuntimeSkip != "" {
		tb.Skip(prepareRuntimeSkip)
	}
	if prepareRuntimeErr != nil {
		tb.Fatal(prepareRuntimeErr)
	}
}

func ensureStreamLoaded(tb testing.TB) {
	tb.Helper()
	prepareStreamRuntime(tb)
	if err := Default.Load(); err != nil {
		tb.Fatalf("Load() error = %v", err)
	}
}
