# Architecture Overview

`go2zig` currently consists of four layers:

## 1. Zig API Description Layer

Users define APIs using restricted Zig declarations, for example:

```zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const User = extern struct {
    id: u64,
    name: String,
};

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
```

This layer does not directly participate in export implementation, but serves as input for the generator.

### Supported API Syntax

- **Primitive types**: `bool`, `u8-u64`, `i8-i64`, `f32`, `f64`
- **Structs**: `extern struct` with nested fields
- **Enums**: `enum(integer_type)` with explicit values
- **Arrays**: Fixed-length `[N]Type` and named aliases
- **Slices**: Named aliases (e.g., `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`)
- **Optionals**: `?POD` (e.g., `?u32`, `?UserKind`)
- **Error handling**: `error{...}!ReturnType`
- **Special types**: `String`, `Bytes`, and experimental `GoReader` / `GoWriter`

### Unsupported Syntax

- Go-specific: `map`, `chan`, `interface{}`, function types, pointers
- Zig-specific: `union`, `comptime`, `@import`
- Limited support: Optional types only support POD, slice elements cannot be String/Bytes

## 2. Go Parsing and Model Layer

Related locations:
- `internal/parser`
- `internal/model`
- `internal/names`

### Responsibilities

- Parse Zig's `extern struct`, `extern fn` / `export fn`
- Model primitive types, `String`, `Bytes`, struct, error union uniformly
- Handle Go naming conventions (e.g., `rename_user -> RenameUser`)
- Validate type support and constraints

### Type Model

`internal/model/model.go` defines the type system:
- `TypeKind` enum: `TypeVoid`, `TypePrimitive`, `TypeString`, `TypeBytes`, `TypeGoReader`, `TypeGoWriter`, `TypeStruct`, `TypeEnum`, `TypeOptional`, `TypeSlice`, `TypeArray`
- `PrimitiveInfo`: Maps Zig/Go/C types
- `TypeRef`: Type reference, supports nesting and aliases

### Parser

`internal/parser/parser.go` uses regex to parse Zig declarations:
- `structPattern`: Parse `extern struct`
- `enumPattern`: Parse `enum(type)`
- `slicePattern`: Parse slice aliases
- `arrayAliasPattern`: Parse array aliases
- `funcPattern`: Parse function declarations

## 3. Code Generation Layer

Related locations:
- `internal/generator`
- `go2zig.go`
- `cmd/go2zig`

### Responsibilities

- Generate Go wrapper layer `gen.go`
- Generate Zig runtime helpers `go2zig_runtime.zig`
- Generate Zig export bridge `go2zig_exports.zig`
- Call `zig build-lib -dynamic` to produce dynamic library

### Generated Files

1. **`gen.go`**:
   - Defines structs and functions exposed to Go business logic
   - Generates Go named types and constants for Zig enums
   - Generates Go named array types for array aliases
   - Generates Go named slice and ABI helpers for POD slice aliases
   - Generates ABI conversion helpers for fixed-length arrays
   - Generates optional tagged wrapper helpers
   - Generates `Go2ZigClient`
   - Default generates `Default` instance and top-level forwarding functions
   - Defines ABI structures and frame structures
   - Handles string, byte, struct, error conversions

2. **`go2zig_runtime.zig`**:
   - Provides `asSlice` / `asBytes`
   - Provides `asXxx` / `ownXxx` for named slice aliases
   - Provides `ownString` / `ownBytes`
   - Provides return value memory release helpers
   - Provides `ErrorInfo`, `okError`, `makeError`
   - Provides `toOptional_xxx` / `fromOptional_xxx` for optional wrappers

3. **`go2zig_exports.zig`**:
   - Generates stable `frame` ABI for each function
   - Exports unified `go2zig_call_<name>` symbols
   - Degrades Zig `error union` to `frame.err + frame.out`
   - Bridges optional wrappers with Zig native optionals

### Builder API

