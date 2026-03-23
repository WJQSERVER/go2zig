# Testing & Benchmarks

`go2zig` currently splits verification into four layers:

## 1. Parser / Model Unit Tests

Related directories:

- `internal/parser`
- `internal/model`

Key verification:

- Whether type parsing is correct
- Whether invalid API declarations can report errors correctly
- Whether `POD` / `keepalive` / `free` judgments are correct

Run:

```bash
go test ./internal/parser ./internal/model
```

## 2. Generator Unit Tests

Related directories:

- `internal/generator`

Key verification:

- Whether generated Go signatures match expectations
- Whether optional / slice / array alias / struct slice helpers are generated
- Whether key helpers exist in Zig runtime / bridge

Run:

```bash
go test ./internal/generator
```

## 3. Integration / Example Tests

Related files:

- `go2zig_test.go`
- `examples/basic/example_test.go`

Key verification:

- Whether real generation flow can run through
- Whether Zig dynamic library can be built correctly
- Whether Go side can get correct results when calling various complex types

Run:

```bash
go test ./...
```

## 4. Benchmarks

### No cgo / syscall Path

Related directories:

- `asmcall`

Run:

```bash
go test -run X -bench . ./asmcall
```

### CGo Comparison Baseline

Related directories:

- `benchcmp`

Available on Windows / PowerShell:

```powershell
Set-Item -Path Env:CGO_ENABLED -Value 1
Set-Item -Path Env:CC -Value 'zig cc'
go test -run X -bench 'Benchmark(CgoAddU64|AsmCallCAddU64)$' ./benchcmp
```

A representative result from the most recent run on the development machine:

- `BenchmarkCgoAddU64`: `28.56 ns/op`
- `BenchmarkAsmCallCAddU64`: `3.352 ns/op`

## Linux Runtime Deep Testing

Linux bottom-level runtime live testing is not enabled by default in main CI.

Enable manually locally:

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

## Recommended Daily Verification Order

Change parser / model:

```bash
go test ./internal/parser ./internal/model
```

Change generator / runtime:

```bash
go test ./internal/generator ./...
```

Change performance-related:

```bash
go test -run X -bench . ./asmcall
```