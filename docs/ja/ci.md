# CI 説明

現在の CI が主にカバーしているのは次の経路です。

- Windows のメイン経路
- Linux の生成 / コンパイル経路

## CI が行うこと

Windows job:
- Go と Zig をインストールする
- `examples/basic` を生成する
- `go test ./...` を実行する
- `go test -bench . ./asmcall` を実行する

Linux job:
- Go と Zig をインストールする
- `examples/basic` を生成する
- `go test ./...` を実行する
- `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...` を実行する

## なぜ Linux runtime の実行テストはデフォルトで無効なのか

現在、Linux 上の no-`cgo` runtime は、性能を重視した低層実装の段階にあります。

メイン CI をより安定させるため、現在のデフォルト方針は次のとおりです。

- メイン CI では Linux 経路が生成、コンパイル、統合ビルドできることを検証する
- 低層 runtime の実行テストは環境変数で手動有効化する

これにより、まだ磨き込み中のアセンブリ呼び出し詳細に、すべての CI 結果が依存してしまうのを避けられます。

## 今後さらに実施できること

- 手動トリガー式の Linux runtime 詳細検証 workflow を追加する
- benchmark 結果のアーカイブを追加する
- action エコシステムが安定したら、Node 24 対応版へ統一移行する
