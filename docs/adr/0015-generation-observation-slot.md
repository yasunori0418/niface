# ADR-0015: subjectResult に観測型の世代遷移スロット generation を追加する

- ステータス: 採用
- 日付: 2026-07-07
- 関連: `spec/v1/spec.md`（§2, §5）, `schema/v1/envelope.schema.json`, ADR-0004, ADR-0005, ADR-0009, ADR-0010, ADR-0014

## 背景

各ツールは nix-env 経由のプロファイル化を予定しており（nput は実装済み）、「どの profile のどの世代に対する実行だったか」は全ツール共通の実行結果メタになる。現状この情報の置き場は `result.info`（ツール固有）しかなく（例: testdata plan-dryrun.json の `info.generation`）、キー名・構造がツール毎に割れると GC 調停・監査・nboot の世代同期が per-tool 知識に依存する。また ADR-0014 は世代を（profile 世代番号, その世代を適用 / 計画したエンベロープ）の対で表すと決めており、対の世代番号側をエンベロープ自身が規格語彙で運ぶ必要がある。（tracking #8 / #17）

## 決定

- `subjectResult` 直下に optional な `generation` を追加する（互換変更・specVersion 据え置き）。
- 構造は**観測型** `{ "profile": <string>, "before": <integer>, "after": <integer> }` とする。
  - `profile`: 実際に使用した profile のパス。`generation` を出力する場合は必須
  - `before` / `after`: 実行開始時点 / 終了時点で profile が指していた世代番号。観測できない場合はそれぞれ省略する（初回実行に `before` は無い。profile 未作成の plan には `after` も無い）
- 意味論は「作成の宣言」ではなく「**観測の記録**」とする。plan / dry-run では切替が起きないため `after` = `before`。新世代を作ったかは `before` ≠ `after` で判定する。
- profile を管理するツールは `generation` を出力すべきである（SHOULD）。schema 上は optional に留める（必須化は specVersion increment のため・→ ADR-0010）。
- id 導出は無改訂とする。「key に世代番号を含めない」（→ ADR-0004）は維持し、世代は item の同一性ではなく実行のメタ情報として扱う。

## 根拠

- 観測型は plan / apply / check / 失敗時のいずれでも**無条件に同じ意味**で出力できる。dry-run スキーマ同一原則（→ ADR-0009）と衝突せず、「これから作られる世代番号」という投機値を出さない。
- 全ツールで同じ意味を持つに至った情報を `info` から規格語彙へ昇格することは、needs 駆動の判断基準（→ ADR-0005 と同じ精神）に適合する。
- 世代は主体（config = profile）ごとの属性であり、`status` / `startedAt` と同格の実行メタである。subjectResult 直下への配置は「メタは外・ペイロードは result 配下」（→ ADR-0001）と整合し、複数主体実行でも主体ごとに世代が付く。
- 世代番号に正式な置き場を与えることで、identity / key へ世代を混入させる誘惑を構造的に断つ。

## 影響

- `spec/v1/spec.md` §2（SubjectResult の jsonc 例と規範文）・§5（世代は id に関与しない旨の整合確認のみ）。
- `schema/v1/envelope.schema.json`: `$defs/generation`（`additionalProperties: false`・`required: ["profile"]`）、`subjectResult.properties.generation`。
- `testdata/v1/`: valid に世代遷移（apply）・初回実行（before 無し）・plan（after = before）のベクタ、invalid に profile 欠落のベクタを追加。
- `go/types.go`: `Generation` 型と `SubjectResult.Generation`。
- `CONTEXT.md` / `docs/design.md`: 用語と索引の同期。
- id 導出系（`testdata/v1/id-vectors.json`・`go/id.go`・`nix/lib.nix`）は無改訂。

## 棄却した代替案

- **作成宣言型（`created` フィールド）**: 「この実行が作った世代」を出す案。plan では未確定番号の予測または省略になり、値の有無と意味がコマンドに依存して分岐する。
- **観測 + 作成の併記（before / after / created）**: `created` = `after` の重複が常態になり、不整合の余地だけが増える。
- **`result.info` 維持（現状のまま）**: キー名・構造がツール毎に割れ、GC 調停・監査・nboot が per-tool 知識に依存し続ける。
- **トップレベル配置**: 複数主体実行で主体ごとの世代を表現できない。
- **実行相関キー（runId）の同時追加（#18）**: システム世代の相関は ncompose が子エンベロープを item.info に束ねる**包含**で構造的に成立する。相関キーが効くのは束ねを経ないバラのエンベロープだけが手元にある場合のみで、これは ADR-0011 / 0013 が投機的として棄却した federation 前提と同型。発番者（ncompose）不在の現時点では死荷重であり、M3 の ncompose 設計で包含では足りないと実証されたときに互換追加する（#18 は凍結して残す）。
