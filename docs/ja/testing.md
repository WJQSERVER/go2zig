# テストとベンチマーク

`go2zig` では現在、検証を 4 層に分けています。

## 1. Parser / Model のユニットテスト

関連ディレクトリ:
- `internal/parser`
- `internal/model`

主な検証内容:
- 型解析が正しいか
- 不正な API 宣言で適切にエラーを返せるか
- `POD` / `keepalive` / `free` の判定が正しいか

実行方法:

```bash
go test ./internal/parser ./internal/model
```

## 2. Generator のユニットテスト

関連ディレクトリ:
- `internal/generator`

主な検証内容:
- 生成された Go シグネチャが期待どおりか
- optional / slice / array alias / struct slice helper が生成されるか
- Zig runtime / bridge に主要な helper が存在するか

実行方法:

```bash
go test ./internal/generator
```

## 3. Integration / Example テスト

関連ファイル:
- `go2zig_test.go`
- `examples/basic/example_test.go`
- `examples/basic/edge_test.go`
- `examples/stream/stream_test.go`

主な検証内容:
- 実際の生成フローが最後まで動作するか
- Zig の動的ライブラリを正しくビルドできるか
- Go 側からさまざまな複雑な型を呼び出して正しい結果を得られるか
- 実験的ストリーム橋接が `io.Reader` / `io.Writer` / `io.Pipe` / `*os.File` を正しく扱えるか

実行方法:

```bash
go test ./...
```

## 4. ベンチマーク

### no-`cgo` / syscall 経路

関連ディレクトリ:
- `asmcall`

実行方法:

```bash
go test -run X -bench . ./asmcall
```

### CGo 比較ベンチマーク

関連ディレクトリ:
- `benchcmp`

Windows / PowerShell では次を使用できます。

```powershell
Set-Item -Path Env:CGO_ENABLED -Value 1
Set-Item -Path Env:CC -Value 'zig cc'
go test -run X -bench 'Benchmark(CgoAddU64|AsmCallCAddU64)$' ./benchcmp
```

開発マシンで直近に観測された代表的な結果:

- `BenchmarkCgoAddU64`: `28.56 ns/op`
- `BenchmarkAsmCallCAddU64`: `3.352 ns/op`

## Linux runtime の詳細テスト

Linux の低層 runtime 実行テストは、現在は Linux CI job で有効化されています。

ローカルで手動有効化する場合:

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

## 日常的に推奨される検証順序

parser / model を変更した場合:

```bash
go test ./internal/parser ./internal/model
```

generator / runtime を変更した場合:

```bash
go test ./internal/generator ./...
```

Linux の低層 runtime 経路に触れる変更なら、さらに次も推奨です。

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

性能関連を変更した場合:

```bash
go test -run X -bench . ./asmcall
```
