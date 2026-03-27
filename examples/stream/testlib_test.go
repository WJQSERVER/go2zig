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

func streamExampleDir(tb testing.TB) string {
	tb.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func readStreamExampleFile(tb testing.TB, name string) []byte {
	tb.Helper()
	content, err := os.ReadFile(filepath.Join(streamExampleDir(tb), name))
	if err != nil {
		tb.Fatalf("ReadFile(%s) error = %v", name, err)
	}
	return content
}

func buildCustomStreamRuntime(tb testing.TB, libSource, runtimeSource string) string {
	tb.Helper()
	zigPath, err := exec.LookPath("zig")
	if err != nil {
		tb.Skip("zig not available in PATH")
	}

	dir, err := os.MkdirTemp("", "go2zig-stream-runtime-*")
	if err != nil {
		tb.Fatalf("MkdirTemp() error = %v", err)
	}
	for _, name := range []string{"api.zig", "go2zig_exports.zig", "go2zig_build_root.zig"} {
		if err := os.WriteFile(filepath.Join(dir, name), readStreamExampleFile(tb, name), 0o644); err != nil {
			tb.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}
	if libSource == "" {
		libSource = string(readStreamExampleFile(tb, "lib.zig"))
	}
	if runtimeSource == "" {
		runtimeSource = string(readStreamExampleFile(tb, "go2zig_runtime.zig"))
	}
	if err := os.WriteFile(filepath.Join(dir, "lib.zig"), []byte(libSource), 0o644); err != nil {
		tb.Fatalf("WriteFile(lib.zig) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go2zig_runtime.zig"), []byte(runtimeSource), 0o644); err != nil {
		tb.Fatalf("WriteFile(go2zig_runtime.zig) error = %v", err)
	}

	libPath := filepath.Join(dir, _go2zigDynamicLibraryName())
	cmd := exec.Command(zigPath, "build-lib", "-dynamic", "-O", "ReleaseSafe", "-femit-bin="+libPath, "go2zig_build_root.zig")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		tb.Fatalf("zig build-lib failed: %v\n%s", err, out)
	}
	return libPath
}

func prepareStreamRuntime(tb testing.TB) {
	tb.Helper()
	prepareRuntimeOnce.Do(func() {
		if _, err := exec.LookPath("zig"); err != nil {
			prepareRuntimeSkip = "zig not available in PATH"
			return
		}
		defer func() {
			if r := recover(); r != nil {
				prepareRuntimeErr = fmt.Errorf("prepare stream runtime panic: %v", r)
			}
		}()
		prepareRuntimePath = buildCustomStreamRuntime(tb, "", "")
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
