# go2zig ドキュメント

Languages: [English](../en/README.md) | [简体中文](../zh/README.md) | [日本語](README.md)

`go2zig` は現在、`cgo` を使わない `Go -> Zig` 呼び出し経路に注力しており、主な目標は次のとおりです。

- Go 側では、通常の SDK に近い呼び出し体験を維持する
- ランタイム側では、`syscall` / `cgo` による追加オーバーヘッドを可能な限り抑える
- ジェネレータによって ABI、frame、エラープロトコル、文字列 / バイト列変換を一元化する

## 対応プラットフォーム

現在サポートしている環境:
- `windows/amd64` - 完全対応、CI テストあり
- `linux/amd64` - 完全対応、CI テストあり

未対応:
- `arm64` - 今後の実装予定
- `macOS` - 現時点では未対応
- その他のアーキテクチャ - 現時点では未対応

## 対応型の概要

### 完全対応している型
- 基本型: `bool`、`u8-u64`、`i8-i64`、`f32`、`f64`
- 複合型: `extern struct`、`enum(整数型)`、固定長配列
- 特殊型: `String`、`Bytes`
- slice alias: POD slice（例: `ScoreList = extern struct { ptr: ?[*]const u16, len: usize }`）
- optional 型: `?POD`（例: `?u32`、`?UserKind`）
- エラー処理: `error{...}!ReturnType`

### 非対応の型
- Go 固有: `map`、`chan`、`interface{}`、関数型、ポインタ
- Zig 固有: `union`、`comptime`、`@import`
- 制限付き対応: optional 型は POD のみ対応、slice 要素には String/Bytes を使えない

## 推奨の読書順

1. `docs/ja/architecture.md` - 全体アーキテクチャを把握する
2. `docs/ja/usage.md` - 基本的な使い方を学ぶ
3. `docs/ja/generator.md` - ジェネレータの詳細を理解する
4. `docs/ja/runtime.md` - ランタイム設計を理解する
5. `docs/ja/testing.md` - テスト方法を把握する
6. `docs/ja/ci.md` - CI 構成を確認する

## クイックスタート

まずは素早く全体像を掴みたい場合、リポジトリ直下の `README.md`（英語）、`README_zh.md`（中国語）、または `README_ja.md`（日本語）を読み、あわせて `examples/basic` を参照すると、完全な生成フローを理解しやすくなります。

## パフォーマンスベンチマーク

Windows 開発機で直近に観測された代表的な結果は次のとおりです。

- `BenchmarkCgoAddU64`: `28.56 ns/op`
- `BenchmarkAsmCallCAddU64`: `3.352 ns/op`

つまり、ごく短い同期呼び出しでは、現在の no-`cgo` asm 経路は `cgo` よりおよそ `8x` 高速です。

## メモリ管理

現在の実装では、固定のメモリ管理方式を採用しています。

- **割り当て側**: Zig がメモリを確保する
- **解放側**: Go が解放する（`go2zig_free_buf` を使用）
- **変換コスト**: データのコピーとライフサイクル管理が必要になる

## 現在の制約

1. **プラットフォーム制約**: `amd64` の Windows と Linux のみ対応
2. **型制約**: Go の `map`、channel、interface などの固有型は未対応
3. **メモリ管理**: 固定の割り当て方式であり、allocator はカスタマイズできない
4. **性能コスト**: 各呼び出しでデータコピーが必要

## 今後の拡張方向

### 優先度高
- `arm64` アーキテクチャ対応
- `?String` と `?Bytes` optional 型のサポート
- エラー診断の改善

### 優先度中
- `union` 型サポート
- カスタム allocator interface
- 性能最適化

### 優先度低
- ジェネリクス対応
- ツールチェーン統合
- クロスプラットフォーム改善

## 関連ドキュメント

### 日本語ドキュメント
- [英語 README](../../README.md)
- [中国語 README](../../README_zh.md)
- [日本語 README](../../README_ja.md)
- [アーキテクチャ概要](architecture.md)
- [使用ガイド](usage.md)
- [ジェネレータ説明](generator.md)
- [ランタイム設計](runtime.md)
- [テストとベンチマーク](testing.md)
- [CI 説明](ci.md)

### English Documentation
- [Architecture Overview](../en/architecture.md)
- [Usage Guide](../en/usage.md)
- [Generator Guide](../en/generator.md)
- [Runtime Design](../en/runtime.md)
- [Testing & Benchmarks](../en/testing.md)
- [CI Guide](../en/ci.md)

### 中文文档
- [架构概览](../zh/architecture.md)
- [使用指南](../zh/usage.md)
- [生成器详情](../zh/generator.md)
- [运行时设计](../zh/runtime.md)
- [测试与基准](../zh/testing.md)
- [CI 配置](../zh/ci.md)
