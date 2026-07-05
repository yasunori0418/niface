# ADR-0002: status を 2 値にし部分失敗を error へ集約する

- ステータス: 採用
- 日付: 2026-07-05
- 関連: `spec/v1/spec.md`（§1, §2, §3）, ADR-0008

## 背景

エンベロープ全体の成否をどう表現するか。`partial`（一部成功）のような中間状態を最上位 `status` に持たせるかどうかで、消費側（特にオーケストレータ ncompose）の分岐の複雑さが変わる。

## 決定

- 最上位 `status` は `success` / `error` の **2 値**とする。
- item 単位にも `status`（`success` / `failed` / `skipped`）を持たせる。
- **1 件でも item が failed なら**最上位 `status` は `error`、exit code は非 0 とする（→ ADR-0008 で exit code と連動）。

## 根拠

- `partial` のような中間値は「partial のとき何をすべきか」の解釈を消費側に強いる。 2 値 +「1 件でも失敗なら error」はオーケストレータの分岐を単純化する（非 0 なら止める / 巻き戻す）。
- **何が適用済みか**という詳細は `result.items` が機械判定可能な形で運ぶので、全体 status を細分化する必要がない。
- nput の部分失敗集約（continue + aggregate）の意味論をエコシステム全体へ広げたものであり、既存ツールの挙動と整合する。

## 影響

- `spec/v1/spec.md` §1（exit code 連動）・§2（status 2 値）・§3（item status）。
- `schema/v1/envelope.schema.json` の `status` enum と item の `status` enum。
- `testdata/v1/invalid/` に「item が failed なのに全体 success」を拒否するベクタを置く。

## 棄却した代替案

- **`partial` 中間値を設ける**: 消費側に解釈を強い、分岐が増える。詳細は items で足りる。
- **全体 status のみで item status を持たない**: どの単位が失敗したか機械判定できず、rollback 指揮の粒度が取れない。
