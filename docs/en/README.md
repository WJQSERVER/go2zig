# go2zig Docs

Languages: [English](README.md) | [简体中文](../zh/README.md) | [日本語](../ja/README.md)

`go2zig` currently focuses on a no-`cgo` `Go -> Zig` calling path. Its core goals are:

- Keep the Go-side experience close to a normal SDK
- Minimize extra overhead introduced by `syscall` / `cgo` at runtime
- Use the generator to unify ABI, frame structs, error protocol, and string/byte conversions

## Platform Support

Currently supported:
- `windows/amd64` - Full support with CI testing
- `windows/arm64` - Supported by the no-cgo asm runtime
- `linux/amd64` - Full support with CI testing
- `linux/arm64` - Supported by the no-cgo asm runtime
- `darwin/arm64` - Dynamic loading and generated wrappers supported

Unsupported:
- `darwin/amd64` - Not currently supported
- Other architectures - Not currently supported

## Type Support Overview

### Fully Supported Types
- Primitive types: `bool`, `u8-u64`, `i8-i64`, `f32`, `f64`
- Composite types: `extern struct`, `enum(integer_type)`, fixed-length arrays
- Special types: `String`, `Bytes`
- Slice aliases: POD slices (for example `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`)
- Optional types: `?POD` (for example `?u32`, `?UserKind`)
- Error handling: `error{...}!ReturnType`

### Unsupported Types
- Go-specific: `map`, `chan`, `interface{}`, function types, pointers
- Zig-specific: `union`, `comptime`, `@import`
- Limited support: Optional types only support POD, slice elements cannot be String/Bytes

## Recommended Reading Order

1. `docs/en/architecture.md` - Understand the overall architecture
2. `docs/en/usage.md` - Learn how to use the project
3. `docs/en/generator.md` - Understand generator details
4. `docs/en/runtime.md` - Understand the runtime design
5. `docs/en/testing.md` - Learn how testing works
6. `docs/en/ci.md` - Review CI configuration

## Quick Start

If you just want to get started quickly, begin with the repository root `README.md` (English), `README_zh.md` (Chinese), or `README_ja.md` (Japanese), then use `examples/basic` to understand the complete generation flow.

## Performance Benchmarks

A representative result recently observed on the Windows development machine:

- `BenchmarkCgoAddU64`: `28.56 ns/op`
- `BenchmarkAsmCallCAddU64`: `3.352 ns/op`

In other words, for very short synchronous calls, the current no-`cgo` asm path is about `8x` faster than cgo.

## Memory Management

The current implementation uses a fixed memory management pattern:

- **Allocator**: Zig allocates memory
- **Releaser**: Go frees memory through `go2zig_free_buf`
- **Conversion cost**: Requires copying data and managing lifetimes

## Current Limitations

1. **Platform limitation**: Only supports Windows/Linux on `amd64` and `arm64`, plus Darwin on `arm64`
2. **Type limitation**: Does not support Go-specific types like maps, channels, and interfaces
3. **Memory management**: Uses a fixed allocation pattern and does not support custom allocators
4. **Performance overhead**: Requires data copying for each call

## Future Directions

### High Priority
- Support for `?String` and `?Bytes` optionals
- Improved error diagnostics

### Medium Priority
- `union` type support
- Custom allocator interface
- Performance optimization

### Low Priority
- Generic support
- Toolchain integration
- Cross-platform improvements

## Related Documentation

### English Documentation
- [English README](../../README.md)
- [Chinese README](../../README_zh.md)
- [Japanese README](../../README_ja.md)
- [Architecture Overview](architecture.md)
- [Usage Guide](usage.md)
- [Generator Guide](generator.md)
- [Runtime Design](runtime.md)
- [Testing & Benchmarks](testing.md)
- [CI Guide](ci.md)

### 中文文档
- [架构概览](../zh/architecture.md)
- [使用指南](../zh/usage.md)
- [生成器详情](../zh/generator.md)
- [运行时设计](../zh/runtime.md)
- [测试与基准](../zh/testing.md)
- [CI 配置](../zh/ci.md)

### 日本語ドキュメント
- [アーキテクチャ概要](../ja/architecture.md)
- [使用ガイド](../ja/usage.md)
- [ジェネレータ説明](../ja/generator.md)
- [ランタイム設計](../ja/runtime.md)
- [テストとベンチマーク](../ja/testing.md)
- [CI 説明](../ja/ci.md)
