# CI 说明

当前 CI 是一个覆盖 5 个目标的矩阵：

- `windows/amd64`
- `windows/arm64`
- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

## 每个 job 做什么

每个矩阵条目都会：

- 安装 `go.mod` 指定的 Go 版本
- 安装 Zig `0.16.0`
- 先为 `examples/basic` 生成最新的 Go 包装与 Zig 桥接文件
- 运行 `go test ./...`
- 运行 `go test -run ^$ -bench . ./...`
- 再做一次对应 `GOOS` / `GOARCH` 的 `go build ./...` 交叉校验

## Linux 的额外验证

Linux 的 `amd64` 与 `arm64` job 会显式设置 `GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1`，因此当前主 CI 已经覆盖：

- `asmcall` Linux runtime 实跑测试
- `dynlib` Linux 动态加载实跑测试
- 依赖 Linux runtime 路径的 benchmark

如果要在本地复现同一路径，可以直接运行：

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./...
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test -run ^$ -bench . ./...
```

## 当前 CI 想回答的问题

这套矩阵主要验证：

- 生成器是否能稳定产出 `gen.go`、`go2zig_runtime.zig`、`go2zig_exports.zig`
- 五个当前支持目标上的 build tag、动态库命名和运行时装载路径是否一致
- 基础示例、流式示例、集成测试和 benchmark 是否还能跑通
- 修改是否破坏无 `cgo` 的调用链

## 后续仍然值得补强的方向

- 归档 benchmark 结果，便于观察回归
- 视 CI 时长把 smoke test 与 full benchmark 拆分
- 增加文档产物或示例产物上传，方便发布前检查
