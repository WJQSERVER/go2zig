# CI 说明

当前 CI 主要覆盖：

- Windows 主路径
- Linux 生成/编译路径

## CI 做什么

Windows job：

- 安装 Go 与 Zig
- 生成 `examples/basic`
- 运行 `go test ./...`
- 运行 `go test -bench . ./asmcall`
- 运行 PowerShell 交叉构建：`$env:GOOS='windows'; $env:GOARCH='arm64'; $env:CGO_ENABLED='0'; go build ./...`

Linux job：

- 安装 Go 与 Zig
- 生成 `examples/basic`
- 运行 `go test ./...`
- 运行 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...`
- 运行 `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build ./...`

## 为什么 Linux runtime 实跑测试默认关闭

当前 Linux 下的无 `cgo` runtime 仍处于性能导向的低层实现阶段。

为了让主 CI 更稳定，目前默认策略是：

- 主 CI 验证 Linux 路径可以生成、编译、集成构建
- 低层 runtime 的实跑测试通过环境变量手动开启

这样可以避免把所有 CI 结果绑定在当前仍在打磨的汇编调用细节上。

## 后续可以继续做的事

- 增加手动触发的 Linux runtime 深度验证 workflow
- 增加 benchmark 结果归档
- 等 action 生态稳定后，统一迁移到支持 Node 24 的版本
