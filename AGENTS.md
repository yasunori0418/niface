# CLAUDE.md

niface — n プレフィックスツール群(nput / nboot / nwrap / nherd / nshadow / ncompose)の実行結果を単一の構造化 JSON(結果エンベロープ)として stdout に出力するための共通規格。規格(spec / schema / testdata)とエコシステム中央ドキュメント(docs/ecosystem/)の 2 つを持つ。

## ドキュメント

作業の文脈に応じて参照する。

- `README.md` / `docs/ecosystem/overview.md` — 全体像
- `CONTEXT.md` — 用語集(規格の語彙の正名と避ける同義語)
- `spec/v1/spec.md` — 規範(RFC 2119 語彙)
- `docs/design.md` — 設計概要と ADR 索引
- `docs/adr/` — 設計判断の個別記録(背景・根拠・棄却案)。**規格を変更する前に必ず参照**

## 検証コマンド

```sh
nix flake check                    # id-vectors / go(適合検証 + ベクタ通過)を一括検証
cd go && go test ./...             # id-vectors 通過 + testdata 適合検証
nix run .#validate -- <file.json>  # 単一エンベロープを適合検証(niface-validate)
cd go && go generate ./...         # 正本(schema / id-vectors)変更後に embed コピーを手動同期
```

`testdata/v1/id-vectors.json` が言語間互換の要。id 導出実装(go/・nix/)を触ったら必ず全ベクタ通過を確認する。

正本(`schema/v1/`・`testdata/v1/id-vectors.json`)を変更したら `go generate ./...` で `go/internal/spec/` の embed コピーを同期する(自動では走らない)。同期し忘れは go test のバイト完全一致検査が CI で検出する。

## 不変条件

変更には `docs/adr/` への ADR 追加が必要。

- stdout は単一の valid JSON 文書のみ。進捗・ログは stderr
- status は success/error の 2 値。1 件でも item failed → error
- ツール固有フィールドは info 配下のみ。規格型は additionalProperties: false で閉じる
- item id は sha256(JCS(identity)) の 64 文字 hex。key にビルド毎に変わる値を入れない
- exit code は POSIX 慣行のまま。独自体系を作らない

## 規約

- ドキュメントは日本語。難読・文語的な漢字表現は使わない。JSON フィールドは camelCase
- 英語版 `README.md` を一次参照、`README.ja.md` はその日本語訳
- `spec/v1/spec.md` は MUST/SHOULD/MAY(RFC 2119)の語彙を維持する
- 互換変更(フィールド追加)は specVersion を上げない。削除・意味変更・必須化のみ specVersion increment(その場合 spec/v2/ を新設)
- identity の値域は spec §5 に定める(文字列/整数±(2^53−1)/bool/null/配列/オブジェクト(メンバー名 ASCII))。域外は各実装が id 導出時に拒否する。値域や JCS 実装を変える場合は Go/Nix/ベクタの 3 点を同時に更新する

## 開発環境

`flake.nix` は flake-parts ベース(規格の lib と checks を提供)。開発環境は `dev/flake.nix` にあり、`.envrc`(`use flake ./dev`)で読み込む。direnv 許可で mattpocock/skills が nput の project mode で `.claude/skills/` へ配置される。

## Agent 向けドキュメント

単一コンテキストのリポジトリ。用語集は `CONTEXT.md`、設計判断は `docs/adr/`。 skill がこれらをどう参照するかは `docs/agents/domain.md`。`/grill-with-docs` 方式で、用語や決定が固まったときに ADR とグロッサリを遅延生成する。
