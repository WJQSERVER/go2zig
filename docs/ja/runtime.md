# ランタイム設計

## 全体目標

`cgo` を避け、Go から動的ライブラリのシンボルアドレスを通じて Zig のエクスポート関数を直接呼び出せるようにすることです。

## `asmcall`

`asmcall` は 2 種類の機能を提供します。

- `CallFuncG0P*`: `g0` スタックへ切り替えて対象関数を実行する
- `CallFuncP*`: goroutine スタック上で直接対象関数を実行する

設計上の動機:

- 高頻度の短い呼び出しでは、`cgo` のスケジューリングやスタック切り替えコストが高くなりやすい
- 固定のアセンブリ glue を使うことで、オーバーヘッドをより制御しやすい水準まで抑えられる

## `dynlib`

`dynlib` はプラットフォームごとに異なる実装を使います。

- Windows: システムの DLL ロードインターフェースに基づく
- Linux: `dlopen` / `dlsym` / `dlclose` に基づく

Linux 経路は現在、次に対応しています。

- メイン CI でのコンパイル検証
- 動的ライブラリの生成とロード

ただし、安定性の観点から、Linux の低層 runtime 実行テストはメイン CI ではデフォルトで有効化されていません。

Linux runtime の実行テストを手動で有効にするには:

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

## エラープロトコル

現在の `error union -> Go error` は固定プロトコルを使います。

- Zig の frame は `err: ErrorInfo` を持つ
- `ErrorInfo` 構造体:
  - `code: u32`
  - `text: api.String`
- Go 側では一律に次の形を生成する
  - 戻り値あり: `(T, error)`
  - 戻り値なし: `error`

これは Zig の error set をそのまま公開するより安定しています。理由は次のとおりです。

- ABI が固定される
- Go 側は標準の `error` だけを扱えばよい
- Go 側が Zig の enum 集合を理解する必要がない

## 現在追加されている型サポート

基本型、`String`、`Bytes`、struct に加えて、現在は次もサポートしています。

- 整数を基底型に持つ Zig enum。例: `enum(u8)`、`enum(u16)`
- POD slice alias。例: `extern struct { ptr: ?[*]const u16, len: usize }`
- 固定長 array。例: `[4]u8`、`[3]u16`、`[2]UserKind`

現在、POD slice でサポートされる要素型:
- 基本数値型
- 整数ベースの enum
- 固定長 array
- 名前付き POD slice alias

## Optional プロトコル

現在の第 1 段階では `optional POD` をサポートしています。

- `?primitive`
- `?enum`
- `?array` / `?array alias`

Go 側の公開型はデフォルトで `*T` にマッピングされますが、ABI 層は Zig ネイティブ optional のレイアウトに直接依存せず、明示的な tagged wrapper を使います。

- Go ABI: `is_set + value`
- Zig runtime: `Optional_xxx`
- Zig bridge: `toOptional_xxx` / `fromOptional_xxx`

この方式の利点:
- ABI がより安定する
- Go 側での表現がより自然になる
- 将来的により複雑な optional の組み合わせへ拡張しやすい

## Slice / Struct のライフサイクル

slice 要素自身が slice フィールドを含む場合、ジェネレータは追加で `keep` 集約結果を生成し、次を保証します。

- 入力時に作成された一時 ABI backing buffer が呼び出し完了前に回収されない
- ネストした slice フィールドの backing buffer もまとめて keepalive される

戻り値側では次の手順を使います。

1. 各要素を `own` して Go 値へ復元する
2. 要素内部の動的フィールドを解放する
3. 最後に外側の戻り値バッファを解放する

現在、array のブリッジは要素ごとの変換 helper を使っており、これによって次が可能になります。

- ABI ルールを明確に保てる
- 既存の要素単位の変換ロジックを再利用できる
- 今後より複雑な要素型をサポートする余地を残せる
