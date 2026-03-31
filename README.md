# go2zig

Languages: [English](README.md) | [简体中文](README_zh.md) | [日本語](README_ja.md)

A lightweight, high-performance code generator for Go-to-Zig FFI, inspired by [rust2go](https://github.com/ihciah/rust2go).

## Key Features

- **Zig-native API declarations** - Use standard Zig syntax as API description, no separate IDL needed
- **Automatic type bridging** - Seamlessly convert `string`, `[]byte`, nested structs, enums, slices, and arrays
- **No cgo required** - Direct assembly-based calling with up to 8x performance improvement over cgo
- **Builder pattern** - Generate Go wrappers and compile Zig dynamic libraries in one step
- **Dual API style** - Both `Client` methods and top-level functions for flexible usage

## Experimental Streaming

`go2zig` now includes an experimental streaming bridge based on `GoReader` and `GoWriter`.

- Go-side helpers currently support `io.Reader`, `io.Writer`, `io.ReadCloser`, `io.WriteCloser`, and `io.Pipe`
- Go-side helpers also support `*os.File`
- Zig side consumes stream handles synchronously through `rt.streamRead(...)` and `rt.streamWrite(...)`
- The current implementation is a block-based, synchronous stream bridge; it is not yet an async or full-duplex protocol layer
- Streaming must be enabled explicitly with `WithStreamExperimental(true)` or `-stream-experimental`

## Platform Tiers

Borrowing the idea of support tiers from `purego`, the current main verified set is:

- **Tier 1** - `windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64`, `darwin/arm64`

These targets currently run tests, benchmarks, and build checks in main CI.

## Platform Support

### Supported Platforms
- ✅ **Windows/amd64** - Full support with CI testing
- ✅ **Windows/arm64** - Supported by the no-cgo asm runtime
- ✅ **Linux/amd64** - Full support with CI testing
- ✅ **Linux/arm64** - Supported by the no-cgo asm runtime
- ✅ **Darwin/arm64** - Dynamic loading and generated wrappers supported

### Unsupported Platforms
- ❌ **Darwin/amd64** - Not supported
- ❌ **Other architectures** - Not currently supported

## Requirements

- **Go** 1.26+
- **Zig** 0.15.2
- **Platform**: Windows/Linux (`amd64` and `arm64`) or Darwin (`arm64` only)

## Supported Types

### Primitive Types
- `bool`
- `u8`, `u16`, `u32`, `u64`, `usize`
- `i8`, `i16`, `i32`, `i64`, `isize`
- `f32`, `f64`

### Composite Types
- **Structs**: `extern struct` with nested fields
- **Enums**: `enum(integer_type)` with explicit values
- **Arrays**: Fixed-length `[N]Type` and named aliases like `pub const Digest = [4]u8`
- **Slices**: Named aliases like `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`, including aliases whose elements are structs
- **Optionals**: `?POD` (e.g., `?u32`, `?UserKind`, `?Digest`)

### Special Types
- **String**: Maps to Go `string` (Zig allocates, Go frees)
- **Bytes**: Maps to Go `[]byte` (Zig allocates, Go frees)

### Error Handling
- **Error unions**: Named or inline `error{...}!ReturnType` mapped to Go `(T, error)` or `error`

## Unsupported Types

### Go-specific Types
- `map[K]V`
- `chan T`
- `interface{}`
- Function types (`func(...)`)
- Pointers (except in String/Bytes)
- `unsafe.Pointer`

### Zig-specific Types
- `union`
- `comptime`
- `@import`

### Limited Support
- Optional support is currently centered on POD-style shapes
- Slice elements cannot be `String` or `Bytes`
- Experimental stream types (`GoReader` / `GoWriter`) can only appear as top-level function parameters

## Quick Start

### 1. Define API in Zig

```zig
// api.zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Bytes = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const UserKind = enum(u8) {
    guest,
    member,
    admin,
};

pub const User = extern struct {
    id: u64,
    kind: UserKind,
    name: String,
    email: String,
};

pub extern fn health() bool;
pub extern fn login(user: User, password: String) String;
```

### 2. Generate Go Wrapper and Build

```bash
# Generate only
go run ./cmd/go2zig -api ./api.zig -out ./gen.go -pkg main -lib mylib -no-build

# Generate and build dynamic library
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib mylib

# Generate without top-level forwarding functions
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib mylib -no-top-level
```

### 3. Use in Go

```go
package main

import "fmt"

func main() {
    // Load the dynamic library
    if err := Default.Load(); err != nil {
        panic(err)
    }

    // Call functions directly
    if !Health() {
        panic("Health check failed")
    }

    // Or use the client
    client := NewGo2ZigClient("")
    if err := client.Load(); err != nil {
        panic(err)
    }

    greeting := client.Login(User{
        ID:     1,
        Kind:   UserKindMember,
        Name:   "alice",
        Email:  "alice@example.com",
    }, "password123")
    fmt.Println(greeting)
}
```

## Performance

Based on benchmarks on Windows/amd64:

| Method | Performance | Relative |
|--------|-------------|----------|
| **asmcall (go2zig)** | **3.35 ns/op** | **1x** |
| cgo | 28.56 ns/op | ~8.5x slower |

The no-cgo approach provides approximately **8x performance improvement** for short, synchronous FFI calls.

## Memory Management

- **Allocation**: Zig side allocates memory for strings, bytes, and slices
- **Deallocation**: Go side frees memory through `go2zig_free_buf`
- **Pattern**: Copy-in for inputs, copy-out for outputs
- **Overhead**: Data copying required for each call

## Generated Files

When you run the generator, it produces:

- `gen.go` - Go wrapper with types and functions
- `go2zig_runtime.zig` - Zig runtime helpers
- `go2zig_exports.zig` - Zig export bridge functions
- `mylib.dll` / `libmylib.so` / `libmylib.dylib` - Dynamic library

When `Build()` also compiles Zig, it additionally writes `go2zig_build_root.zig` before invoking `zig build-lib`.

## Builder API

For programmatic usage in Go:

```go
import "go2zig"

err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("mylib").
    Build()
```

If your project already provides its own higher-level wrapper layer and you want to avoid generated package-level forwarding functions like `Login(...)`, disable them explicitly:

```go
err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("mylib").
    WithTopLevelFunctions(false).
    Build()
```

This keeps `Go2ZigClient` methods such as `client.Login(...)`, but skips top-level forwarding functions and helps avoid symbol collisions with hand-written wrappers.

The current Builder API also includes `WithHeaderOutput`, `WithRuntimeZig`, `WithBridgeZig`, `WithDynamicBuild`, `WithStreamExperimental`, `WithAPIModuleName`, and `WithImplModule`.

## Examples

See `examples/basic/` for a complete working example demonstrating:

- Primitive types
- Structs and nested structs
- Enums with explicit values
- Arrays and array aliases
- Slices and slice aliases
- Optionals
- Error unions
- String and Bytes handling

## Documentation

- [Docs Home](docs/en/README.md)
- [Architecture Overview](docs/en/architecture.md)
- [Usage Guide](docs/en/usage.md)
- [Generator Guide](docs/en/generator.md)
- [Runtime Design](docs/en/runtime.md)
- [Testing & Benchmarks](docs/en/testing.md)
- [CI Guide](docs/en/ci.md)

## Limitations

1. **Platform**: Only Windows/Linux on `amd64` and `arm64`, plus Darwin on `arm64`
2. **Types**: No support for Go maps, channels, interfaces
3. **Memory**: Fixed allocation pattern (Zig allocates, Go frees)
4. **Performance**: Data copying required for each call
5. **Runtime loading**: If you skip explicit `Load()` and the first lazy load fails, generated call paths currently panic

## Future Roadmap

### Short-term
- Support for `?String` and `?Bytes` optionals
- Better error diagnostics

### Medium-term
- `union` type support
- Custom allocator interface
- Performance optimizations

### Long-term
- Generic type support
- Toolchain integration
- Cross-platform improvements

## Contributing

1. Run tests: `go test ./...`
2. Run benchmarks: `go test -bench . ./asmcall`
3. Test on both Windows and Linux
4. Update documentation for new features

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).
