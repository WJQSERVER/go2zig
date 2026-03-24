# go2zig

语言: [English](README.md) | [简体中文](README_zh.md) | [日本語](README_ja.md)

一个轻量级、高性能的 Go 到 Zig FFI 代码生成器，灵感来自 [rust2go](https://github.com/ihciah/rust2go)。

## 核心特性

- **Zig 原生 API 声明** - 使用标准 Zig 语法作为 API 描述，无需额外 IDL
- **自动类型桥接** - 无缝转换 `string`、`[]byte`、嵌套结构体、枚举、切片和数组
- **无需 cgo** - 基于汇编的直接调用，性能比 cgo 提升约 8 倍
- **Builder 模式** - 一步生成 Go 包装并编译 Zig 动态库
- **双 API 风格** - 同时支持 `Client` 方法和顶层函数，使用更灵活

## 平台分级

参考 `purego` 的支持分级思路，`go2zig` 当前将平台支持划分为：

- **Tier 1** - CI 主验证目标：`windows/amd64`、`linux/amd64`
- **Tier 2** - 已支持交叉构建或新接入的平台：`windows/arm64`、`linux/arm64`、`darwin/arm64`

Tier 2 平台按 best-effort 方式支持，默认保证构建和生成包装可用；运行时边界行为仍可能需要额外的平台专项加固。

## 平台支持

### 支持的平台
- ✅ **Windows/amd64** - 完全支持，包含 CI 测试
- ✅ **Windows/arm64** - 无 cgo 汇编运行时已支持
- ✅ **Linux/amd64** - 完全支持，包含 CI 测试
- ✅ **Linux/arm64** - 无 cgo 汇编运行时已支持
- ✅ **Darwin/arm64** - 已支持动态加载与生成包装

### 不支持的平台
- ❌ **Darwin/amd64** - 当前不支持
- ❌ **其他架构** - 当前不支持

## 环境要求

- **Go** 1.26+
- **Zig** 0.15.2
- **平台**：Windows/Linux（支持 `amd64` 和 `arm64`）以及 Darwin（仅 `arm64`）

## 支持的类型

### 基础类型
- `bool`
- `u8`、`u16`、`u32`、`u64`、`usize`
- `i8`、`i16`、`i32`、`i64`、`isize`
- `f32`、`f64`

### 复合类型
- **结构体**：`extern struct` 支持嵌套字段
- **枚举**：`enum(整数类型)` 支持显式值
- **数组**：固定长度 `[N]Type` 和命名别名（如 `pub const Digest = [4]u8`）
- **切片**：命名别名（如 `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`）
- **可选类型**：`?POD`（如 `?u32`、`?UserKind`、`?Digest`）

### 特殊类型
- **String**：映射到 Go `string`（Zig 分配，Go 释放）
- **Bytes**：映射到 Go `[]byte`（Zig 分配，Go 释放）

### 错误处理
- **错误联合**：`error{...}!ReturnType` 映射到 Go `(T, error)` 或 `error`

## 不支持的类型

### Go 特有类型
- `map[K]V`
- `chan T`
- `interface{}`
- 函数类型（`func(...)`）
- 指针（String/Bytes 中的除外）
- `unsafe.Pointer`

### Zig 特有类型
- `union`
- `comptime`
- `@import`
- 复杂错误集

### 有限支持
- 可选类型仅支持 POD（Plain Old Data）
- 切片元素不能是 `String` 或 `Bytes`
- 嵌套可选（`??T`）不支持

## 快速开始

### 1. 在 Zig 中定义 API

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

### 2. 生成 Go 包装并构建

```bash
# 仅生成
go run ./cmd/go2zig -api ./api.zig -out ./gen.go -pkg main -lib mylib -no-build

# 生成并构建动态库
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib mylib

# 生成时禁用顶层转发函数
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib mylib -no-top-level
```

### 3. 在 Go 中使用

```go
package main

import "fmt"

func main() {
    // 加载动态库
    if err := Default.Load(); err != nil {
        panic(err)
    }

    // 直接调用函数
    if !Health() {
        panic("健康检查失败")
    }

    // 或者使用 client
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

## 性能表现

基于 Windows/amd64 的基准测试：

| 方法 | 性能 | 相对性能 |
|------|------|----------|
| **asmcall (go2zig)** | **3.35 ns/op** | **1x** |
| cgo | 28.56 ns/op | 慢约 8.5 倍 |

无 cgo 方法为短同步 FFI 调用提供约 **8 倍性能提升**。

## 内存管理

- **分配**：Zig 负责为字符串、字节和切片分配内存
- **释放**：Go 通过 `go2zig_free_buf` 释放内存
- **模式**：输入复制，输出复制
- **开销**：每次调用需要数据复制

## 生成的文件

运行生成器时会产生：

- `gen.go` - Go 包装层，包含类型和函数
- `go2zig_runtime.zig` - Zig 运行时辅助函数
- `go2zig_exports.zig` - Zig 导出桥接函数
- `mylib.dll` / `libmylib.so` - 动态库

## Builder API

在 Go 中编程使用：

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

如果你的项目已经手写了一层更高层的包装，并且希望避免生成 `Login(...)` 这类包级顶层转发函数，可以显式关闭：

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

这样仍会保留 `Go2ZigClient` 方法，例如 `client.Login(...)`，但不会生成顶层转发函数，从而降低与手写包装层发生重名冲突的概率。

## 示例

查看 `examples/basic/` 获取完整的工作示例，演示：

- 基础类型
- 结构体和嵌套结构体
- 枚举和显式值
- 数组和数组别名
- 切片和切片别名
- 可选类型
- 错误联合
- String 和 Bytes 处理

## 文档

- [架构概览](docs/architecture.md)
- [使用指南](docs/usage.md)
- [生成器详情](docs/generator.md)
- [运行时设计](docs/runtime.md)
- [测试与基准](docs/testing.md)
- [CI 配置](docs/ci.md)

## 限制

1. **平台**：仅支持 Windows/Linux 上的 `amd64` 与 `arm64`，以及 Darwin 上的 `arm64`
2. **类型**：不支持 Go 的 map、channel、interface
3. **内存**：固定分配模式（Zig 分配，Go 释放）
4. **性能**：每次调用需要数据复制
5. **错误处理**：仅支持简单错误集

## 未来路线图

### 短期
- 支持 `?String` 和 `?Bytes` 可选类型
- 更好的错误诊断

### 中期
- `union` 类型支持
- 自定义分配器接口
- 性能优化

### 长期
- 泛型支持
- 工具链集成
- 跨平台改进

## 贡献

1. 运行测试：`go test ./...`
2. 运行基准测试：`go test -bench . ./asmcall`
3. 在 Windows 和 Linux 上测试
4. 为新功能更新文档

## 许可证

本项目采用 [Mozilla Public License 2.0](LICENSE) 许可证。
