# 使用指南

这份文档按“从零到能跑”的顺序说明 `go2zig` 的典型使用方法。

## 1. 环境准备

当前建议环境：

- Go `1.26`
- Zig `0.15.2`

当前主线运行时重点支持：

- `windows/amd64`
- `linux/amd64`

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

建议注意：

- `String` 和 `Bytes` 是约定好的桥接类型别名
- 可以声明 `enum(u8)`、`enum(u16)`、`enum(u32)` 等整型枚举
- 可以声明命名数组别名，例如 `pub const Digest = [4]u8`
- 可以声明 POD 切片别名，例如 `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`
- 可以声明固定长度数组，例如 `[4]u8`、`[3]u16`、`[2]UserKind`
- 可以声明 `optional POD`，例如 `?u32`、`?UserKind`、`?Digest`
- 业务 struct 用 `extern struct`
- 函数声明用 `pub extern fn`
- `error union` 当前建议优先使用命名 error set，例如 `LoginError!LoginResponse`

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

关键点：

- `rt.asSlice` / `rt.asBytes` 用于把 Go 传入内容转成 Zig slice
- `rt.ownString` / `rt.ownBytes` 用于把返回值交给 Go 管理释放
- 不需要手写导出桥接函数，生成器会处理

## 4. 生成 Go 包装与 Zig 桥接文件

只生成源码：

```bash
go run ./cmd/go2zig -api ./api.zig -out ./gen.go -pkg main -lib basic -no-build
```

生成并构建动态库：

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic
```

默认会产出：

- `gen.go`
- `go2zig_runtime.zig`
- `go2zig_exports.zig`
- `basic.dll` 或 `libbasic.so`

## 5. 在 Go 里调用

生成后，可以像普通 Go SDK 一样使用：

```go
if err := Default.Load(); err != nil {
    panic(err)
}

resp := Login(LoginRequest{
    User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
    Password: "secret-123",
})

checked, err := LoginChecked(LoginRequest{
    User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
    Password: "secret-123",
})
if err != nil {
    panic(err)
}

_ = checked
_ = resp
```

调用层有两种风格：

- 顶层函数：`Login(...)`
- client 方法：`Default.Login(...)` 或 `NewGo2ZigClient(path)`

对于新增支持的类型：

- Zig `enum(u8)` 会生成 Go 命名类型和对应常量
- Zig 命名数组别名会生成 Go 命名数组类型
- POD 切片别名会生成 Go `[]T` 命名别名，并自动做零拷贝入参 / 拷贝出参转换
- POD 切片的元素当前可以是基础类型、整型枚举、固定长度数组
- Zig `[N]T` 会生成 Go `[N]T` 数组，并自动做 ABI 转换
- Zig `?T` 当前会在 Go 侧生成 `*T`

例如：

```zig
pub const Digest = [4]u8;
pub extern fn maybe_digest(flag: bool) ?Digest;
```

会生成：

```go
type Digest [4]uint8

func MaybeDigest(flag bool) *Digest
```

## 6. 自定义动态库路径

如果你不想使用默认同目录加载，可以手动指定：

```go
client := NewGo2ZigClient("./dist/libbasic.so")
if err := client.Load(); err != nil {
    panic(err)
}
```

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
- `Build()`

典型写法：

```go
err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("basic").
    Build()
```

## 9. 常见问题

### Q1: 为什么 Go 侧找不到动态库？

默认会从生成的 `gen.go` 所在目录旁边找：

- Windows：`basic.dll`
- Linux：`libbasic.so`

如果路径不同，请用 `NewGo2ZigClient(customPath)`。

### Q2: 为什么 Linux 主 CI 不跑底层 runtime 实跑测试？

因为 Linux 下这条无 `cgo` runtime 仍在持续打磨，当前主 CI 以稳定的生成、编译和集成验证为主。

如果你需要本地开启 Linux runtime 深测：

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

### Q3: 什么时候只生成、不构建？

如果你只想先看 Go 包装和 Zig 桥接源码，用 `-no-build` 即可。

### Q4: 我应该先看哪里？

推荐顺序：

1. `README.md`
2. `docs/architecture.md`
3. `docs/runtime.md`
4. `docs/testing.md`
5. `examples/basic`
