# ADR-0003: items（実行結果）と changes（差分宣言）を分離する

- ステータス: 採用
- 日付: 2026-07-05
- 関連: `spec/v1/spec.md`（§3, §4, §7）, ADR-0009

## 背景

「処理単位の実行結果」と「状態遷移（差分）」は別の概念だが、1 つの配列に混ぜて表現することもできる。また check サブコマンドの結果に専用型を設けるべきかも論点になる。

## 決定

- `result.items` = **処理単位の実行結果の記録**。規格で共通型を強制する。
- `result.changes` = **状態遷移（差分）の宣言**。差分のある項目のみを列挙し（noop を含めない）、`reversible`（この差分単位で巻き戻し可能か）を要素に必須とする。
- check サブコマンドは専用の result 型を設けず、**前提条件 1 つ = Item 1 つ**で表現し items を再利用する。

## 根拠

- 実行結果と差分は別概念であり、混ぜると型と意味の両方が曖昧になる。
- dry-run スキーマ同一原則（→ ADR-0009）により plan / apply の両方が changes を出力する。plan では「実行したら生じる差分」、apply では「実際に生じた差分」。
- 可逆性（`reversible`）は**行為の属性**なので changes 要素に置く。item は結果の記録であり可逆性を持たない。ncompose の rollback 指揮は changes の走査で完結する。
- check を items で表現すれば、サブコマンドごとに result 型を増やさずに済む。

## 影響

- `spec/v1/spec.md` §3（Item）・§4（Change）・§7（check）。
- `schema/v1/envelope.schema.json` の `result.items` / `result.changes`。

## 棄却した代替案

- **items と changes を 1 配列に統合**: 実行結果と差分という別概念が混在し、可逆性を置く場所も曖昧になる。
- **check 専用の result 型**: サブコマンドごとに型が増える。前提条件 = Item で十分表現できる。
- **noop も changes に列挙**: 差分の無い項目まで並び、rollback 対象の走査が冗長になる。
