# go2zig 文档

`go2zig` 当前聚焦在无 `cgo` 的 `Go -> Zig` 调用链，核心目标是：

- 在 Go 侧保留接近普通 SDK 的调用体验
- 在运行时侧尽量减少 `syscall` / `cgo` 带来的额外开销
- 用生成器把 ABI、frame、错误协议、字符串/字节串转换统一收敛

建议阅读顺序：

1. `docs/architecture.md`
2. `docs/usage.md`
3. `docs/generator.md`
4. `docs/runtime.md`
5. `docs/testing.md`
6. `docs/ci.md`

如果你只是想快速上手，可以先看仓库根目录的 `README.md`，再结合 `examples/basic` 理解完整生成流程。
