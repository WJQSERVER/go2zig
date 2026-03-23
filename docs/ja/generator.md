# ジェネレータ説明

`go2zig` のジェネレータは、汎用的な Zig バインディングシステムを目指すものではなく、現在の no-`cgo` ランタイムを中心に固定プロトコルを生成するためのものです。

## 生成されるファイル

実行例:

```bash
go run ./cmd/go2zig -api ./examples/basic/api.zig -zig ./examples/basic/lib.zig -out ./examples/basic/gen.go -pkg main -lib basic
```

通常は次のファイルが生成されます。
- `gen.go`
- `go2zig_runtime.zig`
- `go2zig_exports.zig`
- `basic.dll` または `libbasic.so`

## `gen.go` の役割

- Go のビジネスロジックから利用する struct と関数を定義する
- Zig enum に対応する Go の名前付き型と定数を生成する
- 名前付き array alias に対応する Go の名前付き array 型を生成する
- POD slice alias に対応する Go の名前付き slice と ABI helper を生成する
- 固定長 array の ABI 変換 helper を生成する
- optional tagged wrapper helper を生成する
- `Go2ZigClient` を生成する
- デフォルトで `Default` インスタンスとトップレベル転送関数を生成する
- ABI struct と frame struct を定義する
- 文字列、バイト列、struct、error の変換を担当する

## `go2zig_runtime.zig` の役割

- `asSlice` / `asBytes` を提供する
- 名前付き slice alias 用の `asXxx` / `ownXxx` を提供する
- `ownString` / `ownBytes` を提供する
- 戻り値メモリの解放 helper を提供する
- `ErrorInfo`、`okError`、`makeError` を提供する
- optional wrapper 用の `toOptional_xxx` / `fromOptional_xxx` を提供する

## `go2zig_exports.zig` の役割

- 各関数ごとに安定した `frame` ABI を生成する
- 統一された `go2zig_call_<name>` シンボルをエクスポートする
- Zig の `error union` を `frame.err + frame.out` に降格する
- optional wrapper と Zig ネイティブ optional の相互変換をブリッジする

## 現在ジェネレータがカバーしている型階層

- primitive
- enum
- array / array alias
- slice alias
- struct
- `[]struct`
- `optional POD`
- `error union`

## なぜ frame を使うのか

frame を使う利点:

- 関数ごとに複雑な ABI 分岐を持たずに済む
- エラー戻り値を統一的に処理しやすい
- 将来さらに多くの引数型や戻り値型をサポートしやすい
- ジェネレータのロジックが安定し、呼び出し側でもデバッグしやすい
