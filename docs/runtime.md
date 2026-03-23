# 运行时设计

## 总体目标

避免 `cgo`，直接让 Go 通过动态库符号地址调用 Zig 导出函数。

## `asmcall`

`asmcall` 提供两类能力：

- `CallFuncG0P*`：切到 `g0` 栈执行目标函数
- `CallFuncP*`：直接在 goroutine 栈上执行目标函数

设计动机：

- 高频短调用时，`cgo` 的调度与栈切换成本通常过高
- 通过固定汇编胶水，可以把开销压到更可控的水平

## `dynlib`

`dynlib` 在不同平台采用不同实现：

- Windows：基于系统 DLL 加载接口
- Linux：基于 `dlopen` / `dlsym` / `dlclose`

Linux 路径当前已经支持：

- 主 CI 编译验证
- 动态库生成与加载

但出于稳定性考虑，Linux 底层 runtime 实跑测试默认不在主 CI 中启用。

如果要手动开启 Linux runtime 实跑测试：

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

## 错误协议

当前 `error union -> Go error` 使用固定协议：

- Zig frame 内带 `err: ErrorInfo`
- `ErrorInfo` 结构：
  - `code: u32`
  - `text: api.String`
- Go 侧统一生成：
  - 有返回值：`(T, error)`
  - 无返回值：`error`

这比直接暴露 Zig error set 更稳，因为：

- ABI 固定
- Go 侧只消费标准 `error`
- 不要求 Go 端理解 Zig 枚举集合

## 当前新增类型能力

除了基础类型、`String`、`Bytes`、struct 之外，当前还支持：

- 整型底层的 Zig 枚举，例如 `enum(u8)`、`enum(u16)`
- POD 切片别名，例如 `extern struct { ptr: ?[*]const u16, len: usize }`
- 固定长度数组，例如 `[4]u8`、`[3]u16`、`[2]UserKind`

当前 POD 切片支持的元素类型：

- 基础数值类型
- 整型底层枚举
- 固定长度数组
- 命名 POD 切片别名

## Optional 协议

当前第一阶段已经支持 `optional POD`：

- `?primitive`
- `?enum`
- `?array` / `?array alias`

Go 侧公开类型默认映射为 `*T`，但 ABI 层不会直接依赖 Zig 原生 optional 布局，而是使用显式 tagged wrapper：

- Go ABI：`is_set + value`
- Zig runtime：`Optional_xxx`
- Zig bridge：`toOptional_xxx` / `fromOptional_xxx`

这样做的好处是：

- ABI 更稳定
- Go 侧表达更自然
- 便于继续扩展到更复杂 optional 组合

## Slice / Struct 生命周期

当切片元素本身还包含切片字段时，生成器会额外生成 `keep` 聚合结果，确保：

- 入参阶段临时 ABI backing buffer 在调用结束前不被回收
- 嵌套切片字段的 backing buffer 也能被一并保活

返回值侧则采用：

1. 逐元素 `own` 还原 Go 值
2. 释放元素内部动态字段
3. 最后释放外层返回缓冲区

数组桥接当前走逐元素转换 helper，这样可以：

- 保持 ABI 规则明确
- 复用已有的元素级转换逻辑
- 为后续支持更复杂元素类型预留空间
