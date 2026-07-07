# ADR-0018: エンベロープ直下にツール固有情報の置き場 info を追加する

- ステータス: 採用
- 日付: 2026-07-07
- 関連: `spec/v1/spec.md`（§2）, `schema/v1/envelope.schema.json`, `go/types.go`, ADR-0001, ADR-0007, ADR-0011
- 改訂対象: ADR-0007（ツール固有情報を `info` 配下に隔離する原則と `additionalProperties: false` は不変。「item / change / result の 3 箇所」の列挙をエンベロープ最上位を加えた 4 箇所へ拡張）

## 背景

`info` はツール固有情報の唯一の置き場だが（→ ADR-0007）、現状の置き場は item / change / result の 3 箇所のみで、**主体に紐づかない実行全体のツール固有情報**（実行環境の要約・全体設定・キャッシュ利用の統計等）を置ける場所が無い。複数主体実行では `result.info` は主体ごとに分かれるため、実行全体の情報を運ぼうとすると、全 `result.info` への複製か、規格外の最上位フィールド追加しか手が無い。前者は主体間の値の食い違いという不整合の余地を生み、後者は `additionalProperties: false` と衝突して producer 適合検証を通らない。（tracking #8 / #9）

## 決定

- エンベロープ直下に optional な `info`（type: object）を追加する（互換追加・specVersion 据え置き）。
- 意味論は他の 3 箇所の `info` と同一とする: ツール固有情報の置き場であり、規格はキー・構造を定めない。トップレベル `info` には主体に紐づかない実行全体の情報を置き、主体ごとのツール固有情報は従来どおり `result.info` に置く。
- `go/types.go` は `Envelope` に第 4 型パラメータ `TEnvInfo` を追加し `Info TEnvInfo` とする。go module は未消費のため型シグネチャの破壊は許容する。

## 根拠

- ADR-0007 の隔離原則（規格型は `additionalProperties: false` で閉じ、ツール固有は `info` 配下）をエンベロープ階層にも一貫適用するだけで、新しい概念や規約を持ち込まない。
- 実行全体のメタ情報はエンベロープ自身の属性であり、トップレベル配置は「メタは外・ペイロードは `result` 配下」（→ ADR-0001）と整合する。主体ごとの情報が `result.info` に置かれる責務分離も崩れない。
- Go の型を `map[string]any` でなく型パラメータにするのは、既存の `Item[T]` / `Result[...]` と同じく型付き消費を可能にするため（ツールを知らない消費側は `json.RawMessage` を渡せば規格部分だけを扱える）。型パラメータ追加は既存シグネチャの破壊だが、go module は未消費で破壊コストが無く、Generics パターンの一貫性を失う方が高くつく。

## 影響

- `spec/v1/spec.md` §2（トップレベルの jsonc 例と規範文）。
- `schema/v1/envelope.schema.json`: トップレベル `properties.info`（type: object）。
- `go/types.go`: `Envelope` に第 4 型パラメータ `TEnvInfo` と `Info` フィールド。
- `testdata/v1/valid/`: トップレベル `info` 付きベクタを追加。
- `CONTEXT.md` の info 項（「3 箇所」→「4 箇所」）と `docs/design.md` の索引。
- id 導出系（`testdata/v1/id-vectors.json`・`go/id.go`・`nix/lib.nix`）は無改訂。

## 棄却した代替案

- **各 result.info への複製**: 実行全体の情報を全主体の `result.info` に重複して置く案。複数主体実行で主体間の値の食い違いという不整合の余地だけが増え、消費側も「どの result.info を信じるか」の規約を別途必要とする。
- **トップ平置き（規格フィールドと同一階層にツール固有フィールドを追加）**: `additionalProperties: false` を破り、規格型を機械的に閉じられなくなる。ADR-0007 が棄却済みの案の再来であり採らない。
- **`Envelope.Info` を `map[string]any` にする**: 型パラメータ追加による破壊を避けられるが、go module は未消費で破壊コストが無い。既存の Generics 型付き消費との一貫性を優先する。
