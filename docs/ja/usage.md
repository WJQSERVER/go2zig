# 使用ガイド

このドキュメントでは、`go2zig` の典型的な使い方を「ゼロから動作するまで」の順で説明します。

## 1. 環境準備

### プラットフォーム要件

現在サポートしているプラットフォーム:
- **Windows/amd64** - 完全対応
- **Windows/arm64** - no-cgo asm ランタイムで対応
- **Linux/amd64** - 完全対応
- **Linux/arm64** - no-cgo asm ランタイムで対応
- **Darwin/arm64** - 動的ロードと生成ラッパーに対応

サポートしていないプラットフォーム:
- **Darwin/amd64** - 現在は未サポート
- その他の OS

### ソフトウェア要件

- Go `1.26`
- Zig `0.15.2`

## 2. API 記述ファイルを用意する

まず、`api.zig` のような Zig API ファイルを作成します。

```zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Bytes = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const User = extern struct {
    id: u64,
    name: String,
    email: String,
};

pub const LoginRequest = extern struct {
    user: User,
    password: String,
};

pub const LoginResponse = extern struct {
    ok: bool,
    message: String,
    token: Bytes,
};

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn health() bool;
pub extern fn login(req: LoginRequest) LoginResponse;
pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
```

### 対応している型

#### 完全対応
- **基本型**: `bool`、`u8-u64`、`i8-i64`、`f32`、`f64`
- **struct**: ネストしたフィールドを含む `extern struct`
- **enum**: 明示的な値を持つ `enum(整数型)`（例: `enum(u8)`、`enum(u16)`）
- **array**: 固定長 `[N]Type` と名前付き alias（例: `pub const Digest = [4]u8`）
- **slice**: 名前付き alias（例: `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`）
- **optional**: `?POD`（例: `?u32`、`?UserKind`、`?Digest`）
- **エラー処理**: `error{...}!ReturnType`

#### 特殊型
- **String**: Go の `string` にマッピングされる（Zig が確保し、Go が解放）
- **Bytes**: Go の `[]byte` にマッピングされる（Zig が確保し、Go が解放）

#### 非対応の型
- Go 固有: `map`、`chan`、`interface{}`、関数型、ポインタ
- Zig 固有: `union`、`comptime`、`@import`
- 制限付き対応: optional 型は POD のみ対応、slice 要素には String/Bytes を使えない

### 構文上の注意

- `String` と `Bytes` は慣例的に決められたブリッジ用 type alias
- 業務用 struct は `extern struct` を使う必要がある
- 関数宣言には `pub extern fn` または `pub export fn` を使う
- `error union` では、`LoginError!LoginResponse` のように名前付き error set を使うことを推奨する

## 3. Zig の業務実装を書く

次に、対応する実装ファイル `lib.zig` を書きます。

```zig
const api = @import("api.zig");
const rt = @import("go2zig_runtime.zig");

pub fn health() bool {
    return true;
}

pub fn login_checked(req: api.LoginRequest) api.LoginError!api.LoginResponse {
    if (rt.asSlice(req.password).len < 6) return api.LoginError.InvalidPassword;
    return .{
        .ok = true,
        .message = rt.ownString("welcome alice"),
        .token = rt.ownBytes("token-123"),
    };
}
```

### 主要な関数

- `rt.asSlice` / `rt.asBytes`: Go から渡された入力を Zig の slice に変換する
- `rt.ownString` / `rt.ownBytes`: 戻り値のメモリ管理を Go 側へ引き渡す
- エクスポート用のブリッジ関数を手書きする必要はない。ジェネレータが処理する

## 4. Go ラッパーと Zig ブリッジファイルを生成する

### ソースコードのみ生成する場合

```bash
go run ./cmd/go2zig -api ./api.zig -out ./gen.go -pkg main -lib basic -no-build
```

### 生成して動的ライブラリもビルドする場合

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic
```

### 実験的なストリーム機能を有効にする

Zig API で `GoReader` / `GoWriter` を使う場合は、実験的ストリーム機能を明示的に有効にする必要があります。

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic -stream-experimental
```

### トップレベル転送関数を生成しない場合

`Go2ZigClient` のメソッドだけを残し、`Login(...)` のようなパッケージレベルの転送関数を生成したくない場合は、次のようにします。

```bash
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib basic -no-top-level
```

### 生成されるファイル

デフォルトでは次のファイルが生成されます。
- `gen.go` - Go ラッパー層
- `go2zig_runtime.zig` - Zig ランタイム補助
- `go2zig_exports.zig` - Zig エクスポートブリッジ
- `basic.dll`、`libbasic.so`、または `libbasic.dylib` - 動的ライブラリ

