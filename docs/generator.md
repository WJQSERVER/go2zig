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
- `basic.dll`、`libbasic.so` 或 `libbasic.dylib`

如果通过 Go Builder 以编程方式调用 `WithDynamicBuild(false)`，则构建产物会改为静态库：

- Windows: `basic.lib`
- 非 Windows: `libbasic.a`

## `gen.go` 的职责

- 定义公开给 Go 业务使用的 struct 和函数
- 生成 Zig 枚举对应的 Go 命名类型和常量
- 生成命名数组别名对应的 Go 命名数组类型
- 生成 POD 切片别名对应的 Go 命名 slice 和 ABI helper
- 生成固定长度数组的 ABI 转换 helper
- 生成 optional tagged wrapper helper
- 生成 `Go2ZigClient`
- 默认生成 `Default` 实例和顶层转发函数
- 定义 ABI 结构和 frame 结构
- 负责字符串、字节串、struct、error 的转换

## `go2zig_runtime.zig` 的职责

- 提供 `asSlice` / `asBytes`
- 提供命名 slice alias 的 `asXxx` / `ownXxx`
- 提供 `ownString` / `ownBytes`
- 提供返回值内存释放辅助
- 提供 `ErrorInfo`、`okError`、`makeError`
- 提供 optional wrapper 的 `toOptional_xxx` / `fromOptional_xxx`

## `go2zig_exports.zig` 的职责

- 为每个函数生成稳定的 `frame` ABI
- 导出统一的 `go2zig_call_<name>` 符号
- 把 Zig `error union` 降级为 `frame.err + frame.out`
- 把 optional 的 wrapper 和 Zig 原生 optional 做桥接转换

## 当前生成器已经覆盖的类型层次

- primitive
- enum
- array / array alias
- slice alias
- struct
- `[]struct`
- `optional POD`
- `error union`

这里的 `slice alias` 实际能力比最早文档里的“POD 切片”更宽：

- 支持 primitive / enum / array / array alias
- 支持元素为 struct 的命名切片别名
- 支持 struct 字段中继续包含 POD slice 字段

当前仍然不支持把 `String` 或 `Bytes` 作为切片元素。

## Builder / Generate 真实产物行为

`Generate(...)` 只会：

- 必定写出 Go 文件
- 在 `RuntimeZig` 非空时写出 `go2zig_runtime.zig`
- 在 `BridgeZig` 非空时写出 `go2zig_exports.zig`

`Builder.Build()` 则会默认补齐 runtime/bridge 路径，因此通常会生成这三个文件；如果同时提供 `WithZigSource(...)`，还会额外：

- 写出 `go2zig_build_root.zig`
- 调用 `zig build-lib`

## 当前 Builder 常用方法

除了文档里常见的几项，当前公开 Builder 还包括：

- `WithHeaderOutput(path)`
- `WithRuntimeZig(path)`
- `WithBridgeZig(path)`
- `WithDynamicBuild(enabled)`
- `WithStreamExperimental(enabled)`
- `WithAPIModuleName(name)`
- `WithImplModule(name)`

其中 CLI 目前只暴露了 `-header`、`-runtime-zig`、`-bridge-zig`、`-stream-experimental`；静态库构建和自定义模块名仍主要走 Go Builder API。

## 一个当前实现限制

虽然 Builder 和 `GenerateConfig` 都允许自定义 `RuntimeZig` 输出路径，但 `go2zig_exports.zig` 内部当前仍固定使用：

```zig
const rt = @import("go2zig_runtime.zig");
```

因此最稳妥的用法仍然是让 runtime 文件与 bridge 文件保持默认相对位置和默认文件名。

## 为什么要用 frame

frame 的好处：

- 避免为每个函数做复杂 ABI 分支
- 便于统一处理错误返回
- 便于未来支持更多参数和返回值类型
- 生成器逻辑更稳定，调用方也更容易调试
