# ADR-0016: 部分失敗時も適用済み changes を必ず出力する

- ステータス: 採用
- 日付: 2026-07-07
- 関連: `spec/v1/spec.md`（§4, §7）, ADR-0003, ADR-0009, ADR-0014, ADR-0015

## 背景

ADR-0014 は rollback の呼び出し規約を ncompose の領分とし、niface は判断材料（`changes` / `reversible` / 適用済み差分）の完全性のみを担うと決めた。しかし spec §4 は「差分のある項目のみを列挙する」としか規定しておらず、apply が途中で失敗したとき `changes[]` がどこまで完全であるべきかは未規定だった。特に item が `failed` のとき change を出すべきかは曖昧で、「途中まで適用されて失敗した」（例: 旧 symlink を除去した後、新 symlink の作成に失敗した）場合に現実の変化が changes から欠落しうる。ncompose の rollback 指揮は changes の走査で完結する設計（→ ADR-0003）のため、失敗した実行でこそ changes の完全性が要る。（tracking #8 / #11）

## 決定

- apply（`dryRun: false`）の `changes[]` の意味論を**観測主義**で確定する: changes は実際に生じた状態遷移の観測記録であり、change を出すかどうかは item の status と独立に「現実が変化したか」で決まる。item が `failed` でも、失敗までに現実が変化したなら対応する change を出力する。
- failed item に紐づく change の `kind` は意図した遷移ではなく**実際に生じた遷移**を表す（旧 symlink 除去後に失敗したなら `modify` ではなく `remove`）。`reversible` はその生じた遷移についてツールが判断する。
- apply では、`status` が `error` で終わる実行でも、`changes[]` は失敗時点までに実際に生じた差分を**全て**含めなければならない（MUST）。
- plan（`dryRun: true`）が error で終わる場合の changes の完全性は要求しない。check は副作用を持たないため対象外（§7）。

## 根拠

- 失敗した実行こそ rollback・復旧の判断材料が要る。適用済み差分が changes に漏れなく残ることは、ADR-0014「niface は判断材料の完全性で支える」の具体化である。
- 観測主義は generation スロット（→ ADR-0015）と同じ意味論であり、「宣言」ではなく「観測」なら成功・失敗を問わず無条件に同じ意味で出力できる。ADR-0009 の「apply では実際に生じた差分」という既存文言の自然な精緻化でもある。
- item の status（結果の記録）と change の有無（差分の観測）は別概念であり、failed だから change を省くのは両者の再結合になる（→ ADR-0003 の分離原則）。
- plan は現実を変えないため「実際に生じた差分」が存在せず、error で終わった plan に完全性を課しても「出力できた分の予測」以上の意味を持たない。検証もできない規範は課さない。

## 影響

- `spec/v1/spec.md` §4 に観測主義と error 時の changes 完全性 MUST を追記。
- `testdata/v1/valid/`: 部分失敗 + 適用済み changes（failed item に紐づく change を含む）のベクタを追加。
- `schema/v1/` / `go/`: 変更なし（規範文のみ。「現実との一致」は schema では検証できない）。
- `CONTEXT.md`（change 項）/ `docs/design.md`（索引）: 同期。

## 棄却した代替案

- **failed item の change を省略する**: item の status から差分の有無を推測させる案。「途中まで適用されて失敗した」現実の変化が不可視になり、rollback の判断材料が欠ける。items（結果の記録）と changes（差分の宣言）の分離（→ ADR-0003）にも反する。
- **plan への MUST 拡張**: plan の changes は予測であり、error 時の「全て」が定義できない（実行していないため現実の差分が存在しない）。守れているか検証できない MUST は適合判定を壊す。
