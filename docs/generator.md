# 生成器说明

`go2zig` 的生成器目标不是做一个通用 Zig 绑定系统，而是围绕当前无 `cgo` 运行时，生成一套固定协议。

## 生成的文件

调用：

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -zig ./examples/basic/lib.zig -out ./examples/basic/gen.go -pkg main -lib basic
```

通常会生成：

- `gen.go`
- `go2zig_runtime.zig`
- `go2zig_exports.zig`
- `basic.dll` 或 `libbasic.so`

## `gen.go` 的职责

- 定义公开给 Go 业务使用的 struct 和函数
- 生成 Zig 枚举对应的 Go 命名类型和常量
- 生成 POD 切片别名对应的 Go 命名 slice 和 ABI helper
- 生成固定长度数组的 ABI 转换 helper
- 生成 `Go2ZigClient`
- 默认生成 `Default` 实例和顶层转发函数
- 定义 ABI 结构和 frame 结构
- 负责字符串、字节串、struct、error 的转换

## `go2zig_runtime.zig` 的职责

- 提供 `asSlice` / `asBytes`
- 提供 `ownString` / `ownBytes`
- 提供返回值内存释放辅助
- 提供 `ErrorInfo`、`okError`、`makeError`

## `go2zig_exports.zig` 的职责

- 为每个函数生成稳定的 `frame` ABI
- 导出统一的 `go2zig_call_<name>` 符号
- 把 Zig `error union` 降级为 `frame.err + frame.out`

## 为什么要用 frame

frame 的好处：

- 避免为每个函数做复杂 ABI 分支
- 便于统一处理错误返回
- 便于未来支持更多参数和返回值类型
- 生成器逻辑更稳定，调用方也更容易调试
