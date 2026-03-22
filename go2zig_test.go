package go2zig

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const integrationAPI = `
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Bytes = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const User = extern struct {
    id: u64,
    name: String,
    email: String,
};

pub const LoginRequest = extern struct {
    user: User,
    password: String,
};

pub const LoginResponse = extern struct {
    ok: bool,
    message: String,
    token: Bytes,
};

pub extern fn health() bool;
pub extern fn login(req: LoginRequest) LoginResponse;
pub extern fn rename_user(user: User, next_name: String) User;
`

const integrationLib = `
const std = @import("std");
const api = @import("api.zig");

fn allocCopy(bytes: []const u8) [*]u8 {
    const size = if (bytes.len == 0) 1 else bytes.len;
    const ptr = std.heap.c_allocator.alloc(u8, size) catch @panic("alloc failed");
    if (bytes.len > 0) {
        @memcpy(ptr[0..bytes.len], bytes);
    }
    return ptr.ptr;
}

fn asSlice(value: api.String) []const u8 {
    if (value.len == 0) return "";
    return value.ptr[0..value.len];
}

fn ownString(text: []const u8) api.String {
    return .{ .ptr = allocCopy(text), .len = text.len };
}

fn ownBytes(bytes: []const u8) api.Bytes {
    return .{ .ptr = allocCopy(bytes), .len = bytes.len };
}

pub export fn go2zig_free_buf(ptr: ?*anyopaque, len: usize) void {
    if (ptr) |value| {
        const size = if (len == 0) 1 else len;
        std.heap.c_allocator.free(@as([*]u8, @ptrCast(value))[0..size]);
    }
}

pub export fn health() bool {
    return true;
}

pub export fn login(req: api.LoginRequest) api.LoginResponse {
    const ok = asSlice(req.password).len >= 6;
    return .{
        .ok = ok,
        .message = ownString(if (ok) "welcome alice" else "bad password"),
        .token = ownBytes(if (ok) "token-123" else &.{}),
    };
}

pub export fn rename_user(user: api.User, next_name: api.String) api.User {
    return .{
        .id = user.id,
        .name = ownString(asSlice(next_name)),
        .email = ownString(asSlice(user.email)),
    };
}
`

func TestGenerateWritesGoFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	apiPath := filepath.Join(dir, "api.zig")
	outPath := filepath.Join(dir, "gen.go")
	if err := os.WriteFile(apiPath, []byte(integrationAPI), 0o644); err != nil {
		t.Fatalf("WriteFile(api) error = %v", err)
	}

	if err := Generate(GenerateConfig{
		API:         apiPath,
		Output:      outPath,
		PackageName: "sample",
		LibraryName: "sample",
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(gen) error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "package sample") {
		t.Fatalf("generated file missing package name\n%s", text)
	}
	if !strings.Contains(text, "func Login(req LoginRequest) LoginResponse") {
		t.Fatalf("generated file missing top-level Login wrapper\n%s", text)
	}
}

func TestBuilderBuildsZigLibrary(t *testing.T) {
	zigPath, err := exec.LookPath("zig")
	if err != nil {
		t.Skip("zig not available in PATH")
	}
	_ = zigPath

	dir := t.TempDir()
	apiPath := filepath.Join(dir, "api.zig")
	libPath := filepath.Join(dir, "lib.zig")
	outPath := filepath.Join(dir, "gen.go")

	writeFile(t, apiPath, integrationAPI)
	writeFile(t, libPath, integrationLib)

	if err := NewBuilder().
		WithAPI(apiPath).
		WithZigSource(libPath).
		WithOutput(outPath).
		WithPackageName("main").
		WithLibraryName("sample").
		Build(); err != nil {
		t.Fatalf("Builder.Build() error = %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("generated go file missing: %v", err)
	}
	libFile := filepath.Join(dir, "libsample.a")
	if _, err := os.Stat(libFile); err != nil {
		t.Fatalf("zig static library missing: %v", err)
	}
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(gen) error = %v", err)
	}
	if !strings.Contains(string(content), "func (Client) RenameUser(user User, nextName string) User") {
		t.Fatalf("generated file missing ergonomic RenameUser wrapper\n%s", string(content))
	}
}

func TestBuilderGeneratedProgramRuns(t *testing.T) {
	zigPath, err := exec.LookPath("zig")
	if err != nil {
		t.Skip("zig not available in PATH")
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/sample\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(dir, "api.zig"), integrationAPI)
	writeFile(t, filepath.Join(dir, "lib.zig"), integrationLib)

	outPath := filepath.Join(dir, "gen.go")
	if err := NewBuilder().
		WithAPI(filepath.Join(dir, "api.zig")).
		WithZigSource(filepath.Join(dir, "lib.zig")).
		WithOutput(outPath).
		WithPackageName("main").
		WithLibraryName("sample").
		Build(); err != nil {
		t.Fatalf("Builder.Build() error = %v", err)
	}

	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	if !Health() {
		panic("health check failed")
	}
	resp := Login(LoginRequest{
		User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
		Password: "secret-123",
	})
	if !resp.OK {
		panic("login failed")
	}
	renamed := RenameUser(User{ID: 7, Name: "alice", Email: "alice@example.com"}, "ally")
	fmt.Printf("%s|%s|%s", resp.Message, string(resp.Token), renamed.Name)
}
`)

	ccPath := makeCCWrapper(t, dir, zigPath)
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1", "CC="+ccPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if isToolchainIssue(string(out)) {
			t.Skipf("cgo toolchain unavailable: %v\n%s", err, out)
		}
		t.Fatalf("go run failed: %v\n%s", err, out)
	}
	if got, want := strings.TrimSpace(string(out)), "welcome alice|token-123|ally"; got != want {
		t.Fatalf("program output = %q, want %q", got, want)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func makeCCWrapper(t *testing.T, dir, zigPath string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		path := filepath.Join(dir, "zigcc.cmd")
		writeFile(t, path, "@\""+zigPath+"\" cc %*\r\n")
		return path
	}
	path := filepath.Join(dir, "zigcc.sh")
	writeFile(t, path, "#!/bin/sh\nexec \""+zigPath+"\" cc \"$@\"\n")
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("Chmod(%s) error = %v", path, err)
	}
	return path
}

func isToolchainIssue(output string) bool {
	markers := []string{
		"C compiler",
		"executable file not found",
		"file does not exist",
		"CreateProcess",
		"is not recognized",
		"cannot find",
	}
	for _, marker := range markers {
		if strings.Contains(strings.ToLower(output), strings.ToLower(marker)) {
			return true
		}
	}
	return false
}
