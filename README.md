# go2zig

`go2zig` 仿照 `outrepo/rust2go` 的思路，做一个面向 Go -> Zig 的轻量代码生成器：

- 用 Zig 原生声明作为 API 描述，不额外引入 IDL
- 生成 Go 包装层，把 `string`、`[]byte`、嵌套 struct 自动转成 FFI 结构
- 提供 `Builder`，可在生成 Go 包装的同时直接编译 Zig 静态库
- Go 侧调用优先做得更顺手：默认生成 `Client` 方法和同名顶层函数，业务代码不用手写 `C.xxx`

当前版本先聚焦单向调用：Go 调 Zig。

## 当前进展

- 已完成 `cgo` 版本代码生成与端到端验证
- 已新增无 `cgo` 性能路径的底层基础设施：`asmcall` + 动态库符号加载（当前优先 `windows/amd64`）
- 下一步会把生成器切到无 `cgo` 运行时，直接面向 `Go -> Zig` 高频调用场景

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
- 特殊类型：`String` `Bytes`
- `extern struct` 组合和嵌套
- `pub extern fn` / `pub export fn` 风格函数签名解析

## 生成方式

只生成 Go 包装：

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -out ./examples/basic/gen.go -pkg main -lib basic -no-build
```

同时编译 Zig 静态库：

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -zig ./examples/basic/lib.zig -out ./examples/basic/gen.go -pkg main -lib basic
```

会在输出目录下生成：

- `gen.go`
- `libbasic.a`

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

不需要显式创建 `Impl{}`，也不需要自己维护 `C` 层结构转换。

## 运行示例

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -zig ./examples/basic/lib.zig -out ./examples/basic/gen.go -pkg main -lib basic
go run ./examples/basic
```

说明：示例依赖 `cgo`。如果当前环境默认关闭 `cgo`，需要显式启用，并提供可用的 C 编译器，例如：

```bash
CGO_ENABLED=1 CC="zig cc" go run ./examples/basic
```

## 后续可扩展方向

- `error union` / Zig error 到 Go `error` 的映射
- 切片和数组支持
- `go:generate` 辅助指令
- 更完整的 build helper，与 `go build`/`go generate` 深度集成
- 基于 `asmcall` 的无 `cgo` 高频调用后端
