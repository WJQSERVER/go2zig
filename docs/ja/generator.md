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
- `basic.dll`、`libbasic.so`、または `libbasic.dylib`

Go Builder から `WithDynamicBuild(false)` を使うと、ビルド成果物は静的ライブラリに切り替わります。

- Windows: `basic.lib`
- Windows 以外: `libbasic.a`

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

ここでいう `slice alias` は、初期ドキュメントの「POD slice」よりも広い範囲を指します。

- primitive / enum / array / array alias 要素
- 要素が struct の名前付き slice alias
- struct 内に POD slice フィールドを含むケース

ただし、`String` と `Bytes` を slice 要素にすることは引き続き未対応です。

## Builder / Generate の実際の出力挙動

`Generate(...)` が行うのは次です。

- Go ファイルは常に出力する
- `RuntimeZig` が非空なら `go2zig_runtime.zig` を出力する
- `BridgeZig` が非空なら `go2zig_exports.zig` を出力する

`Builder.Build()` は runtime / bridge のパスをデフォルトで補完するため、通常はこの 3 ファイルを生成します。さらに `WithZigSource(...)` を指定した場合は次も行います。

- `go2zig_build_root.zig` を出力する
- `zig build-lib` を実行する

## 現在の Builder メソッド

一般によく紹介されるメソッドに加えて、現在の公開 Builder には次もあります。

- `WithHeaderOutput(path)`
- `WithRuntimeZig(path)`
- `WithBridgeZig(path)`
- `WithDynamicBuild(enabled)`
- `WithStreamExperimental(enabled)`
- `WithAPIModuleName(name)`
- `WithImplModule(name)`

CLI が現在直接公開しているのは `-header`、`-runtime-zig`、`-bridge-zig`、`-stream-experimental` です。静的ライブラリ生成やモジュール名の上書きは主に Go Builder API を使います。

## 現在の実装上の制限

Builder と `GenerateConfig` は `RuntimeZig` の出力先を変更できますが、`go2zig_exports.zig` 側では現在も次を固定で使っています。

```zig
const rt = @import("go2zig_runtime.zig");
```

そのため、現時点で最も安全なのは runtime ファイルを bridge ファイルの近くにデフォルト名で置く構成です。

## なぜ frame を使うのか

frame を使う利点:

- 関数ごとに複雑な ABI 分岐を持たずに済む
- エラー戻り値を統一的に処理しやすい
- 将来さらに多くの引数型や戻り値型をサポートしやすい
- ジェネレータのロジックが安定し、呼び出し側でもデバッグしやすい
