//go:build linux

package asmcall_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	"go2zig/asmcall"
	"go2zig/dynlib"
)

func TestCallFuncG0P3StoreSumLinux(t *testing.T) {
	if os.Getenv("GO2ZIG_RUN_LINUX_RUNTIME_TESTS") != "1" {
		t.Skip("linux runtime execution tests are disabled by default")
	}
	t.Parallel()

	addr := buildAndLookupStoreSumLinux(t)
	var slot uint64
	asmcall.CallFuncG0P3(
		unsafe.Pointer(addr),
		unsafe.Pointer(&slot),
		unsafe.Pointer(uintptr(10)),
		unsafe.Pointer(uintptr(32)),
	)
	if slot != 42 {
		t.Fatalf("slot = %d, want 42", slot)
	}
}

func TestCallFuncP3StoreSumLinux(t *testing.T) {
	if os.Getenv("GO2ZIG_RUN_LINUX_RUNTIME_TESTS") != "1" {
		t.Skip("linux runtime execution tests are disabled by default")
	}
	t.Parallel()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	addr := buildAndLookupStoreSumLinux(t)
	var slot uint64
	asmcall.CallFuncP3(
		unsafe.Pointer(addr),
		unsafe.Pointer(&slot),
		unsafe.Pointer(uintptr(7)),
		unsafe.Pointer(uintptr(8)),
	)
	if slot != 15 {
		t.Fatalf("slot = %d, want 15", slot)
	}
}

func buildAndLookupStoreSumLinux(t testing.TB) uintptr {
	t.Helper()

	zigPath, err := exec.LookPath("zig")
	if err != nil {
		t.Skip("zig not available in PATH")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "asmcall_test.zig")
	so := filepath.Join(dir, "asmcall_test.so")
	content := "pub export fn go2zig_store_sum(slot: *u64, a: u64, b: u64) void { slot.* = a + b; }\n"
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", src, err)
	}
	cmd := exec.Command(zigPath, "build-lib", "-dynamic", "-O", "ReleaseSafe", "-femit-bin="+so, src)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zig build-lib failed: %v\n%s", err, out)
	}
	lib, err := dynlib.Load(so)
	if err != nil {
		t.Fatalf("Load(%s) error = %v", so, err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	addr, err := lib.Lookup("go2zig_store_sum")
	if err != nil {
		t.Fatalf("Lookup(go2zig_store_sum) error = %v", err)
	}
	return addr
}
