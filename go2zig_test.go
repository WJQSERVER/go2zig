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

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn health() bool;
pub extern fn login(req: LoginRequest) LoginResponse;
pub extern fn rename_user(user: User, next_name: String) User;
pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
`

const integrationLib = `
const std = @import("std");
const api = @import("api.zig");
const rt = @import("go2zig_runtime.zig");

pub fn health() bool {
    return true;
}

pub fn login(req: api.LoginRequest) api.LoginResponse {
    const ok = rt.asSlice(req.password).len >= 6;
    return .{
        .ok = ok,
        .message = rt.ownString(if (ok) "welcome alice" else "bad password"),
        .token = rt.ownBytes(if (ok) "token-123" else &.{}),
    };
}

pub fn rename_user(user: api.User, next_name: api.String) api.User {
    return .{
        .id = user.id,
        .name = rt.ownString(rt.asSlice(next_name)),
        .email = rt.ownString(rt.asSlice(user.email)),
    };
}

pub fn login_checked(req: api.LoginRequest) api.LoginError!api.LoginResponse {
    if (rt.asSlice(req.password).len < 6) return api.LoginError.InvalidPassword;
    return .{
        .ok = true,
        .message = rt.ownString("welcome alice"),
        .token = rt.ownBytes("token-123"),
    };
}
`

func TestGenerateWritesGoFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	apiPath := filepath.Join(dir, "api.zig")
	outPath := filepath.Join(dir, "gen.go")
	writeFile(t, apiPath, integrationAPI)

	if err := Generate(GenerateConfig{
		API:         apiPath,
		Output:      outPath,
		PackageName: "sample",
		LibraryName: "sample",
		RuntimeZig:  filepath.Join(dir, "go2zig_runtime.zig"),
		BridgeZig:   filepath.Join(dir, "go2zig_exports.zig"),
		APIModule:   "api.zig",
		ImplModule:  "lib.zig",
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(gen) error = %v", err)
	}
	text := string(content)
	checks := []string{
		"//go:build amd64 && (windows || linux)",
		"package sample",
		"type Go2ZigClient struct",
		"var Default = NewGo2ZigClient(\"\")",
		"func (c *Go2ZigClient) Login(req LoginRequest) LoginResponse",
		"func Login(req LoginRequest) LoginResponse",
		"func (c *Go2ZigClient) LoginChecked(req LoginRequest) (LoginResponse, error)",
		"type Go2ZigError struct",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("generated file missing %q\n%s", check, text)
		}
	}

	runtimeText, err := os.ReadFile(filepath.Join(dir, "go2zig_runtime.zig"))
	if err != nil {
		t.Fatalf("ReadFile(runtime) error = %v", err)
	}
	if !strings.Contains(string(runtimeText), "std.heap.smp_allocator") {
		t.Fatalf("runtime zig should use smp_allocator\n%s", runtimeText)
	}
	bridgeText, err := os.ReadFile(filepath.Join(dir, "go2zig_exports.zig"))
	if err != nil {
		t.Fatalf("ReadFile(bridge) error = %v", err)
	}
	if !strings.Contains(string(bridgeText), "pub export fn go2zig_call_login") {
		t.Fatalf("bridge zig missing login export\n%s", bridgeText)
	}
	if !strings.Contains(string(bridgeText), "catch |err|") {
		t.Fatalf("bridge zig should include error union catch path\n%s", bridgeText)
	}
}

func TestGenerateValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	if err := Generate(GenerateConfig{}); err == nil || !strings.Contains(err.Error(), "api path is required") {
		t.Fatalf("Generate() error = %v, want api path validation", err)
	}
	if err := Generate(GenerateConfig{API: "api.zig"}); err == nil || !strings.Contains(err.Error(), "output path is required") {
		t.Fatalf("Generate() error = %v, want output path validation", err)
	}
}

func TestBuilderBuildsZigDynamicLibrary(t *testing.T) {
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

	for _, path := range []string{
		outPath,
		filepath.Join(dir, "go2zig_runtime.zig"),
		filepath.Join(dir, "go2zig_exports.zig"),
		filepath.Join(dir, outputLibraryFilename("sample", true)),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated artifact missing %s: %v", path, err)
		}
	}
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(gen) error = %v", err)
	}
	if !strings.Contains(string(content), "func (c *Go2ZigClient) RenameUser(user User, nextName string) User") {
		t.Fatalf("generated file missing ergonomic RenameUser wrapper\n%s", string(content))
	}
	if !strings.Contains(string(content), "func (c *Go2ZigClient) LoginChecked(req LoginRequest) (LoginResponse, error)") {
		t.Fatalf("generated file missing error-return wrapper\n%s", string(content))
	}
}

func TestBuilderGenerateOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	apiPath := filepath.Join(dir, "api.zig")
	outPath := filepath.Join(dir, "gen.go")
	writeFile(t, apiPath, integrationAPI)

	if err := NewBuilder().
		WithAPI(apiPath).
		WithOutput(outPath).
		WithPackageName("sample").
		WithLibraryName("sample").
		Build(); err != nil {
		t.Fatalf("Builder.Build() error = %v", err)
	}

	for _, path := range []string{
		outPath,
		filepath.Join(dir, "go2zig_runtime.zig"),
		filepath.Join(dir, "go2zig_exports.zig"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file missing %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, outputLibraryFilename("sample", true))); !os.IsNotExist(err) {
		t.Fatalf("dynamic library should not exist in generate-only mode, stat error = %v", err)
	}
}

func TestBuilderValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	if err := NewBuilder().Build(); err == nil || !strings.Contains(err.Error(), "output path is required") {
		t.Fatalf("Build() error = %v, want output path validation", err)
	}
	if err := NewBuilder().WithOutput("gen.go").Build(); err == nil || !strings.Contains(err.Error(), "api path or zig source path is required") {
		t.Fatalf("Build() error = %v, want api/zig path validation", err)
	}
}

func TestBuilderGeneratedProgramRuns(t *testing.T) {
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" {
		t.Skip("end-to-end no-cgo runtime test currently targets windows/amd64")
	}
	if _, err := exec.LookPath("zig"); err != nil {
		t.Skip("zig not available in PATH")
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/sample\n\ngo 1.26.0\n\nrequire go2zig v0.0.0\n\nreplace go2zig => "+filepath.ToSlash(mustAbs(t, "."))+"\n")
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
	if err := Default.Load(); err != nil {
		panic(err)
	}
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
	checked, err := LoginChecked(LoginRequest{
		User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
		Password: "secret-123",
	})
	if err != nil {
		panic(err)
	}
	_, err = LoginChecked(LoginRequest{
		User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
		Password: "bad",
	})
	if err == nil {
		panic("expected login_checked error")
	}
	fmt.Printf("%s|%s|%s|%s|%v", resp.Message, string(resp.Token), renamed.Name, checked.Message, err != nil)
}
`)

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run failed: %v\n%s", err, out)
	}
	if got, want := strings.TrimSpace(string(out)), "welcome alice|token-123|ally|welcome alice|true"; got != want {
		t.Fatalf("program output = %q, want %q", got, want)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func mustAbs(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Abs(%s) error = %v", path, err)
	}
	return abs
}
