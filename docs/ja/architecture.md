# アーキテクチャ概要

`go2zig` は現在、4 層構造で構成されています。

## 1. Zig API 記述層

ユーザーは、制限付きの Zig 宣言を使って API を定義します。例:

```zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const User = extern struct {
    id: u64,
    name: String,
};

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
```

この層はエクスポート実装に直接参加するものではなく、ジェネレータへの入力として使われます。

### 対応している API 構文

- **基本型**: `bool`、`u8-u64`、`i8-i64`、`f32`、`f64`
- **struct**: ネストしたフィールドを含む `extern struct`
- **enum**: 明示的な値を持つ `enum(整数型)`
- **array**: 固定長 `[N]Type` と名前付き alias
- **slice**: 名前付き alias（例: `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`）
- **optional**: `?POD`（例: `?u32`、`?UserKind`）
- **エラー処理**: `error{...}!ReturnType`
- **特殊型**: `String`、`Bytes`

### 対応していない構文

- Go 固有: `map`、`chan`、`interface{}`、関数型、ポインタ
- Zig 固有: `union`、`comptime`、`@import`
- 制限付き対応: optional 型は POD のみ対応、slice 要素には String/Bytes を使えない

## 2. Go 解析・モデル層

関連箇所:
- `internal/parser`
- `internal/model`
- `internal/names`

### 役割

- Zig の `extern struct`、`extern fn` / `export fn` を解析する
- 基本型、`String`、`Bytes`、struct、error union を統一的にモデル化する
- Go の命名規則を処理する。例: `rename_user -> RenameUser`
- 型サポートと制約を検証する

### 型モデル

`internal/model/model.go` では型システムを定義しています。
- `TypeKind` enum: `TypeVoid`、`TypePrimitive`、`TypeString`、`TypeBytes`、`TypeStruct`、`TypeEnum`、`TypeOptional`、`TypeSlice`、`TypeArray`
- `PrimitiveInfo`: Zig/Go/C 型の対応を表す
- `TypeRef`: 型参照。ネストや alias をサポートする

### パーサ

`internal/parser/parser.go` は正規表現を使って Zig 宣言を解析します。
- `structPattern`: `extern struct` を解析
- `enumPattern`: `enum(type)` を解析
- `slicePattern`: slice alias を解析
- `arrayAliasPattern`: array alias を解析
- `funcPattern`: 関数宣言を解析

## 3. コード生成層

関連箇所:
- `internal/generator`
- `go2zig.go`
- `cmd/go2zig`

### 役割

- Go ラッパー層 `gen.go` を生成する
- Zig ランタイム補助 `go2zig_runtime.zig` を生成する
- Zig エクスポートブリッジ `go2zig_exports.zig` を生成する
- `zig build-lib -dynamic` を呼び出して動的ライブラリを生成する

### 生成されるファイル

1. **`gen.go`**:
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

2. **`go2zig_runtime.zig`**:
   - `asSlice` / `asBytes` を提供する
   - 名前付き slice alias 用の `asXxx` / `ownXxx` を提供する
   - `ownString` / `ownBytes` を提供する
   - 戻り値メモリの解放 helper を提供する
   - `ErrorInfo`、`okError`、`makeError` を提供する
   - optional wrapper 用の `toOptional_xxx` / `fromOptional_xxx` を提供する

3. **`go2zig_exports.zig`**:
   - 各関数ごとに安定した `frame` ABI を生成する
   - 統一された `go2zig_call_<name>` シンボルをエクスポートする
   - Zig の `error union` を `frame.err + frame.out` に降格する
   - optional wrapper と Zig ネイティブ optional の相互変換をブリッジする

### Builder API

