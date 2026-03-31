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
- `basic.dll`, `libbasic.so`, or `libbasic.dylib`

If you call the Go Builder programmatically with `WithDynamicBuild(false)`, the build artifact switches to a static library instead:

- Windows: `basic.lib`
- non-Windows: `libbasic.a`

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

In practice, `slice alias` support is broader than the older “POD slice” wording suggests:

- primitive / enum / array / array alias elements are supported
- named slice aliases whose elements are structs are supported
- structs can themselves contain POD slice fields

Using `String` or `Bytes` as slice elements is still unsupported.

## Actual Builder / Generate Output Behavior

`Generate(...)` only does the following:

- always writes the Go file
- writes `go2zig_runtime.zig` only when `RuntimeZig` is non-empty
- writes `go2zig_exports.zig` only when `BridgeZig` is non-empty

`Builder.Build()` fills in runtime/bridge paths by default, so it usually produces all three files; if `WithZigSource(...)` is also set, it additionally:

- writes `go2zig_build_root.zig`
- invokes `zig build-lib`

## Current Builder Methods

Beyond the commonly documented methods, the current public Builder also exposes:

- `WithHeaderOutput(path)`
- `WithRuntimeZig(path)`
- `WithBridgeZig(path)`
- `WithDynamicBuild(enabled)`
- `WithStreamExperimental(enabled)`
- `WithAPIModuleName(name)`
- `WithImplModule(name)`

The CLI currently exposes `-header`, `-runtime-zig`, `-bridge-zig`, and `-stream-experimental`; static-library builds and custom module naming are still mainly Go-Builder features.

## One Current Implementation Limitation

Although Builder and `GenerateConfig` both let you customize the `RuntimeZig` output path, `go2zig_exports.zig` still hardcodes:

```zig
const rt = @import("go2zig_runtime.zig");
```

So the most reliable setup is still to keep the runtime file next to the bridge file under the default filename.

## Why Use Frame

Benefits of frame:

- Avoids complex ABI branching for each function
- Facilitates unified error return handling
- Supports future expansion to more parameter and return value types
- Generator logic is more stable, callers can debug more easily