`go2zig.go` provides a Builder pattern:
- `WithAPI(path)`: Set API file path
- `WithZigSource(path)`: Set Zig source file path
- `WithOutput(path)`: Set output file path
- `WithPackageName(name)`: Set Go package name
- `WithLibraryName(name)`: Set library name
- `WithOptimize(mode)`: Set optimization level
- `WithHeaderOutput(path)`: Emit a Zig header
- `WithRuntimeZig(path)` / `WithBridgeZig(path)`: Customize generated file locations
- `WithDynamicBuild(enabled)`: Switch between dynamic and static library builds
- `WithTopLevelFunctions(enabled)`: Control top-level forwarding generation
- `WithStreamExperimental(enabled)`: Enable the experimental stream bridge
- `WithAPIModuleName(name)` / `WithImplModule(name)`: Override Zig `@import` module names
- `Build()`: Execute generation and build

## 4. Runtime Call Layer

Related locations:
- `asmcall`
- `dynlib`

### Responsibilities

- `dynlib` handles loading dynamic libraries and exported symbols by platform
- `asmcall` handles high-frequency function calls without going through `cgo`
- Generated Go wrapper layer only cares about frame struct and business signatures, no need for users to write `unsafe`/ABI details

### `asmcall`

Provides two capabilities:
- `CallFuncG0P*`: Switch to `g0` stack to execute target function
- `CallFuncP*`: Execute target function directly on goroutine stack

Design motivation:
- For high-frequency short calls, `cgo`'s scheduling and stack switching costs are usually too high
- Through fixed assembly glue, overhead can be controlled to a more manageable level

### `dynlib`

Uses different implementations on different platforms:
- **Windows**: Based on system DLL loading interface (`syscall.LoadDLL`)
- **Linux**: Based on `dlopen` / `dlsym` / `dlclose` (via assembly calls)
- **Darwin**: Based on `dlopen` / `dlsym` / `dlclose`

### Error Protocol

Current `error union -> Go error` uses a fixed protocol:
- Zig frame contains `err: ErrorInfo`
- `ErrorInfo` structure:
  - `code: u32`
  - `text: api.String`
- Go side uniformly generates:
  - With return value: `(T, error)`
  - Without return value: `error`

### Optional Protocol

Currently supports `optional POD`:
- `?primitive`
- `?enum`
- `?array` / `?array alias`

Go public types default to `*T`, but ABI layer does not directly depend on Zig native optional layout, instead uses explicit tagged wrapper:
- Go ABI: `is_set + value`
- Zig runtime: `Optional_xxx`
- Zig bridge: `toOptional_xxx` / `fromOptional_xxx`

### Slice / Struct Lifecycle

When slice elements themselves contain slice fields, the generator additionally generates `keep` aggregation results to ensure:
- Temporary ABI backing buffers during input phase are not reclaimed before call ends
- Backing buffers of nested slice fields are also kept alive

Return value side uses:
1. Element-by-element `own` to restore Go values
2. Release dynamic fields inside elements
3. Finally release outer return buffer

## Currently Supported Platforms

- `windows/amd64`
- `windows/arm64`
- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

Where:
- All five currently supported targets run tests, benchmarks, and build checks in main CI
- Linux jobs additionally enable bottom-level runtime execution tests through `GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1`

## Performance Characteristics

### Advantages
- Approximately 8x faster than cgo (3.35ns vs 28.56ns)
- No cgo dependency required
- Type-safe

### Disadvantages
- Data copying required for each call
- Only supports Windows/Linux on `amd64` / `arm64`, plus Darwin on `arm64`
- Limited type support

## Extension Points

### Short-term Extensions
- Support `?String` and `?Bytes` optional types
- Improved error diagnostics

### Medium-term Extensions
- `union` type support
- Custom allocator interface
- Performance optimization

### Long-term Extensions
- Generic support
- Toolchain integration
- Cross-platform improvements
