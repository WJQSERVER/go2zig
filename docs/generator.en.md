# Generator Guide

The `go2zig` generator is not intended to be a general Zig binding system, but rather to generate a fixed protocol around the current no-`cgo` runtime.

## Generated Files

Calling:

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -zig ./examples/basic/lib.zig -out ./examples/basic/gen.go -pkg main -lib basic
```

Usually generates:

- `gen.go`
- `go2zig_runtime.zig`
- `go2zig_exports.zig`
- `basic.dll` or `libbasic.so`

## `gen.go` Responsibilities

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

## `go2zig_runtime.zig` Responsibilities

- Provides `asSlice` / `asBytes`
- Provides `asXxx` / `ownXxx` for named slice aliases
- Provides `ownString` / `ownBytes`
- Provides return value memory release helpers
- Provides `ErrorInfo`, `okError`, `makeError`
- Provides `toOptional_xxx` / `fromOptional_xxx` for optional wrappers

## `go2zig_exports.zig` Responsibilities

- Generates stable `frame` ABI for each function
- Exports unified `go2zig_call_<name>` symbols
- Degrades Zig `error union` to `frame.err + frame.out`
- Bridges optional wrappers with Zig native optionals

## Current Type Coverage

- primitive
- enum
- array / array alias
- slice alias
- struct
- `[]struct`
- `optional POD`
- `error union`

## Why Use Frame

Benefits of frame:

- Avoids complex ABI branching for each function
- Facilitates unified error return handling
- Supports future expansion to more parameter and return value types
- Generator logic is more stable, callers can debug more easily