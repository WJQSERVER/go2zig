# go2zig

A lightweight, high-performance code generator for Go-to-Zig FFI, inspired by [rust2go](https://github.com/ihciah/rust2go).

## Key Features

- **Zig-native API declarations** - Use standard Zig syntax as API description, no separate IDL needed
- **Automatic type bridging** - Seamlessly convert `string`, `[]byte`, nested structs, enums, slices, and arrays
- **No cgo required** - Direct assembly-based calling with up to 8x performance improvement over cgo
- **Builder pattern** - Generate Go wrappers and compile Zig dynamic libraries in one step
- **Dual API style** - Both `Client` methods and top-level functions for flexible usage

## Platform Support

### Supported Platforms
- ✅ **Windows/amd64** - Full support with CI testing
- ✅ **Linux/amd64** - Full support with CI testing

### Unsupported Platforms
- ❌ **arm64** - Planned for future implementation
- ❌ **macOS** - Not currently supported
- ❌ **Other architectures** - Not currently supported

## Requirements

- **Go** 1.26+
- **Zig** 0.15.2
- **Platform**: Windows or Linux (amd64 only)

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
- **Slices**: Named aliases like `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`
- **Optionals**: `?POD` (e.g., `?u32`, `?UserKind`, `?Digest`)

### Special Types
- **String**: Maps to Go `string` (Zig allocates, Go frees)
- **Bytes**: Maps to Go `[]byte` (Zig allocates, Go frees)

### Error Handling
- **Error unions**: `error{...}!ReturnType` mapped to Go `(T, error)` or `error`

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
- Complex error sets

### Limited Support
- Optional types only support POD (Plain Old Data)
- Slice elements cannot be `String` or `Bytes`
- Nested optionals (`??T`) not supported

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
- `mylib.dll` / `libmylib.so` - Dynamic library

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

- [Architecture Overview](docs/architecture.md)
- [Usage Guide](docs/usage.md)
- [Generator Details](docs/generator.md)
- [Runtime Design](docs/runtime.md)
- [Testing & Benchmarks](docs/testing.md)
- [CI Configuration](docs/ci.md)

## Limitations

1. **Platform**: Only amd64 architecture (Windows/Linux)
2. **Types**: No support for Go maps, channels, interfaces
3. **Memory**: Fixed allocation pattern (Zig allocates, Go frees)
4. **Performance**: Data copying required for each call
5. **Error handling**: Limited to simple error sets

## Future Roadmap

### Short-term
- arm64 architecture support
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