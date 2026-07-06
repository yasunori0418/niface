# ADR-0001: ペイロードを result 配下に入れ子にする

- ステータス: 採用
- 日付: 2026-07-05
- 関連: `spec/v1/spec.md`（§2）, `schema/v1/envelope.schema.json`

> **2026-07-07 改訂注記（ADR-0011）**: 本 ADR の「メタ情報は最上位・ペイロードは入れ子」という原則は不変。ただしペイロードの容器は単数 `result` から `results[]`（主体ごとの SubjectResult の配列）に一般化され、ペイロードは `results[].result` に入れ子となった（→ ADR-0011）。

## 背景

全ツールが stdout に出す JSON エンベロープの最上位構造を決める必要がある。実行の中身（処理単位の結果・差分・ツール固有情報）を最上位に平置きするか、専用キーの配下に入れ子にするかで、消費側の型の閉じ方が変わる。

## 決定

- メタ情報（`specVersion` / `tool` / `command` / `status` / `dryRun` / `startedAt` / `finishedAt` / `errors`）は最上位に置く。
- 実行ペイロードは `result` 配下に入れ子にする（`result.items` / `result.changes` / `result.info`）。

## 根拠

- メタ情報とペイロードを構造で分離できる。消費側は `result` の有無と型だけを見てペイロードに到達でき、メタ情報のパースと混線しない。
- 将来ペイロードの種類が増えても `result` 配下に閉じ込められ、最上位の契約を汚さない。

## 影響

- `schema/v1/envelope.schema.json` の最上位に `result` オブジェクトを置く。
- `result` 直下も規格型として閉じる（ツール固有は `result.info` のみ・ → ADR-0007）。

## 棄却した代替案

- **フラット構造（`items` / `changes` を最上位に平置き）**: メタ情報とデータが同一階層で混ざり、「ペイロード全体」を 1 単位として型付け・分岐できない。
