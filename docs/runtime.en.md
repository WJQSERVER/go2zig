# Runtime Design

## Overall Goal

Avoid `cgo` and let Go directly call Zig exported functions through dynamic library symbol addresses.

## `asmcall`

`asmcall` provides two capabilities:

- `CallFuncG0P*`: Switch to `g0` stack to execute target function
- `CallFuncP*`: Execute target function directly on goroutine stack

Design motivation:

- For high-frequency short calls, `cgo`'s scheduling and stack switching costs are usually too high
- Through fixed assembly glue, overhead can be controlled to a more manageable level

## `dynlib`

`dynlib` uses different implementations on different platforms:

- Windows: Based on system DLL loading interface
- Linux: Based on `dlopen` / `dlsym` / `dlclose`
- Darwin: Based on `dlopen` / `dlsym` / `dlclose`

Linux path currently supports:

- Main CI live testing
- Dynamic library generation and loading
- Benchmark coverage

To manually enable Linux runtime live testing:

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

## Error Protocol

Current `error union -> Go error` uses a fixed protocol:

- Zig frame contains `err: ErrorInfo`
- `ErrorInfo` structure:
  - `code: u32`
  - `text: api.String`
- Go side uniformly generates:
  - With return value: `(T, error)`
  - Without return value: `error`

This is more stable than directly exposing Zig error sets because:

- ABI is fixed
- Go side only consumes standard `error`
- Doesn't require Go side to understand Zig enum sets

## Current New Type Capabilities

In addition to primitive types, `String`, `Bytes`, struct, currently also supports:

- Zig enums with integer base types, e.g., `enum(u8)`, `enum(u16)`
- POD slice aliases, e.g., `extern struct { ptr: ?[*]const u16, len: usize }`
- Fixed-length arrays, e.g., `[4]u8`, `[3]u16`, `[2]UserKind`
- Named slice aliases in `[]struct`-style shapes

Current slice-alias support is broader than just “POD slices” and mainly includes:

- Primitive numeric types
- Integer-based enums
- Fixed-length arrays
- Named POD slice aliases
- Structs recursively composed from the supported element forms above

## Optional Protocol

The stable optional shapes currently covered are mainly `optional POD`:

- `?primitive`
- `?enum`
- `?array` / `?array alias`

On the Go side, public optionals map to `*T`. Current tests and examples are also centered on these POD forms.

The ABI layer does not directly depend on Zig native optional layout and instead uses an explicit tagged wrapper:

- Go ABI: `is_set + value`
- Zig runtime: `Optional_xxx`
- Zig bridge: `toOptional_xxx` / `fromOptional_xxx`

Benefits of this approach:

- ABI is more stable
- Go side expression is more natural
- Facilitates future expansion to more complex optional combinations

## Slice / Struct Lifecycle

When slice elements themselves contain slice fields, the generator additionally generates `keep` aggregation results to ensure:

- Temporary ABI backing buffers during input phase are not reclaimed before call ends
- Backing buffers of nested slice fields are also kept alive

Return value side uses:

1. Element-by-element `own` to restore Go values
2. Release dynamic fields inside elements
3. Finally release outer return buffer

Array bridging currently uses element-by-element conversion helpers, which allows:

- Keeping ABI rules clear
- Reusing existing element-level conversion logic
- Reserving space for future support of more complex element types

## Runtime Loading Behavior

The generated `Go2ZigClient` lazily loads the dynamic library and caches resolved symbols:

- Calling `client.Load()` explicitly returns a normal Go `error`
- Calling a generated method also triggers lazy loading automatically
- If that automatic load fails, the current call path `panic(err)` instead of returning the load failure as a normal result

So if you need predictable error handling, it is best to call `Load()` explicitly first.
