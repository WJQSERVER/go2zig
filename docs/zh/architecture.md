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

### 支持的 API 语法

- **基础类型**：`bool`、`u8-u64`、`i8-i64`、`f32`、`f64`
- **结构体**：`extern struct` 支持嵌套字段
- **枚举**：`enum(整数类型)` 支持显式值
- **数组**：固定长度 `[N]Type` 和命名别名
- **切片**：命名别名（如 `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`）
- **可选类型**：`?POD`（如 `?u32`、`?UserKind`）
- **错误处理**：`error{...}!ReturnType`
- **特殊类型**：`String`、`Bytes`、实验性的 `GoReader` / `GoWriter`

### 不支持的语法

- Go 特有：`map`、`chan`、`interface{}`、函数类型、指针
- Zig 特有：`union`、`comptime`、`@import`
- 有限支持：可选类型仅支持 POD，切片元素不能是 String/Bytes

## 2. Go 解析与模型层

相关位置：
- `internal/parser`
- `internal/model`
- `internal/names`

### 职责

- 解析 Zig 里的 `extern struct`、`extern fn` / `export fn`
- 把基础类型、`String`、`Bytes`、struct、error union 统一建模
- 处理 Go 命名风格，例如 `rename_user -> RenameUser`
- 验证类型支持和约束

### 类型模型

`internal/model/model.go` 定义了类型系统：
- `TypeKind` 枚举：`TypeVoid`、`TypePrimitive`、`TypeString`、`TypeBytes`、`TypeGoReader`、`TypeGoWriter`、`TypeStruct`、`TypeEnum`、`TypeOptional`、`TypeSlice`、`TypeArray`
- `PrimitiveInfo`：映射 Zig/Go/C 类型
- `TypeRef`：类型引用，支持嵌套和别名

### 解析器

`internal/parser/parser.go` 使用正则表达式解析 Zig 声明：
- `structPattern`：解析 `extern struct`
- `enumPattern`：解析 `enum(类型)`
- `slicePattern`：解析切片别名
- `arrayAliasPattern`：解析数组别名
- `funcPattern`：解析函数声明

## 3. 代码生成层

相关位置：
- `internal/generator`
- `go2zig.go`
- `cmd/go2zig`

### 职责

- 生成 Go 包装层 `gen.go`
- 生成 Zig 运行时辅助 `go2zig_runtime.zig`
- 生成 Zig 导出桥接 `go2zig_exports.zig`
- 调用 `zig build-lib -dynamic` 产出动态库

### 生成的文件

1. **`gen.go`**：
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

2. **`go2zig_runtime.zig`**：
   - 提供 `asSlice` / `asBytes`
   - 提供命名 slice alias 的 `asXxx` / `ownXxx`
   - 提供 `ownString` / `ownBytes`
   - 提供返回值内存释放辅助
   - 提供 `ErrorInfo`、`okError`、`makeError`
   - 提供 optional wrapper 的 `toOptional_xxx` / `fromOptional_xxx`

3. **`go2zig_exports.zig`**：
   - 为每个函数生成稳定的 `frame` ABI
   - 导出统一的 `go2zig_call_<name>` 符号
   - 把 Zig `error union` 降级为 `frame.err + frame.out`
   - 把 optional 的 wrapper 和 Zig 原生 optional 做桥接转换

### Builder API

`go2zig.go` 提供了 Builder 模式：
- `WithAPI(path)`：设置 API 文件路径
- `WithZigSource(path)`：设置 Zig 源文件路径
- `WithOutput(path)`：设置输出文件路径
- `WithPackageName(name)`：设置 Go 包名
- `WithLibraryName(name)`：设置库名
- `WithOptimize(mode)`：设置优化级别
- `WithHeaderOutput(path)`：输出 Zig 头文件
- `WithRuntimeZig(path)` / `WithBridgeZig(path)`：自定义生成文件位置
- `WithDynamicBuild(enabled)`：切换动态库 / 静态库构建
- `WithTopLevelFunctions(enabled)`：控制是否生成顶层转发函数
- `WithStreamExperimental(enabled)`：开启实验性流桥接
- `WithAPIModuleName(name)` / `WithImplModule(name)`：覆盖 Zig `@import` 模块名
- `Build()`：执行生成和构建

## 4. 运行时调用层

相关位置：
- `asmcall`
- `dynlib`

### 职责

- `dynlib` 负责按平台加载动态库和导出符号
- `asmcall` 负责在不经过 `cgo` 的情况下完成高频函数调用
- 生成的 Go 包装层只关心 frame struct 和业务签名，不需要用户手写 `unsafe`/ABI 细节

### `asmcall`

提供两类能力：
- `CallFuncG0P*`：切到 `g0` 栈执行目标函数
- `CallFuncP*`：直接在 goroutine 栈上执行目标函数

设计动机：
- 高频短调用时，`cgo` 的调度与栈切换成本通常过高
- 通过固定汇编胶水，可以把开销压到更可控的水平

### `dynlib`

在不同平台采用不同实现：
- **Windows**：基于系统 DLL 加载接口（`syscall.LoadDLL`）
- **Linux**：基于 `dlopen` / `dlsym` / `dlclose`（通过汇编调用）
- **Darwin**：基于 `dlopen` / `dlsym` / `dlclose`

### 错误协议

当前 `error union -> Go error` 使用固定协议：
- Zig frame 内带 `err: ErrorInfo`
- `ErrorInfo` 结构：
  - `code: u32`
  - `text: api.String`
- Go 侧统一生成：
  - 有返回值：`(T, error)`
  - 无返回值：`error`

### Optional 协议

当前支持 `optional POD`：
- `?primitive`
- `?enum`
- `?array` / `?array alias`

Go 侧公开类型默认映射为 `*T`，但 ABI 层不会直接依赖 Zig 原生 optional 布局，而是使用显式 tagged wrapper：
- Go ABI：`is_set + value`
- Zig runtime：`Optional_xxx`
- Zig bridge：`toOptional_xxx` / `fromOptional_xxx`

### Slice / Struct 生命周期

当切片元素本身还包含切片字段时，生成器会额外生成 `keep` 聚合结果，确保：
- 入参阶段临时 ABI backing buffer 在调用结束前不被回收
- 嵌套切片字段的 backing buffer 也能被一并保活

返回值侧则采用：
1. 逐元素 `own` 还原 Go 值
2. 释放元素内部动态字段
3. 最后释放外层返回缓冲区

## 当前支持的平台

- `windows/amd64`
- `windows/arm64`
- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

其中：
- 五个目标当前都在主 CI 中跑测试、benchmark 与构建校验
- Linux job 额外通过 `GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1` 开启底层 runtime 实跑测试

## 性能特点

### 优点
- 比 cgo 快约 8 倍（3.35ns vs 28.56ns）
- 无需 cgo 依赖
- 类型安全

### 缺点
- 每次调用需要数据复制
- 仅支持 Windows / Linux 上的 `amd64` / `arm64`，以及 Darwin 上的 `arm64`
- 类型支持有限

## 扩展点

### 短期扩展
- 支持 `?String` 和 `?Bytes` 可选类型
- 改进错误诊断

### 中期扩展
- `union` 类型支持
- 自定义分配器接口
- 性能优化

### 长期扩展
- 泛型支持
- 工具链集成
- 跨平台改进
