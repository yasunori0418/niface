# CLAUDE.md

niface — n プレフィックスツール群(nput / nboot / nwrap / nherd / nshadow /
ncompose)が stdout/stdin で会話するための共通 JSON 規格。規格(spec / schema /
testdata)とエコシステム中央ドキュメント(docs/ecosystem/)の 2 つを持つ。

## ドキュメント

作業の文脈に応じて参照する。

- `README.md` / `docs/ecosystem/overview.md` — 全体像
- `spec/v1/spec.md` — 規範(RFC 2119 語彙)
- `docs/design.md` — 全 19 設計判断とその根拠。**規格を変更する前に必ず参照**
- `docs/HANDOFF.md` — 前回セッションの到達点・未完了タスク

## 検証コマンド

```sh
nix flake check                    # id-vectors / schema / go を一括検証
python3 scripts/validate.py schema/v1/envelope.schema.json testdata/v1
cd go && go test ./...             # id-vectors 通過テスト
```

`testdata/v1/id-vectors.json` が言語間互換の要。id 導出実装(go/・nix/)を触ったら
必ず全ベクタ通過を確認する。

## 不変条件

変更には `docs/design.md` への判断追記が必要。

- stdout は単一の valid JSON 文書のみ。進捗・ログは stderr
- status は success/error の 2 値。1 件でも item failed → error
- ツール固有フィールドは info 配下のみ。規格型は additionalProperties: false で閉じる
- item id は sha256(JCS(identity)) の 64 文字 hex。key にビルド毎に変わる値を入れない
- exit code は POSIX 慣行のまま。独自体系を作らない

## 規約

- ドキュメントは日本語。難読・文語的な漢字表現は使わない。JSON フィールドは camelCase
- 英語版 `README.md` を一次参照、`README.ja.md` はその日本語訳
- `spec/v1/spec.md` は MUST/SHOULD/MAY(RFC 2119)の語彙を維持する
- 互換変更(フィールド追加)は specVersion を上げない。削除・意味変更・必須化のみ
  specVersion increment(その場合 spec/v2/ を新設)
- JCS 実装はサブセット。拡張する場合は Go/Nix/ベクタの 3 点を同時に更新する

## 開発環境

`flake.nix` は flake-parts ベース(規格の lib と checks を提供)。開発環境は
`dev/flake.nix` にあり、`.envrc`(`use flake ./dev`)で読み込む。direnv 許可で
mattpocock/skills が nput の project mode で `.claude/skills/` へ配置される。
