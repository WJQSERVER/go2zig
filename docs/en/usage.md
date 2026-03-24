# Usage Guide

This document explains the typical usage of `go2zig` in a "from zero to running" sequence.

## 1. Environment Preparation

### Platform Requirements

Currently supported platforms:
- **Windows/amd64** - Full support
- **Windows/arm64** - Supported by the no-cgo asm runtime
- **Linux/amd64** - Full support
- **Linux/arm64** - Supported by the no-cgo asm runtime
- **Darwin/arm64** - Dynamic loading and generated wrappers supported

Unsupported platforms:
- **Darwin/amd64** - Not currently supported
- Other operating systems

### Software Requirements

- Go `1.26`
- Zig `0.15.2`

## 2. Prepare API Description File

First, write a Zig API file, for example `api.zig`:

```zig
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
pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
```

### Supported Types

#### Fully Supported
- **Primitive types**: `bool`, `u8-u64`, `i8-i64`, `f32`, `f64`
- **Structs**: `extern struct` with nested fields
- **Enums**: `enum(integer_type)` with explicit values (e.g., `enum(u8)`, `enum(u16)`)
- **Arrays**: Fixed-length `[N]Type` and named aliases (e.g., `pub const Digest = [4]u8`)
- **Slices**: Named aliases (e.g., `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`)
- **Optionals**: `?POD` (e.g., `?u32`, `?UserKind`, `?Digest`)
- **Error handling**: `error{...}!ReturnType`

#### Special Types
- **String**: Maps to Go `string` (Zig allocates, Go frees)
- **Bytes**: Maps to Go `[]byte` (Zig allocates, Go frees)

#### Unsupported Types
- Go-specific: `map`, `chan`, `interface{}`, function types, pointers
- Zig-specific: `union`, `comptime`, `@import`
- Limited support: Optional types only support POD, slice elements cannot be String/Bytes

### Syntax Notes

- `String` and `Bytes` are conventional bridging type aliases
- Business structs must use `extern struct`
- Function declarations use `pub extern fn` or `pub export fn`
- `error union` is recommended to use named error sets, e.g., `LoginError!LoginResponse`

## 3. Write Zig Business Implementation

Then write the corresponding implementation file, for example `lib.zig`:

```zig
const api = @import("api.zig");
const rt = @import("go2zig_runtime.zig");

pub fn health() bool {
    return true;
}

pub fn login_checked(req: api.LoginRequest) api.LoginError!api.LoginResponse {
    if (rt.asSlice(req.password).len < 6) return api.LoginError.InvalidPassword;
    return .{
        .ok = true,
        .message = rt.ownString("welcome alice"),
        .token = rt.ownBytes("token-123"),
    };
}
```

### Key Functions

- `rt.asSlice` / `rt.asBytes`: Convert Go input content to Zig slice
- `rt.ownString` / `rt.ownBytes`: Hand over return values to Go for memory management
- No need to write export bridge functions manually, the generator handles it

## 4. Generate Go Wrapper and Zig Bridge Files

### Generate Source Only

```bash
go run ./cmd/go2zig -api ./api.zig -out ./gen.go -pkg main -lib basic -no-build
```

### Generate and Build Dynamic Library

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic
```

### Enable Experimental Streaming

If your Zig API uses `GoReader` / `GoWriter`, enable experimental stream support explicitly:

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic -stream-experimental
```

### Generate Without Top-Level Forwarders

If you want to keep only `Go2ZigClient` methods and skip package-level forwarding functions like `Login(...)`, use:

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic -no-top-level
```

### Generated Files

By default, the following are produced:
- `gen.go` - Go wrapper layer
- `go2zig_runtime.zig` - Zig runtime helpers
- `go2zig_exports.zig` - Zig export bridge
- `basic.dll` or `libbasic.so` - Dynamic library

## 5. Call from Go

After generation, you can use it like a normal Go SDK:

```go
package main

import "fmt"

func main() {
    // Load dynamic library
    if err := Default.Load(); err != nil {
        panic(err)
    }

    // Call top-level functions directly
    if !Health() {
        panic("Health check failed")
    }

    resp := Login(LoginRequest{
        User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
        Password: "secret-123",
    })

    // Or use the client
    client := NewGo2ZigClient("")
    if err := client.Load(); err != nil {
        panic(err)
    }

    checked, err := client.LoginChecked(LoginRequest{
        User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
        Password: "secret-123",
    })
    if err != nil {
        panic(err)
    }

    _ = resp
    _ = checked
}
```

### Two Calling Styles

- Top-level functions: `Login(...)`
- Client methods: `Default.Login(...)` or `NewGo2ZigClient(path)`

If `-no-top-level` is enabled, only client methods are generated.

### Type Mapping

For supported types:
- Zig `enum(u8)` generates Go named types and corresponding constants
- Zig named array aliases generate Go named array types
- POD slice aliases generate Go `[]T` named aliases, with automatic zero-copy input / copy output conversion
- Zig `[N]T` generates Go `[N]T` arrays, with automatic ABI conversion
- Zig `?T` currently generates `*T` on the Go side

### Experimental Stream Types

The current experimental stream bridge uses reserved names:

- `GoReader`
- `GoWriter`

They can only be used as top-level function parameters and cannot appear in:

- return values
- `extern struct` fields
- `optional`
- `slice`
- `array`

Declare them explicitly in the Zig API:

```zig
pub const GoReader = usize;
pub const GoWriter = usize;

