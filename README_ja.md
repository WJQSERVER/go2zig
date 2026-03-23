# go2zig

Languages: [English](README.md) | [简体中文](README_zh.md) | [日本語](README_ja.md)

[rust2go](https://github.com/ihciah/rust2go) に着想を得た、軽量で高性能な Go-to-Zig FFI コードジェネレータです。

## Key Features

- **Zig ネイティブな API 宣言** - API の記述に標準の Zig 構文をそのまま使えるため、別途 IDL は不要です
- **自動型ブリッジ** - `string`、`[]byte`、ネストした struct、enum、slice、array をシームレスに変換します
- **cgo 不要** - アセンブリベースの直接呼び出しにより、cgo と比べて最大 8 倍の性能向上を実現します
- **Builder パターン** - Go ラッパーの生成と Zig 動的ライブラリのビルドを 1 ステップで行えます
- **デュアル API スタイル** - `Client` メソッドとトップレベル関数の両方を提供し、柔軟に利用できます

## Platform Support

### Supported Platforms
- ✅ **Windows/amd64** - CI テストを含む完全サポート
- ✅ **Windows/arm64** - no-cgo asm ランタイムでサポート
- ✅ **Linux/amd64** - CI テストを含む完全サポート
- ✅ **Linux/arm64** - no-cgo asm ランタイムでサポート
- ✅ **Darwin/arm64** - 動的ロードと生成ラッパーに対応

### Unsupported Platforms
- ❌ **Darwin/amd64** - 現在は未サポート
- ❌ **Other architectures** - 現在は未サポート

## Requirements

- **Go** 1.26+
- **Zig** 0.15.2
- **Platform**: Windows/Linux（`amd64` と `arm64`）および Darwin（`arm64` のみ）

## Supported Types

### Primitive Types
- `bool`
- `u8`, `u16`, `u32`, `u64`, `usize`
- `i8`, `i16`, `i32`, `i64`, `isize`
- `f32`, `f64`

### Composite Types
- **Structs**: ネストしたフィールドを含む `extern struct`
- **Enums**: 明示的な値を持つ `enum(integer_type)`
- **Arrays**: 固定長 `[N]Type` および `pub const Digest = [4]u8` のような名前付きエイリアス
- **Slices**: `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }` のような名前付きエイリアス
- **Optionals**: `?POD`（例: `?u32`, `?UserKind`, `?Digest`）

### Special Types
- **String**: Go の `string` にマッピングされます（Zig が確保し、Go が解放）
- **Bytes**: Go の `[]byte` にマッピングされます（Zig が確保し、Go が解放）

### Error Handling
- **Error unions**: `error{...}!ReturnType` は Go の `(T, error)` または `error` にマッピングされます

## Unsupported Types

### Go-specific Types
- `map[K]V`
- `chan T`
- `interface{}`
- 関数型（`func(...)`）
- ポインタ（String/Bytes 内のものを除く）
- `unsafe.Pointer`

### Zig-specific Types
- `union`
- `comptime`
- `@import`
- 複雑な error set

### Limited Support
- Optional 型は POD（Plain Old Data）のみサポートされます
- Slice の要素に `String` または `Bytes` は使用できません
- ネストした Optional（`??T`）はサポートされません

## Quick Start

### 1. Define API in Zig

```zig
// api.zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Bytes = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const UserKind = enum(u8) {
    guest,
    member,
    admin,
};

pub const User = extern struct {
    id: u64,
    kind: UserKind,
    name: String,
    email: String,
};

pub extern fn health() bool;
pub extern fn login(user: User, password: String) String;
```

### 2. Generate Go Wrapper and Build

```bash
# Generate only
go run ./cmd/go2zig -api ./api.zig -out ./gen.go -pkg main -lib mylib -no-build

# Generate and build dynamic library
go run ./cmd/go2zig -api ./api.zig -zig ./lib.zig -out ./gen.go -pkg main -lib mylib
```

### 3. Use in Go

```go
package main

import "fmt"

func main() {
    // Load the dynamic library
    if err := Default.Load(); err != nil {
        panic(err)
    }

    // Call functions directly
    if !Health() {
        panic("Health check failed")
    }

    // Or use the client
    client := NewGo2ZigClient("")
    if err := client.Load(); err != nil {
        panic(err)
    }

    greeting := client.Login(User{
        ID:     1,
        Kind:   UserKindMember,
        Name:   "alice",
        Email:  "alice@example.com",
    }, "password123")
    fmt.Println(greeting)
}
```

## Performance

Windows/amd64 でのベンチマーク結果に基づいています。

| Method | Performance | Relative |
|--------|-------------|----------|
| **asmcall (go2zig)** | **3.35 ns/op** | **1x** |
| cgo | 28.56 ns/op | 約 8.5 倍遅い |

この no-cgo アプローチは、短く同期的な FFI 呼び出しにおいて約 **8 倍の性能向上** をもたらします。

## Memory Management

- **Allocation**: string、bytes、slice のメモリは Zig 側で確保されます
- **Deallocation**: Go 側で `go2zig_free_buf` を通じて解放します
- **Pattern**: 入力は copy-in、出力は copy-out です
- **Overhead**: 各呼び出しでデータコピーが必要です

## Generated Files

ジェネレータを実行すると、以下のファイルが生成されます。

- `gen.go` - 型と関数を含む Go ラッパー
- `go2zig_runtime.zig` - Zig ランタイム補助コード
- `go2zig_exports.zig` - Zig エクスポート用ブリッジ関数
- `mylib.dll` / `libmylib.so` - 動的ライブラリ

## Builder API

Go からプログラムとして利用する場合:

```go
import "go2zig"

err := go2zig.NewBuilder().
    WithAPI("./api.zig").
    WithZigSource("./lib.zig").
    WithOutput("./gen.go").
    WithPackageName("main").
    WithLibraryName("mylib").
    Build()
```

## Examples

完全な動作例については `examples/basic/` を参照してください。以下の内容を実演しています。

- Primitive types
- Structs and nested structs
- Enums with explicit values
- Arrays and array aliases
- Slices and slice aliases
- Optionals
- Error unions
- String and Bytes handling

## Documentation

- [Docs Home](docs/ja/README.md)
- [Architecture Overview](docs/ja/architecture.md)
- [Usage Guide](docs/ja/usage.md)
- [Generator Guide](docs/ja/generator.md)
- [Runtime Design](docs/ja/runtime.md)
- [Testing & Benchmarks](docs/ja/testing.md)
- [CI Guide](docs/ja/ci.md)

## Limitations

1. **Platform**: 対応するのは Windows/Linux 上の `amd64` と `arm64`、および Darwin 上の `arm64` のみ
2. **Types**: Go の maps、channels、interfaces は未サポート
3. **Memory**: 固定のメモリ管理方式（Zig が確保し、Go が解放）
4. **Performance**: 各呼び出しでデータコピーが必要
5. **Error handling**: 単純な error set のみ対応

## Future Roadmap

### Short-term
- `?String` と `?Bytes` の Optional 対応
- より良いエラー診断

### Medium-term
- `union` 型のサポート
- カスタム allocator interface
- 性能最適化

### Long-term
- ジェネリック型のサポート
- ツールチェーン統合
- クロスプラットフォーム対応の改善

## Contributing

1. テストを実行: `go test ./...`
2. ベンチマークを実行: `go test -bench . ./asmcall`
3. Windows と Linux の両方でテスト
4. 新機能に合わせてドキュメントを更新

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).
