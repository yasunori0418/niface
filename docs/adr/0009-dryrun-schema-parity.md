# ADR-0009: dry-run は apply と同一スキーマで出力する

- ステータス: 採用
- 日付: 2026-07-05
- 関連: `spec/v1/spec.md`（§2, §3, §4, §7）, ADR-0003

## 背景

plan / dry-run の出力を apply と別スキーマにするか、同一スキーマにするかで、消費側が扱う形式の数が変わる。

## 決定

- 最上位に `dryRun: boolean` を置く。
- `dryRun: true` の出力は、`dryRun` の値以外において apply と**同一スキーマ**でなければならない。
- plan では `changes` が「実行したら生じる差分」、apply では「実際に生じた差分」を表す。
- dry-run の items は原則 `success` とし、`skipped`（前段の失敗による未実行）は使わない。

## 根拠

- スキーマを分けると、消費側が plan 用と apply 用の 2 形式を実装・検証しなければならない。同一スキーマなら検証・消費コードを共有でき、ncompose も plan → apply で処理を切り替えずに済む。
- 「plan で見えた差分」と「apply で起きた差分」を同じ型で比較できる。

## 影響

- `spec/v1/spec.md` §2（`dryRun`）・§3（item）・§4（change）・§7（check も `dryRun` を出力）。
- `schema/v1/envelope.schema.json`（`dryRun` の値以外に差を設けない）。

## 棄却した代替案

- **plan 専用の result 型**: 消費側が 2 形式を二重実装することになる。
- **dry-run で items を省略する**: plan の情報量が落ち、「何が起きるか」を item 粒度で見られなくなる。
