# 使用指南

这份文档按"从零到能跑"的顺序说明 `go2zig` 的典型使用方法。

## 1. 环境准备

### 平台要求

当前支持的平台：
- **Windows/amd64** - 完全支持
- **Windows/arm64** - 无 cgo 汇编运行时已支持
- **Linux/amd64** - 完全支持
- **Linux/arm64** - 无 cgo 汇编运行时已支持
- **Darwin/arm64** - 已支持动态加载与生成包装

不支持的平台：
- **Darwin/amd64** - 当前不支持
- 其他操作系统

### 软件要求

- Go `1.26`
- Zig `0.16.0`

## 2. 准备 API 描述文件

先写一个 Zig API 文件，例如 `api.zig`：

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

### 支持的类型

#### 完全支持
- **基础类型**：`bool`、`u8-u64`、`i8-i64`、`f32`、`f64`
- **结构体**：`extern struct` 支持嵌套字段
- **枚举**：`enum(整数类型)` 支持显式值（如 `enum(u8)`、`enum(u16)`）
- **数组**：固定长度 `[N]Type` 和命名别名（如 `pub const Digest = [4]u8`）
- **切片**：命名别名（如 `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`）
- **可选类型**：`?POD`（如 `?u32`、`?UserKind`、`?Digest`）
- **错误处理**：`error{...}!ReturnType`

#### 特殊类型
- **String**：映射到 Go `string`（Zig 分配，Go 释放）
- **Bytes**：映射到 Go `[]byte`（Zig 分配，Go 释放）

#### 不支持的类型
- Go 特有：`map`、`chan`、`interface{}`、函数类型、指针
- Zig 特有：`union`、`comptime`、`@import`
- 有限支持：可选类型仅支持 POD，切片元素不能是 String/Bytes

### 语法注意事项

- `String` 和 `Bytes` 是约定好的桥接类型别名
- 业务 struct 必须用 `extern struct`
- 函数声明用 `pub extern fn` 或 `pub export fn`
- `error union` 建议使用命名 error set，例如 `LoginError!LoginResponse`

## 3. 编写 Zig 业务实现

再写对应实现文件，例如 `lib.zig`：

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

### 关键函数

- `rt.asSlice` / `rt.asBytes`：把 Go 传入内容转成 Zig slice
- `rt.ownString` / `rt.ownBytes`：把返回值交给 Go 管理释放
- 不需要手写导出桥接函数，生成器会处理

## 4. 生成 Go 包装与 Zig 桥接文件

### 仅生成源码

```bash
go run ./cmd/go2zig -api ./api.zig -out ./gen.go -pkg main -lib basic -no-build
```

### 生成并构建动态库

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic
```

### 启用实验性流式支持

如果你的 Zig API 使用 `GoReader` / `GoWriter`，需要显式启用实验性流支持：

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic -stream-experimental
```

### 禁用顶层转发函数生成

如果你只想保留 `Go2ZigClient` 方法，而不希望生成 `Login(...)` 这类包级顶层转发函数，可以这样做：

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic -no-top-level
```

### 生成的文件

默认会产出：
- `gen.go` - Go 包装层
- `go2zig_runtime.zig` - Zig 运行时辅助
- `go2zig_exports.zig` - Zig 导出桥接
- `basic.dll`、`libbasic.so` 或 `libbasic.dylib` - 动态库

如果通过 Go Builder 以编程方式关闭动态构建（`WithDynamicBuild(false)`），则会改为产出静态库：Windows 下是 `.lib`，其他平台是 `.a`。

## 5. 在 Go 里调用

生成后，可以像普通 Go SDK 一样使用：

```go
package main

import "fmt"

