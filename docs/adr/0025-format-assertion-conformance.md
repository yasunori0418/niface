# ADR-0025: format 検証を assertion として規範化し、適合検査の素通りを塞ぐ

- ステータス: 採用
- 日付: 2026-07-11
- 関連: `spec/v1/spec.md`(§2, §8), `schema/v1/envelope.schema.json`, `go/conformance/conformance.go`, `testdata/v1/invalid/`, ADR-0010, ADR-0021, ADR-0023

## 背景

schema は `startedAt` / `finishedAt` に `format: date-time` を宣言し、spec §2 の例にも `// RFC 3339` と注記していたが、prose の規範(MUST)は無く、適合検証も日時形式を実際には検査していなかった。

- JSON Schema Draft 2020-12 は `format` を既定で**注釈(annotation)扱い**とし、assertion として評価しない。go/conformance が使う `santhosh-tekuri/jsonschema` も既定は無効で、`startedAt: "not-a-date"` が schema 検証を通過していた。schema に `format: date-time` と書いてあるのに検証されない = format 検証が実質 no-op という穴が残っていた。
- この穴は #15 が「format 検証の no-op」として認定していたが、その後の適合検証の Go 集約(ADR-0023、`scripts/validate.py` 撤去)で FormatChecker 前提の対応計画ごと消え、未対応のまま #15 がクローズされた。
- ADR-0021 は「schema で表現しきれない MUST」をリント層に取り込んだが、format は「schema で表現できているのに評価されていない」別種の穴で、リントの対象外だった。§8 の適合定義「schema 検証を通ること」も、Draft 2020-12 の意味論では format を含むか曖昧で、第三者検証器の format 設定次第で適合判定が割れる余地があった。

本 ADR は「format を宣言したが評価していない」空白を埋める新規決定であり、ADR-0021 の適合層構成(schema / lint / id-vectors)・ADR-0023 の Go 集約は改訂しない。

## 決定

- spec §2 に `startedAt` / `finishedAt` の RFC 3339 MUST を prose で明文化する(`subjectResult` の同フィールドも同じ制約に従う)。
- spec §8 に「(1) の schema 検証は `format` を assertion として評価する」を明記する。Draft 2020-12 の既定(format = 注釈扱い)に依らず、niface の適合はこの評価を要求する。
- go/conformance の `NewChecker` で `santhosh-tekuri/jsonschema` の `AssertFormat()` を呼び、format assertion を有効化する。santhosh-tekuri は `date-time` を RFC 3339 に基づき検証する(`T` 区切り必須・オフセット必須。閏秒や秒小数などの裁量部分の扱いは同実装に委ねる)。既存 API の呼び出しのみで**依存追加なし → vendorHash 不変 → `flake.nix` 不変**。
- `testdata/v1/invalid/` に不正日時ベクタを追加する(top-level `startedAt` / `subjectResult` の `finishedAt`)。他フィールドは valid に保ち、format assertion のみが弾くことを保証する。
- `specVersion` は 1 のまま。`format: date-time` は当初から宣言済みで RFC 3339 が意図だった。評価漏れの修正であり新規制約でも意味変更でもない(ADR-0021 / ADR-0024 と同判断)。

## 根拠

- schema が既に持つ宣言(`format: date-time`)を評価に回すだけで穴が塞がる。日時検証をリント層(ADR-0021)へ自前実装するより安く、「表現できるものは schema を使う」方針(ADR-0021 根拠)と一致する。
- `AssertFormat()` は santhosh-tekuri の既存 API。新依存ゼロで `vendorHash` / `flake.nix` は不変。検証器の設定は conformance パッケージに閉じ、id 導出パッケージ(`go/id.go`)へは波及しない(ADR-0023 の分離を維持)。
- `date-time` の検証は santhosh-tekuri が RFC 3339 に基づき行う(細部は同実装依存)。日時検証は適合検証器(Go)の領分で、Nix は id 導出のみを担い日時に関与しないため、値域変更のような Go/Nix/ベクタ 3 点同期の対象外。
- 既存 valid ベクタの `startedAt` / `finishedAt` は全て正しい RFC 3339 で、assertion 有効化で 1 件も落ちない(ADR-0021 の「valid を invalid 化しない」制約と同じ)。

## 影響

- `spec/v1/spec.md`: §2 に日時の RFC 3339 MUST、§8 に format assertion 評価を追記。
- `go/conformance/conformance.go`: `NewChecker` に `AssertFormat()`。`conformance_test.go`: 日時形式を強制する狙い撃ちテスト。
- `testdata/v1/invalid/`: `started-at-not-rfc3339.json` / `subject-finished-at-not-rfc3339.json` を追加。
- `docs/design.md` の ADR 索引に本 ADR を追加。
- `schema/v1/`(`format: date-time` は宣言済み)・`go/id.go`・`go/types.go`・`nix/`・`flake.nix` は無改訂。`specVersion` は 1 のまま。

## 棄却した代替案

- **format をリント層(ADR-0021)で自前検証する**: schema が既に `format: date-time` を持つのに、日時検証を Go の手書きリントへ二重化することになる。`AssertFormat()` で schema 宣言をそのまま評価する方が単純で、検証器の RFC 3339 実装を再利用できる。
- **spec §2 の例コメントのみで RFC 3339 を示し prose MUST を書かない**: 現状維持。第三者検証器の format 設定次第で適合判定が割れ、§8 の「schema 検証を通ること」が format を含むか曖昧なまま残る。prose MUST と §8 の assertion 明記で、first-party ツールが従う profile(`T` 区切り + オフセット必須)に判定を揃える。RFC 3339 の裁量部分(空白区切りの可否等)まで完全に一意化するわけではないが、素通りしていた no-op を塞ぎ実用上の割れを無くす。
- **format assertion 有効化を非互換とし `specVersion` を上げる(v2 新設)**: `format: date-time` は当初から宣言され RFC 3339 が意図だった。評価漏れの修正で、不正日時を出す下書き実装だけが invalid 化される。v2 分岐は過剰(ADR-0021 / ADR-0013 / ADR-0019 と同判断)。
