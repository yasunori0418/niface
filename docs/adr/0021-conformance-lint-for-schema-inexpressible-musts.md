# ADR-0021: schema で表現しきれない MUST を適合検査に取り込む

- ステータス: 採用
- 日付: 2026-07-10
- 関連: `spec/v1/spec.md`（§2, §5, §8）, `schema/v1/envelope.schema.json`, `scripts/validate.py`, `testdata/v1/`, ADR-0002, ADR-0004, ADR-0006, ADR-0012

> **2026-07-10 改訂注記（ADR-0023）**: 本 ADR の強制対象 MUST（status 整合・itemId 参照整合・一意性）と schema/lint の分担は不変。ただし適合検証の**参照実装**を `scripts/validate.py`（Python + jsonschema）から Go 実装（`santhosh-tekuri/jsonschema` + lint、`niface-validate` CLI）へ移し、Python は撤去した（→ ADR-0023）。

## 背景

§8 は適合を「`schema/v1/` の JSON Schema 検証 + `testdata/v1/id-vectors.json` の全ベクタ通過」で定義する。しかし規格の MUST の一部は純 JSON Schema で表現できず、この定義を素通りする。

- **status ⇔ errors / item の整合**（§2）: `status` は 2 値で「1 件でも failed → error」「主体列挙前の全体エラーがあれば error」の集約規則（→ ADR-0002）を持つが、schema 上 `status` は単なる enum で、`errors` の有無・`results[]` の status・item の status と連動していない。`status: "error"` なのに errors も error 主体も無い出力、`status: "success"` なのに error 主体（や failed item）を抱える出力が schema を通過する。
- **`changes[].itemId` の同一 result 内参照整合**（§5）: `itemId` は同じ subjectResult の `items[].id` を指さねばならない（result を跨いだ参照は MUST NOT）が、schema は 64 文字 hex の形式しか見ず、参照先の実在も同一 result 内かも検査しない。
- **一意性 MUST**（§5, producer MUST）: `subject.name` の 1 エンベロープ内一意、`item.id` の 1 result 内一意は、schema の `uniqueItems`（要素まるごと一致しか見ない）では表現できない。

これらは意図された意味論だが producer 適合検証が強制しておらず、適合と無関係に破れる。W_ / E_ の取り違えを型分離で塞いだ ADR-0019 と同じ問題が、status 整合・参照整合・一意性に残っている。（tracking #8 / #15）

## 決定

- **§8 の適合定義を拡張する**: 適合 = 「JSON Schema 検証 + schema で表現しきれない MUST のリント検査 + id-vectors 全ベクタ通過」。リント検査の参照実装は `scripts/validate.py`。MUST 自体の規範は §2 / §5 に残し、§8 は適合検証がそれらをカバーすることを宣言する。
- **schema で表現できるものは schema を強化する**（if/then + contains）:
  - トップレベル: `status: "error"` ⟹ `errors[]` 非空 ∨ status=error の result あり。`status: "success"` ⟹ その否定（errors 空 ∧ error 主体なし）。
  - `subjectResult`: `status: "error"` ⟹ failed item あり ∨ `errors[]` 非空。`status: "success"` ⟹ その否定。
- **schema で表現できないものは `validate.py` にリント層を足す**（再利用可能な `lint_envelope(doc)` に分離）:
  - `changes[].itemId` は同一 subjectResult の `items[].id` を参照する（result 跨ぎ・dangling は違反）。
  - `subject.name` は 1 エンベロープ内で一意。
  - `item.id` は 1 result 内で一意。
- リント強化は既存 valid ベクタを 1 件も落とさない（従来 valid だった正当な出力を invalid 化しない）。schema の if/then 強化は validation 強化だが、invalid 化されるのは意図に反する不整合出力のみで、下書き段階の修正として specVersion は 1 のまま据え置く（ADR-0013 / ADR-0019 と同判断）。

## 根拠

- 表現できるものは schema、できないものだけリント、と分けることで、機械可読の正（schema）を最大限使い、script 依存を最小の残余に閉じる。schema の if/then は item.allOf（failed → error 必須）で既に使う技法で、新しい仕組みを増やさない。
- subjectResult 階層の status 整合まで強制するのは、トップレベルの success 方向チェックが `results[].status` の自己申告に依存するため。result が「success」と偽って failed item を抱えると、トップレベル整合だけでは穴が残る。item 階層まで塞いで初めて「success なのに失敗が紛れる」を検出できる。
- リントを §8 の適合の柱に格上げすることで、producer 適合検証が schema の表現限界を越えて MUST をカバーする。ADR-0019 が型分離で W_ / E_ 取り違えを検証に取り込んだのと同じ方向の補完。
- 一意性・参照整合を validate.py に置くのは、JSON Schema がこれらを構造的に表現できない（`uniqueItems` はフィールド単位の一意を見ない・`contains` は参照解決をしない）ため。schema を歪めて近似するより、明示的なリントの方が読みやすく誤検出も無い。

## 影響

- `spec/v1/spec.md` §8: 適合定義にリント検査を追記。§2 / §5 の MUST 文面は無改訂（意味論は元から規範）。
- `schema/v1/envelope.schema.json`: トップレベルと `$defs/subjectResult` に status 整合の `allOf`（if/then + contains）を追加。
- `scripts/validate.py`: `lint_envelope(doc)` を新設し、valid は schema ∧ lint 全通過、invalid は schema ∨ lint のいずれかが弾く、という判定に拡張。
- `testdata/v1/invalid/`: schema 4 分岐（top × subjectResult × error / success 方向）と lint 3 検査（itemId 跨ぎ・subject.name 重複・item.id 重複）の 7 ベクタを追加。
- `CONTEXT.md` の「適合 (conformance)」項、`docs/design.md` の ADR 索引と適合戦略。
- id 導出系（`testdata/v1/id-vectors.json`・`go/`・`nix/`）は無改訂。

## 棄却した代替案

- **全て schema で表現する**: 一意性・参照整合を schema に押し込む案。`uniqueItems` はフィールド単位の一意を見られず、`contains` は「同一 result 内の item を参照」という参照解決を表現できない。無理に近似すると schema が歪み誤検出を生む。表現限界を越える部分はリントに分けるのが正しい。
- **リントを testdata 品質ゲート止まりにし §8 を据え置く**: validate.py のリントを「この規格リポジトリの testdata 検査」に留める案。第三者 producer は schema + id-vectors だけで適合判定されるため、これらの MUST がすり抜ける問題（本 ADR の動機）が未解決のまま残る。
- **subjectResult 階層の status 整合を入れない**: issue の 4 項目リストに忠実にトップレベル整合だけ入れる案。トップレベルの success 方向チェックが result の自己申告に依存し、「subject.status=success なのに failed item」が schema・lint 双方を通過する穴が残る。同じ if/then 技法で安く塞げるため見送る理由が無い。
- **specVersion increment**: if/then 強化は validation 強化だが、invalid 化されるのは意図に反する不整合出力のみで、下書きを consume した実装は存在しない。v2 分岐は過剰（ADR-0013 / ADR-0019 と同判断）。