func main() {
    // 加载动态库
    if err := Default.Load(); err != nil {
        panic(err)
    }

    // 直接调用顶层函数
    if !Health() {
        panic("Health check failed")
    }

    resp := Login(LoginRequest{
        User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
        Password: "secret-123",
    })

    // 或者使用 client
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

### 调用层有两种风格

- 顶层函数：`Login(...)`
- client 方法：`Default.Login(...)` 或 `NewGo2ZigClient(path)`

如果启用了 `-no-top-level`，则只会保留 client 方法，不再生成顶层函数。

### 类型映射

对于支持的类型：
- Zig `enum(u8)` 会生成 Go 命名类型和对应常量
- Zig 命名数组别名会生成 Go 命名数组类型
- 命名切片别名会生成 Go `[]T` 命名别名；当前不仅支持 POD 切片，也支持元素为 struct 的切片别名
- Zig `[N]T` 会生成 Go `[N]T` 数组，并自动做 ABI 转换
- Zig `?T` 当前会在 Go 侧生成 `*T`

### 实验性流类型

当前实验性流类型使用保留名：

- `GoReader`
- `GoWriter`

它们只能作为顶层函数参数使用，不能放入：

- 返回值
- `extern struct` 字段
- `optional`
- `slice`
- `array`

Zig API 需要显式声明：

```zig
pub const GoReader = usize;
pub const GoWriter = usize;

pub extern fn copy_stream(reader: GoReader, writer: GoWriter) u64;
```

在 Zig 实现中，通过 `go2zig_runtime.zig` 提供的 helper 使用：

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

Go 侧则通过辅助构造函数包装标准流对象：

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

当前支持包装：

- `io.Reader`
- `io.Writer`
- `io.ReadCloser`
- `io.WriteCloser`
- `io.Pipe`
- `*os.File`

当前限制：

- 这是实验性能力，必须显式开启
- 当前是同步分块流，不是异步或全双工协议
- Zig 侧实际接收到的是文件句柄风格的 `usize`

## 6. 自定义动态库路径

如果你不想使用默认同目录加载，可以手动指定：

```go
client := NewGo2ZigClient("./dist/libbasic.so")
if err := client.Load(); err != nil {
    panic(err)
}
```

建议在正式调用前显式执行一次 `Load()`。当前生成方法内部虽然也会懒加载动态库，但如果首次加载失败，调用路径会直接 `panic(err)`。

## 7. 错误返回怎么工作

对于 Zig `error union`，Go 侧会自动生成：
- 有 payload：`(T, error)`
- 无 payload：`error`

例如：

```zig
pub extern fn flush() FlushError!void;
```

会生成：

```go
func Flush() error
```

失败时你会拿到 `*Go2ZigError`：
- `Code`：Zig 错误码
- `Message`：当前默认是 Zig `@errorName(err)`

## 8. Builder 常用方法

如果你在 Go 代码里直接调用生成器，最常用的是：

- `WithAPI(path)`
- `WithZigSource(path)`
- `WithOutput(path)`
- `WithPackageName(name)`
- `WithLibraryName(name)`
- `WithOptimize(mode)`
- `WithTopLevelFunctions(enabled)`
- `Build()`

另外当前公开 API 还包括：

- `WithHeaderOutput(path)`
- `WithRuntimeZig(path)`
- `WithBridgeZig(path)`
- `WithDynamicBuild(enabled)`
- `WithStreamExperimental(enabled)`
- `WithAPIModuleName(name)`
- `WithImplModule(name)`

典型写法：

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

如果项目已经手写了一层更高层的包装，也可以在 Builder 中关闭顶层函数生成：

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

## 9. 性能考虑

当前实现的特点：
- **优点**：比 cgo 快约 8 倍（3.35ns vs 28.56ns）
- **缺点**：每次调用需要数据复制
- **适用**：高频短调用场景
- **不适用**：需要零拷贝或大数据传输的场景

## 10. 常见问题

### Q1: 为什么 Go 侧找不到动态库？

默认会从生成的 `gen.go` 所在目录旁边找：
- Windows：`basic.dll`
- Linux：`libbasic.so`
- Darwin：`libbasic.dylib`

如果路径不同，请用 `NewGo2ZigClient(customPath)`。

### Q2: 为什么 Linux 主 CI 不跑底层 runtime 实跑测试？

它现在已经在 Linux CI job 中运行了。主 CI 会通过 `GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1` 显式开启这组测试。

这里保留原问题标题，是为了兼容你可能按旧文档关键词来查找。

如果你需要本地开启 Linux runtime 深测：

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

### Q3: 什么时候只生成、不构建？

如果你只想先看 Go 包装和 Zig 桥接源码，用 `-no-build` 即可。

### Q4: 我应该先看哪里？

推荐顺序：
1. `README.md` 或 `README_zh.md`
2. `docs/architecture.md`
3. `docs/runtime.md`
4. `docs/testing.md`
5. `examples/basic`

### Q5: 为什么有些类型不支持？

当前设计限制：
- **平台限制**：仅支持 Windows/Linux 上的 `amd64` 与 `arm64`，以及 Darwin 上的 `arm64`
- **类型限制**：为了保持 ABI 稳定性和性能，不支持动态类型
- **内存管理**：固定分配模式，无法自定义

### Q6: 如何扩展支持更多类型？

需要修改：
1. `internal/model/model.go` - 添加新类型定义
2. `internal/parser/parser.go` - 添加解析逻辑
3. `internal/generator/generator.go` - 添加代码生成逻辑

参考现有类型的实现方式。

## 11. 调试技巧

### 启用详细日志

目前没有内置的详细日志，但你可以：
1. 检查生成的 `gen.go` 文件
2. 检查 `go2zig_runtime.zig` 和 `go2zig_exports.zig`
3. 使用 `go test -v` 查看测试输出

### 常见错误

1. **类型不支持**：检查是否使用了不支持的类型
2. **语法错误**：确保使用了正确的 Zig 语法
3. **平台不支持**：确保在 Windows/Linux 的 `amd64` 或 `arm64` 上运行，或在 Darwin 的 `arm64` 上运行

## 12. 最佳实践

1. **从简单开始**：先测试基础类型，再逐步添加复杂类型
2. **使用示例**：参考 `examples/basic/` 中的代码
3. **测试覆盖**：为所有 API 函数编写测试
4. **性能测试**：使用基准测试验证性能改进
5. **错误处理**：为所有可能失败的操作添加错误处理
