# 架构概览

`go2zig` 当前由四层组成：

## 1. Zig API 描述层

用户用受限 Zig 声明定义 API，例如：

```zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const User = extern struct {
    id: u64,
    name: String,
};

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
```

这一层不直接参与导出实现，而是作为生成器输入。

## 2. Go 解析与模型层

相关位置：

- `internal/parser`
- `internal/model`
- `internal/names`

职责：

- 解析 Zig 里的 `extern struct`、`extern fn` / `export fn`
- 把基础类型、`String`、`Bytes`、struct、error union 统一建模
- 处理 Go 命名风格，例如 `rename_user -> RenameUser`

## 3. 代码生成层

相关位置：

- `internal/generator`
- `go2zig.go`
- `cmd/go2zig`

职责：

- 生成 Go 包装层 `gen.go`
- 生成 Zig 运行时辅助 `go2zig_runtime.zig`
- 生成 Zig 导出桥接 `go2zig_exports.zig`
- 调用 `zig build-lib -dynamic` 产出动态库

## 4. 运行时调用层

相关位置：

- `asmcall`
- `dynlib`

职责：

- `dynlib` 负责按平台加载动态库和导出符号
- `asmcall` 负责在不经过 `cgo` 的情况下完成高频函数调用
- 生成的 Go 包装层只关心 frame struct 和业务签名，不需要用户手写 `unsafe`/ABI 细节

## 当前支持的平台

- `windows/amd64`
- `linux/amd64`

其中：

- Windows 路径已经进入主 CI 实跑
- Linux 路径在主 CI 里做生成、编译与集成验证；底层 runtime 实跑测试默认关闭，需要显式开启