Go Builder から `WithDynamicBuild(false)` を使って動的ビルドを無効化した場合は、Windows では `.lib`、それ以外では `.a` の静的ライブラリが生成されます。

## 5. Go から呼び出す

生成後は、通常の Go SDK のように利用できます。

```go
package main

import "fmt"

func main() {
    // Load dynamic library
    if err := Default.Load(); err != nil {
        panic(err)
    }

    // Call top-level functions directly
    if !Health() {
        panic("Health check failed")
    }

    resp := Login(LoginRequest{
        User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
        Password: "secret-123",
    })

    // Or use the client
    client := NewGo2ZigClient("")
    if err := client.Load(); err != nil {
        panic(err)
    }

    checked, err := client.LoginChecked(LoginRequest{
        User: User{ID: 7, Name: "alice", Email: "alice@example.com"},
        Password: "secret-123",
    })
    if err != nil {
        panic(err)
    }

    _ = resp
    _ = checked
}
```

### 呼び出しスタイルは 2 種類ある

- トップレベル関数: `Login(...)`
- client メソッド: `Default.Login(...)` または `NewGo2ZigClient(path)`

`-no-top-level` を有効にした場合、生成されるのは client メソッドのみです。

### 型マッピング

対応している型について:
- Zig `enum(u8)` は Go の名前付き型と対応する定数を生成する
- Zig の名前付き array alias は Go の名前付き array 型を生成する
- 名前付き slice alias は Go の `[]T` 名前付き alias を生成する。現在は POD slice に加えて、要素が struct の slice alias も扱える
- Zig `[N]T` は Go `[N]T` array を生成し、ABI 変換を自動で行う
- Zig `?T` は現在 Go 側では `*T` として生成される

### 実験的なストリーム型

現在の実験的ストリーム橋接では、予約名として次を使います。

- `GoReader`
- `GoWriter`

これらはトップレベル関数の引数としてのみ使用でき、次の場所では使えません。

- 戻り値
- `extern struct` のフィールド
- `optional`
- `slice`
- `array`

Zig API では明示的に宣言してください。

```zig
pub const GoReader = usize;
pub const GoWriter = usize;

pub extern fn copy_stream(reader: GoReader, writer: GoWriter) u64;
```

Zig 側では `go2zig_runtime.zig` の helper を使います。

```zig
const rt = @import("go2zig_runtime.zig");

pub fn copy_stream(reader: usize, writer: usize) u64 {
    var total: u64 = 0;
    var buf: [32]u8 = undefined;
    while (true) {
        const n = rt.streamRead(reader, buf[0..]) catch |err| switch (err) {
            error.EndOfStream => break,
            else => @panic("stream read failed"),
        };
        const written = rt.streamWrite(writer, buf[0..n]) catch @panic("stream write failed");
        total += @as(u64, @intCast(written));
    }
    return total;
}
```

Go 側では、生成される helper を使って標準的なストリーム値を包みます。

```go
reader, err := NewGoReader(strings.NewReader("hello"))
if err != nil {
    panic(err)
}

var out bytes.Buffer
writer, err := NewGoWriter(&out)
if err != nil {
    panic(err)
}

copied := CopyStream(reader, writer)
_ = copied
```

現時点でラップ可能なのは次の型です。

- `io.Reader`
- `io.Writer`
- `io.ReadCloser`
- `io.WriteCloser`
- `io.Pipe`
- `*os.File`

現在の制限:

- 実験的機能のため、明示的に有効化する必要がある
- 現状は同期的なブロックストリームであり、async や full-duplex のプロトコルではない
- Zig 側には内部的にファイルハンドル相当の `usize` が渡される

## 6. 動的ライブラリのパスをカスタマイズする

生成ファイルと同じディレクトリからのデフォルトロードを使いたくない場合は、手動でパスを指定できます。

```go
client := NewGo2ZigClient("./dist/libbasic.so")
if err := client.Load(); err != nil {
    panic(err)
}
```

実運用では、最初に明示的に `Load()` を呼ぶのがおすすめです。生成メソッド側も内部で遅延ロードを試みますが、最初のロードに失敗した場合は現在の呼び出し経路が `panic(err)` します。

## 7. エラー戻り値の仕組み

Zig の `error union` に対して、Go 側では自動的に次の形を生成します。
- payload あり: `(T, error)`
- payload なし: `error`

例:

```zig
pub extern fn flush() FlushError!void;
```

生成される Go 側シグネチャ:

```go
func Flush() error
```

