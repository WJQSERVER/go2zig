# CI Guide

The current CI runs as a 5-target matrix:

- `windows/amd64`
- `windows/arm64`
- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

## What Each Job Does

Every matrix entry does the following:

- Install the Go version declared in `go.mod`
- Install Zig `0.16.0`
- Regenerate `examples/basic` before testing
- Run `go test ./...`
- Run `go test -run ^$ -bench . ./...`
- Cross-check `go build ./...` for the matching `GOOS` / `GOARCH`

## Extra Linux Verification

The Linux `amd64` and `arm64` jobs explicitly set `GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1`, so main CI already covers:

- Linux runtime execution tests in `asmcall`
- Linux dynamic loading live tests in `dynlib`
- Benchmarks that exercise the Linux runtime path

To reproduce the same path locally, run:

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./...
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test -run ^$ -bench . ./...
```

## What CI Is Trying to Validate

This matrix is mainly checking that:

- The generator still emits consistent `gen.go`, `go2zig_runtime.zig`, and `go2zig_exports.zig`
- Build tags, dynamic library naming, and runtime loading remain correct on all supported targets
- The basic example, stream example, integration tests, and benchmarks still run end to end
- Changes do not break the no-`cgo` calling path

## Still Worth Improving

- Archive benchmark results to make regressions easier to spot
- Split smoke tests from full benchmarks if CI time grows further
- Upload documentation or example artifacts for pre-release inspection
