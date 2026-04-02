# CI 説明

現在の CI は 5 ターゲットのマトリクスで動作しています。

- `windows/amd64`
- `windows/arm64`
- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

## 各 job が行うこと

各マトリクス項目では次を実行します。

- `go.mod` で宣言された Go バージョンを導入する
- Zig `0.15.2` を導入する
- テスト前に `examples/basic` を再生成する
- `go test ./...` を実行する
- `go test -run ^$ -bench . ./...` を実行する
- 対応する `GOOS` / `GOARCH` で `go build ./...` を再確認する

## Linux の追加検証

Linux の `amd64` / `arm64` job では `GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1` を明示的に設定しています。つまり現在のメイン CI ではすでに次も検証しています。

- `asmcall` の Linux runtime 実行テスト
- `dynlib` の Linux 動的ロード実行テスト
- Linux runtime 経路を使うベンチマーク

ローカルで同じ経路を再現するには次を使えます。

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./...
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test -run ^$ -bench . ./...
```

## CI が主に確認していること

このマトリクスは主に次を確認しています。

- `gen.go`、`go2zig_runtime.zig`、`go2zig_exports.zig` が安定して生成できるか
- build tag、動的ライブラリ名、ランタイムロード経路が全サポートターゲットで整合しているか
- basic / stream / integration / benchmark が端から端まで壊れていないか
- no-`cgo` 呼び出し経路を変更が壊していないか

## 今後さらに改善したい点

- benchmark 結果を保存して回帰を見つけやすくする
- CI 時間が増えたら smoke test と full benchmark を分離する
- リリース前確認用にドキュメントや example 生成物を artifact 化する
