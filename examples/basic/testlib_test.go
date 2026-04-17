//go:build ((windows || linux) && (amd64 || arm64)) || (darwin && arm64)

package main

import (
	"fmt"
	"os"
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

const basicRuntimePathEnv = "GO2ZIG_BASIC_RUNTIME_PATH"

func prepareExampleRuntime(tb testing.TB) {
	tb.Helper()
	prepareRuntimeOnce.Do(func() {
		if override := os.Getenv(basicRuntimePathEnv); override != "" {
			path, err := filepath.Abs(override)
			if err != nil {
				prepareRuntimeErr = fmt.Errorf("resolve %s: %w", basicRuntimePathEnv, err)
				return
			}
			if _, err := os.Stat(path); err != nil {
				prepareRuntimeErr = fmt.Errorf("stat %s=%q: %w", basicRuntimePathEnv, path, err)
				return
			}
			prepareRuntimePath = path
			return
		}

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
	})
	if prepareRuntimeSkip != "" {
		tb.Skip(prepareRuntimeSkip)
	}
	if prepareRuntimeErr != nil {
		tb.Fatal(prepareRuntimeErr)
	}
	Default = NewGo2ZigClient(prepareRuntimePath)
}
