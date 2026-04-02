# 测试与基准

`go2zig` 目前把验证拆成四层：

## 1. Parser / Model 单元测试

相关目录：

- `internal/parser`
- `internal/model`

重点验证：

- 类型解析是否正确
- 非法 API 声明是否能正确报错
- `POD` / `keepalive` / `free` 判定是否正确

运行：

```bash
go test ./internal/parser ./internal/model
```

## 2. Generator 单元测试

相关目录：

- `internal/generator`

重点验证：

- 生成的 Go 签名是否符合预期
- optional / slice / array alias / struct slice helper 是否被生成
- Zig runtime / bridge 中关键 helper 是否存在

运行：

```bash
go test ./internal/generator
```

## 3. Integration / Example 测试

相关文件：

- `go2zig_test.go`
- `examples/basic/example_test.go`
- `examples/basic/edge_test.go`
- `examples/stream/stream_test.go`

重点验证：

- 真实生成流程是否能跑通
- Zig 动态库是否能正确构建
- Go 侧调用各种复杂类型是否能得到正确结果
- 实验性流桥接是否能处理 `io.Reader` / `io.Writer` / `io.Pipe` / `*os.File`

运行：

```bash
go test ./...
```

## 4. 基准测试

### 无 cgo / syscall 路径

相关目录：

- `asmcall`

运行：

```bash
go test -run X -bench . ./asmcall
```

### CGo 对比基准

相关目录：

- `benchcmp`

Windows / PowerShell 下可用：

```powershell
Set-Item -Path Env:CGO_ENABLED -Value 1
Set-Item -Path Env:CC -Value 'zig cc'
go test -run X -bench 'Benchmark(CgoAddU64|AsmCallCAddU64)$' ./benchcmp
```

当前在开发机上最近一次观测到的一个代表性结果：

- `BenchmarkCgoAddU64`: `28.56 ns/op`
- `BenchmarkAsmCallCAddU64`: `3.352 ns/op`

## Linux runtime 深测

Linux 底层 runtime 实跑测试当前已经在 CI 的 Linux job 中开启。

本地手动开启：

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

## 建议的日常验证顺序

改 parser / model：

```bash
go test ./internal/parser ./internal/model
```

改 generator / runtime：

```bash
go test ./internal/generator ./...
```

如果改动涉及 Linux runtime 低层路径，建议额外跑：

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

改性能相关：

```bash
go test -run X -bench . ./asmcall
```
