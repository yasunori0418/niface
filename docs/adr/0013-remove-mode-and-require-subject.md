# ADR-0013: mode 判別子を廃止し subject を常時必須にする

- ステータス: 採用
- 日付: 2026-07-07
- 関連: `spec/v1/spec.md`（§2, §5）, `schema/v1/envelope.schema.json`, ADR-0011, ADR-0012
- 改訂対象: ADR-0011（常時 `results[]` は不変。判別子 `mode` と single/batch 条件分岐を除去）, ADR-0012（subject 導入と 3 層参照キーは不変。single で任意だった subject を常時必須へ格上げ）

## 背景

ADR-0011 が導入した `mode`（`single` / `batch`）は、エンベロープの実行形態判別子として必須化された。しかし容器が常に `results[]` に統一された今、`mode` は結果件数を表さず（single / batch とも 0 / 1 件を取りうる）、判別子としての価値が薄い。現スキーマで `mode` が実際に担っているのは (1) `single` → `results` maxItems:1、(2) `batch` → 各要素 `subject` 必須、(3) 疎結合消費者向けの判別、の 3 点のみである。このうち (3) は ADR-0011 が投機的として棄却した federation（単独再流通・複数ツール集約）前提に依存しており弱く、(1)(2) は起動の性質を件数と subject 有無で二重管理しているだけである。加えて ADR-0012 で single の subject を任意としたため、§5 の参照キー 3 つ組 `(tool.name, subject, id)` が single では欠ける穴が残っていた。

ADR-0011 / 0012 は本日採用でまだ何も consume しておらず、nput も未完のため、v2 を切らず未リリース下書き段階の v1 を修正する。

## 決定

- `mode` フィールドをスキーマ・spec・型・testdata から全廃する。実行形態を件数・起動性質で二重管理する判別子を除去する。
- `subjectResult.subject` を single / batch を問わず**常時必須**にする。single / batch の条件分岐（`allOf` の `if/then`）と `single` の `results` maxItems:1 を全て削除する。
- 「複数主体を 1 実行で扱う」概念（`apply --all`）自体は残す。消える語は `mode` という**判別子フィールド**のみで、複数主体実行は非公式呼称としての「batch」で呼ぶに留める。
- specVersion は 1 のまま据え置く。未リリース下書き段階の非互換修正として、プロジェクト規約「必須削除は specVersion increment」の文面とのズレを許容する。

## 根拠

- 容器が常に `results[]` に統一された以上、実行形態の判別子は結果の形を決めない。`mode` は起動性質（`--all` か否か）と件数・subject 有無の二重管理に堕しており、除去してもコンテナの一様性は損なわれない。
- subject を常時必須にすることで、§5 の参照キー `(tool.name, subject, id)` が single でも欠けず、`item.id` と同格に `subject` が全 subjectResult に常在する。参照解決の 3 層が構造で保証される。
- (3) の疎結合消費者向け判別は federation 前提に依存し、その federation 自体が ADR-0011 で投機的として棄却済み。弱い価値のために判別子を必須で抱えるのは割に合わない。

## 影響

- `spec/v1/spec.md` §2（`mode` 記述・jsonc・`tool`/`command`/`mode`/`specVersion` 列挙の削除、判別子説明の書換え）・SubjectResult §（subject を常時必須へ）・§5（subject 常在への整合。空 `results` = success の vacuous 規定は維持）。
- `schema/v1/envelope.schema.json`: `required` から `mode` 除去、`mode` プロパティ定義削除、`allOf` ブロック削除、`$defs/subjectResult.required` に `subject` 追加、description の single/batch/mode 言及調整。
- `go/types.go`: `Mode` 型・`ModeSingle`/`ModeBatch` const・`Envelope.Mode` 削除、`SubjectResult.Subject` を非ポインタ・必須へ。
- `testdata/v1/`: 全ファイルから `mode` 行を除去、subject 必須化に伴う subject 付与、`mode`/single 条件前提の invalid ベクタ（`mode-missing` / `single-with-multiple-results`）削除、`batch-subresult-without-subject` を `subresult-without-subject` へ改名、`batch-*` を実態名へ改名。
- id 導出系（`testdata/v1/id-vectors.json`・`go/id.go`・`nix/lib.nix`）は無改訂（subject は id 導出に非関与のまま）。

## 棄却した代替案

- **subject を常時任意のまま SHOULD 文で担保する**: `mode` だけ廃し subject は任意に据え置き、「subject を出すべき（SHOULD）」と規範文で促す案。producer 適合検証（schema）で強制されないため、§5 の参照キーが single で欠ける穴が残る。3 層参照キーを構造で保証できないため棄却。
- **specVersion を 2 に上げ spec/v2 を新設する**: 必須フィールド削除・必須化を伴うため規約文面上は v2 相当。しかし ADR-0011 / 0012 は本日採用で consumer も nput も未着手であり、下書きを consume した実装が存在しない。運用未開始の下書きに v2 ディレクトリ分岐を作るのは過剰で、v1 内修正に留める。