pub extern fn copy_stream(reader: GoReader, writer: GoWriter) u64;
```

On the Zig side, use the helpers from `go2zig_runtime.zig`:

```zig
const rt = @import("go2zig_runtime.zig");

pub fn copy_stream(reader: usize, writer: usize) u64 {
    var total: u64 = 0;
    var buf: [32]u8 = undefined;
    while (true) {
        const n = rt.streamRead(reader, buf[0..]) catch |err| switch (err) {
            error.EndOfStream => break,
            else => @panic("stream read failed"),
        };
        const written = rt.streamWrite(writer, buf[0..n]) catch @panic("stream write failed");
        total += @as(u64, @intCast(written));
    }
    return total;
}
```

On the Go side, wrap standard stream values with the generated helpers:

```go
reader, err := NewGoReader(strings.NewReader("hello"))
if err != nil {
    panic(err)
}

var out bytes.Buffer
writer, err := NewGoWriter(&out)
if err != nil {
    panic(err)
}

copied := CopyStream(reader, writer)
_ = copied
```

Currently supported wrappers include:

- `io.Reader`
- `io.Writer`
- `io.ReadCloser`
- `io.WriteCloser`
- `io.Pipe`
- `*os.File`

Current limitations:

- This feature is experimental and must be enabled explicitly
- The current bridge is a synchronous block stream, not an async or full-duplex protocol
- Zig currently receives file-handle-style `usize` values under the hood

## 6. Custom Dynamic Library Path

If you don't want to use the default same-directory loading, you can specify manually:

```go
client := NewGo2ZigClient("./dist/libbasic.so")
if err := client.Load(); err != nil {
    panic(err)
}
```

## 7. How Error Returns Work

For Zig `error union`, Go side automatically generates:
- With payload: `(T, error)`
- Without payload: `error`

For example:

```zig
pub extern fn flush() FlushError!void;
```

Generates:

```go
func Flush() error
```

On failure, you get `*Go2ZigError`:
- `Code`: Zig error code
- `Message`: Currently defaults to Zig `@errorName(err)`

## 8. Common Builder Methods

If you call the generator directly in Go code, the most commonly used are:

- `WithAPI(path)`
- `WithZigSource(path)`
- `WithOutput(path)`
- `WithPackageName(name)`
- `WithLibraryName(name)`
- `WithOptimize(mode)`
- `WithTopLevelFunctions(enabled)`
- `Build()`

Typical usage:

```go
import "go2zig"

err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("basic").
    Build()
```

If your project already has a hand-written wrapper layer, you can also disable top-level forwarding functions in the Builder:

```go
import "go2zig"

err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("basic").
    WithTopLevelFunctions(false).
    Build()
```

## 9. Performance Considerations

Current implementation characteristics:
- **Advantages**: Approximately 8x faster than cgo (3.35ns vs 28.56ns)
- **Disadvantages**: Data copying required for each call
- **Suitable for**: High-frequency short call scenarios
- **Not suitable for**: Scenarios requiring zero-copy or large data transfer

## 10. FAQ

### Q1: Why can't Go find the dynamic library?

By default, it looks next to the generated `gen.go` file:
- Windows: `basic.dll`
- Linux: `libbasic.so`

If the path is different, use `NewGo2ZigClient(customPath)`.

### Q2: Why doesn't Linux main CI run bottom-level runtime live tests?

Because the no-`cgo` runtime on Linux is still being refined, current main CI focuses on stable generation, compilation and integration verification.

If you need to enable Linux runtime deep testing locally:

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

### Q3: When to generate only without building?

If you just want to see the Go wrapper and Zig bridge source code first, use `-no-build`.

### Q4: Where should I look first?

Recommended order:
1. `README.md` or `README_zh.md`
2. `docs/architecture.md` or `docs/architecture.en.md`
3. `docs/runtime.md` or `docs/runtime.en.md`
4. `docs/testing.md` or `docs/testing.en.md`
5. `examples/basic`

### Q5: Why are some types not supported?

Current design limitations:
- **Platform limitation**: Only supports Windows/Linux on `amd64` and `arm64`, plus Darwin on `arm64`
- **Type limitation**: To maintain ABI stability and performance, dynamic types are not supported
- **Memory management**: Fixed allocation pattern, cannot be customized

### Q6: How to extend support for more types?

Need to modify:
1. `internal/model/model.go` - Add new type definitions
2. `internal/parser/parser.go` - Add parsing logic
3. `internal/generator/generator.go` - Add code generation logic

Reference existing type implementations.

## 11. Debugging Tips

### Enable Verbose Logging

Currently no built-in verbose logging, but you can:
1. Check generated `gen.go` file
2. Check `go2zig_runtime.zig` and `go2zig_exports.zig`
3. Use `go test -v` to see test output

### Common Errors

1. **Type not supported**: Check if unsupported types are used
2. **Syntax error**: Ensure correct Zig syntax is used
3. **Platform not supported**: Ensure running on Windows/Linux with `amd64` or `arm64`, or on Darwin with `arm64`

## 12. Best Practices

1. **Start simple**: Test primitive types first, then gradually add complex types
2. **Use examples**: Reference code in `examples/basic/`
3. **Test coverage**: Write tests for all API functions
4. **Performance testing**: Use benchmarks to verify performance improvements
5. **Error handling**: Add error handling for all operations that may fail
