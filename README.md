# go2zig

`go2zig` 仿照 `outrepo/rust2go` 的思路，做一个面向 Go -> Zig 的轻量代码生成器：

- 用 Zig 原生声明作为 API 描述，不额外引入 IDL
- 生成 Go 包装层，把 `string`、`[]byte`、嵌套 struct 自动转成 FFI 结构
- 提供 `Builder`，可在生成 Go 包装的同时直接编译 Zig 动态库
- Go 侧调用优先做得更顺手：默认生成 `Client` 方法和同名顶层函数，业务代码不用手写 `syscall`/`unsafe`
- 当前主线转向无 `cgo`：Go 通过 `asmcall + dynlib + 生成桥接层` 直接调用 Zig 导出函数

当前版本先聚焦单向调用：Go 调 Zig。

更多细节见 `docs/README.md`，如果你想按步骤落地，可以直接看 `docs/usage.md`。

## 当前进展

- 已完成 `windows/amd64` 与 `linux/amd64` 下无 `cgo` 的底层调用运行时：`asmcall` + 动态库符号加载
- 正在把代码生成器切换到新的无 `cgo` 桥接方案
- 目标是优先覆盖高频短调用场景，并结合 Zig 分配器特性降低返回值拷贝成本

## 基线环境

- Go: `1.26`
- Zig: `0.15.2`

## 支持的 API 语法

在 Zig 文件里使用一组受限声明：

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
};

pub extern fn health() bool;
pub extern fn rename_user(user: User, next_name: String) User;
```

目前支持：

- 基础类型：`bool` `u8/u16/u32/u64/usize` `i8/i16/i32/i64/isize` `f32` `f64`
- `enum(<int>)` 风格枚举
- 特殊类型：`String` `Bytes`
- `extern struct` 组合和嵌套
- POD 切片别名，例如 `extern struct { ptr: ?[*]const u16, len: usize }`
- 固定长度数组，例如 `[4]u8`、`[3]u16`、`[2]MyEnum`
- `pub extern fn` / `pub export fn` 风格函数签名解析
- `[*]const u8` 与 `?[*]const u8` 这类 Zig 指针形式的 `String`/`Bytes` 别名

## 生成方式

只生成 Go 包装和 Zig 桥接文件：

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -out ./examples/basic/gen.go -pkg main -lib basic -no-build
```

同时编译 Zig 动态库：

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -zig ./examples/basic/lib.zig -out ./examples/basic/gen.go -pkg main -lib basic
```

会在输出目录下生成：

- `gen.go`
- `go2zig_runtime.zig`
- `go2zig_exports.zig`
- `basic.dll` / `libbasic.so`（按目标平台生成）

如果你显式传入 `-header`，会额外尝试让 Zig 输出 C 头文件。

## Go 侧调用便利性改进

相较于 `rust2go` 的 `G2RCallImpl{}` 风格，这里默认直接生成两层 API：

- `Client` 方法，便于做依赖注入或未来扩展实例级配置
- 顶层函数，如 `Login(...)`、`RenameUser(...)`，默认转发到 `Default`，业务调用更短

示例：

```go
resp := Login(LoginRequest{
    User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
    Password: "secret-123",
})

renamed := RenameUser(respUser, "ally")
```

不需要显式创建 `Impl{}`，也不需要自己维护导出函数地址、frame struct 和内存释放协议。

## 运行示例

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -zig ./examples/basic/lib.zig -out ./examples/basic/gen.go -pkg main -lib basic
go run ./examples/basic
```

说明：当前无 `cgo` 高性能运行时优先支持 `windows/amd64` 与 `linux/amd64`。

## 后续可扩展方向

- `error union` / Zig error 到 Go `error` 的映射
- 切片和数组支持
- `go:generate` 辅助指令
- 更完整的 build helper，与 `go build`/`go generate` 深度集成
- 扩展到 `arm64`
- 进一步压缩 frame 布局和返回值分配开销