失敗時には `*Go2ZigError` が返されます。
- `Code`: Zig のエラーコード
- `Message`: 現在のデフォルトは Zig の `@errorName(err)`

## 8. よく使う Builder メソッド

Go コードからジェネレータを直接呼び出す場合、よく使うのは次のメソッドです。

- `WithAPI(path)`
- `WithZigSource(path)`
- `WithOutput(path)`
- `WithPackageName(name)`
- `WithLibraryName(name)`
- `WithOptimize(mode)`
- `WithTopLevelFunctions(enabled)`
- `Build()`

加えて、現在の公開 API には次もあります。

- `WithHeaderOutput(path)`
- `WithRuntimeZig(path)`
- `WithBridgeZig(path)`
- `WithDynamicBuild(enabled)`
- `WithStreamExperimental(enabled)`
- `WithAPIModuleName(name)`
- `WithImplModule(name)`

典型的な書き方:

```go
import "go2zig"

err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("basic").
    Build()
```

プロジェクト側ですでに独自のラッパー層がある場合は、Builder でもトップレベル関数生成を無効化できます。

```go
import "go2zig"

err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("basic").
    WithTopLevelFunctions(false).
    Build()
```

## 9. パフォーマンス上の考慮

現在の実装の特徴:
- **利点**: `cgo` より約 8 倍高速（3.35ns vs 28.56ns）
- **欠点**: 各呼び出しでデータコピーが必要
- **向いている用途**: 高頻度・短時間の呼び出し
- **向いていない用途**: zero-copy や大きなデータ転送が必要なケース

## 10. よくある質問

### Q1: Go 側で動的ライブラリが見つからないのはなぜですか？

デフォルトでは、生成された `gen.go` の隣にある次のファイルを探します。
- Windows: `basic.dll`
- Linux: `libbasic.so`
- Darwin: `libbasic.dylib`

パスが異なる場合は `NewGo2ZigClient(customPath)` を使ってください。

### Q2: なぜ Linux のメイン CI では低層 runtime の実行テストを行わないのですか？

現在は Linux CI job で実行されています。メイン CI は `GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1` を使ってこの経路を明示的に有効化しています。

古い説明を探している場合に見つけやすいよう、見出しだけはそのまま残しています。

ローカルで Linux runtime の詳細テストを有効にする場合:

```bash
GO2ZIG_RUN_LINUX_RUNTIME_TESTS=1 go test ./asmcall ./dynlib
```

### Q3: 生成だけ行い、ビルドしないのはどんな場合ですか？

まず Go ラッパーや Zig ブリッジのソースを確認したい場合は、`-no-build` を使います。

### Q4: どこから読むのがおすすめですか？

推奨順:
1. `README.md` または `README_zh.md`
2. `docs/architecture.md`
3. `docs/runtime.md`
4. `docs/testing.md`
5. `examples/basic`

### Q5: 一部の型がサポートされないのはなぜですか？

現在の設計上の制約:
- **プラットフォーム制約**: Windows/Linux 上の `amd64` と `arm64`、および Darwin 上の `arm64` のみ対応
- **型制約**: ABI の安定性と性能維持のため、動的型はサポートしない
- **メモリ管理**: 固定の割り当て方式であり、カスタマイズできない

### Q6: さらに多くの型をサポートするにはどうすればよいですか？

次の箇所を変更する必要があります。
1. `internal/model/model.go` - 新しい型定義を追加
2. `internal/parser/parser.go` - 解析ロジックを追加
3. `internal/generator/generator.go` - コード生成ロジックを追加

既存の型実装を参考にしてください。

## 11. デバッグのコツ

### 詳細ログを有効にする

現時点では組み込みの詳細ログはありませんが、次を確認できます。
1. 生成された `gen.go` を確認する
2. `go2zig_runtime.zig` と `go2zig_exports.zig` を確認する
3. `go test -v` でテスト出力を見る

### よくあるエラー

1. **型がサポートされていない**: 非対応の型を使っていないか確認する
2. **構文エラー**: 正しい Zig 構文を使っているか確認する
3. **プラットフォーム未対応**: Windows/Linux の `amd64` または `arm64`、または Darwin の `arm64` 上で実行しているか確認する

## 12. ベストプラクティス

1. **単純なものから始める**: まず基本型を試し、その後に複雑な型を追加する
2. **サンプルを活用する**: `examples/basic/` のコードを参照する
3. **テストを充実させる**: すべての API 関数にテストを書く
4. **性能を測定する**: ベンチマークで性能改善を確認する
5. **エラー処理を入れる**: 失敗しうる操作には必ずエラー処理を加える
