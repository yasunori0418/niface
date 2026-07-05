# ADR-0007: ツール固有フィールドを info 配下に隔離する

- ステータス: 採用
- 日付: 2026-07-05
- 関連: `spec/v1/spec.md`（§2, §3, §4）, ADR-0001, `go/types.go`

## 背景

各ツールは規格が定めない固有の情報を出力したい（nput なら配置先パスなど）。これを規格フィールドと同一階層に混ぜると、規格型を機械的に閉じられず、型付き消費が難しくなる。

## 決定

- ツール固有情報は `info` 配下にのみ置く。item / change / result の 3 箇所とも同じパターン。
- 規格型は `additionalProperties: false` で閉じる。

## 根拠

- 同一階層マージは Go の `encoding/json` で flatten 非対応のため二段デコードを強い、JSON Schema も `additionalProperties` 頼みで規格型を閉じられない。
- `info` 一段掘りにより、規格型は `additionalProperties: false` で閉じ、`Item[T any]` の Generics で型付けでき、ツールを知らない消費側は `Item[json.RawMessage]` で規格部分だけを安全に扱える。
- 代償は jq パスが一段深くなること（`.info.target`）のみで、閉じた型と型安全な消費に見合う。

## 影響

- `spec/v1/spec.md` §2・§3・§4。
- `schema/v1/envelope.schema.json` の全規格オブジェクトに `additionalProperties: false` と `info`。
- `go/types.go` の Generics 型。

## 棄却した代替案

- **規格フィールドと同一階層にツール固有フィールドをマージ**: 規格型を閉じられず、二段デコードを強いる。
- **`info` を持たずツール拡張を禁止**: 実用に耐えない（各ツールが固有情報を運べない）。