`go2zig.go` は Builder パターンを提供します。
- `WithAPI(path)`: API ファイルパスを設定する
- `WithZigSource(path)`: Zig ソースファイルパスを設定する
- `WithOutput(path)`: 出力ファイルパスを設定する
- `WithPackageName(name)`: Go パッケージ名を設定する
- `WithLibraryName(name)`: ライブラリ名を設定する
- `WithOptimize(mode)`: 最適化レベルを設定する
- `Build()`: 生成とビルドを実行する

## 4. ランタイム呼び出し層

関連箇所:
- `asmcall`
- `dynlib`

### 役割

- `dynlib` はプラットフォームごとに動的ライブラリとエクスポートシンボルをロードする
- `asmcall` は `cgo` を経由せずに高頻度関数呼び出しを行う
- 生成された Go ラッパー層は frame struct と業務シグネチャだけを扱えばよく、ユーザーが `unsafe` / ABI 詳細を書く必要はない

### `asmcall`

`asmcall` は 2 種類の機能を提供します。
- `CallFuncG0P*`: `g0` スタックへ切り替えて対象関数を実行する
- `CallFuncP*`: goroutine スタック上で直接対象関数を実行する

設計上の動機:
- 高頻度の短い呼び出しでは、`cgo` のスケジューリングやスタック切り替えコストが高くなりやすい
- 固定のアセンブリ glue を使うことで、オーバーヘッドをより制御しやすい水準まで抑えられる

### `dynlib`

プラットフォームごとに異なる実装を使います。
- **Windows**: システムの DLL ロードインターフェース（`syscall.LoadDLL`）に基づく
- **Linux**: `dlopen` / `dlsym` / `dlclose` に基づく（アセンブリ呼び出し経由）

### エラープロトコル

現在の `error union -> Go error` は固定プロトコルを使います。
- Zig の frame は `err: ErrorInfo` を持つ
- `ErrorInfo` 構造体:
  - `code: u32`
  - `text: api.String`
- Go 側では一律に次の形を生成する
  - 戻り値あり: `(T, error)`
  - 戻り値なし: `error`

### Optional プロトコル

現在は `optional POD` をサポートしています。
- `?primitive`
- `?enum`
- `?array` / `?array alias`

Go 側の公開型はデフォルトで `*T` に対応しますが、ABI 層は Zig ネイティブ optional のレイアウトに直接依存せず、明示的な tagged wrapper を使います。
- Go ABI: `is_set + value`
- Zig runtime: `Optional_xxx`
- Zig bridge: `toOptional_xxx` / `fromOptional_xxx`

### Slice / Struct のライフサイクル

slice 要素自身が slice フィールドを含む場合、ジェネレータは追加で `keep` 集約結果を生成し、次を保証します。
- 入力時に作成された一時 ABI backing buffer が呼び出し完了前に回収されない
- ネストした slice フィールドの backing buffer もまとめて keepalive される

戻り値側では次の手順を取ります。
1. 各要素を `own` して Go 値へ復元する
2. 要素内部の動的フィールドを解放する
3. 最後に外側の戻り値バッファを解放する

## 現在対応しているプラットフォーム

- `windows/amd64`
- `windows/arm64`
- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

補足:
- Windows 経路はメイン CI で実行されています
- Linux 経路はメイン CI で生成、コンパイル、統合検証を行っていますが、低層 runtime の実行テストはデフォルトで無効であり、明示的に有効化する必要があります

## パフォーマンス特性

### 長所
- `cgo` よりおよそ 8 倍高速（3.35ns vs 28.56ns）
- `cgo` 依存が不要
- 型安全

### 短所
- 各呼び出しでデータコピーが必要
- Windows / Linux 上の `amd64` / `arm64`、および Darwin 上の `arm64` のみ対応
- 対応型が限定的

## 拡張ポイント

### 短期的な拡張
- `?String` と `?Bytes` optional 型のサポート
- エラー診断の改善

### 中期的な拡張
- `union` 型サポート
- カスタム allocator interface
- 性能最適化

### 長期的な拡張
- ジェネリクス対応
- ツールチェーン統合
- クロスプラットフォーム改善
