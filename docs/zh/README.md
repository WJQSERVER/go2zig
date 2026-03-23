# go2zig 文档

Languages: [English](../en/README.md) | [简体中文](README.md) | [日本語](../ja/README.md)

`go2zig` 当前聚焦在无 `cgo` 的 `Go -> Zig` 调用链，核心目标是：

- 在 Go 侧保留接近普通 SDK 的调用体验
- 在运行时侧尽量减少 `syscall` / `cgo` 带来的额外开销
- 用生成器把 ABI、frame、错误协议、字符串/字节串转换统一收敛

## 平台支持

当前支持：
- `windows/amd64` - 完全支持，包含 CI 测试
- `windows/arm64` - 无 cgo 汇编运行时已支持
- `linux/amd64` - 完全支持，包含 CI 测试
- `linux/arm64` - 无 cgo 汇编运行时已支持
- `darwin/amd64` - 已支持动态加载与生成包装
- `darwin/arm64` - 已支持动态加载与生成包装

不支持：
- 其他架构 - 当前不支持

## 类型支持概览

### 完全支持的类型
- 基础类型：`bool`、`u8-u64`、`i8-i64`、`f32`、`f64`
- 复合类型：`extern struct`、`enum(整数类型)`、固定长度数组
- 特殊类型：`String`、`Bytes`
- 切片别名：POD 切片（如 `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`）
- 可选类型：`?POD`（如 `?u32`、`?UserKind`）
- 错误处理：`error{...}!ReturnType`

### 不支持的类型
- Go 特有：`map`、`chan`、`interface{}`、函数类型、指针
- Zig 特有：`union`、`comptime`、`@import`
- 有限支持：可选类型仅支持 POD，切片元素不能是 String/Bytes

## 建议阅读顺序

1. `docs/architecture.md` - 了解整体架构
2. `docs/usage.md` - 学习使用方法
3. `docs/generator.md` - 了解生成器细节
4. `docs/runtime.md` - 了解运行时设计
5. `docs/testing.md` - 了解测试方法
6. `docs/ci.md` - 了解 CI 配置

## 快速开始

如果你只是想快速上手，可以先看仓库根目录的 `README.md`（英文）或 `README_zh.md`（中文），再结合 `examples/basic` 理解完整生成流程。

## 性能基准

当前在 Windows 开发机上最近一次实测到的一个代表性结果：

- `BenchmarkCgoAddU64`: `28.56 ns/op`
- `BenchmarkAsmCallCAddU64`: `3.352 ns/op`

也就是说，在极短同步调用上，当前无 `cgo` asm 路径大约比 cgo 快 `8x` 左右。

## 内存管理

当前实现采用固定内存管理模式：

- **分配方**：Zig 负责分配内存
- **释放方**：Go 负责释放内存（通过 `go2zig_free_buf`）
- **转换开销**：需要复制数据并管理生命周期

## 当前限制

1. **平台限制**：仅支持 Windows、Linux 和 Darwin 上的 `amd64` / `arm64`
2. **类型限制**：不支持 Go 的 map、channel、interface 等特有类型
3. **内存管理**：固定的分配模式，无法自定义分配器
4. **性能开销**：每次调用需要数据复制

## 后续扩展方向

### 高优先级
- 支持 `?String` 和 `?Bytes` 可选类型
- 改进错误诊断

### 中优先级
- `union` 类型支持
- 自定义分配器接口
- 性能优化

### 低优先级
- 泛型支持
- 工具链集成
- 跨平台改进

## 相关文档

### 中文文档
- [英文 README](../../README.md)
- [中文 README](../../README_zh.md)
- [日文 README](../../README_ja.md)
- [架构概览](architecture.md)
- [使用指南](usage.md)
- [生成器详情](generator.md)
- [运行时设计](runtime.md)
- [测试与基准](testing.md)
- [CI 配置](ci.md)

### English Documentation
- [Architecture Overview](../en/architecture.md)
- [Usage Guide](../en/usage.md)
- [Generator Guide](../en/generator.md)
- [Runtime Design](../en/runtime.md)
- [Testing & Benchmarks](../en/testing.md)
- [CI Guide](../en/ci.md)

### 日本語ドキュメント
- [アーキテクチャ概要](../ja/architecture.md)
- [使用ガイド](../ja/usage.md)
- [ジェネレータ説明](../ja/generator.md)
- [ランタイム設計](../ja/runtime.md)
- [テストとベンチマーク](../ja/testing.md)
- [CI 説明](../ja/ci.md)
