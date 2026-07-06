# ADR-0011: エンベロープを常に results[] にし mode で単一/複数主体を表す

- ステータス: 採用
- 日付: 2026-07-07
- 関連: `spec/v1/spec.md`（§2, §5）, `schema/v1/envelope.schema.json`, ADR-0001, ADR-0002, ADR-0006, ADR-0007
- 改訂対象: ADR-0001（ペイロードの容器を `result` 単数から `results[]` に一般化）, ADR-0002（status 集約を results[] 全体へ拡張）, ADR-0006（全体エラーの置き場に解決済み主体の subjectResult.errors 層を追加）

## 背景

現行エンベロープは「1 実行 = 1 論理操作 = 1 result」を暗黙前提とし、単一主体しか表せない。nput `apply --all`（1 実行 = N 個の独立 config への配置）や、複数対象を束ねる将来ツール（nherd 等）はこれを表現できない。stdout の単一文書規律（§1・NDJSON 禁止）を保ったまま、複数主体の結果を 1 文書に並べる手段が要る。

## 決定

- トップレベルは single / batch を問わず常に `results[]` を持つ。要素は `SubjectResult`（`{ subject?, status, startedAt, finishedAt, errors?, result }`）とする。
- `mode`（`single` / `batch`）を唯一の判別子とし、必須にする。`mode` は**起動の性質**で決まり結果の件数に依らない（`apply --all` は対象 0 / 1 個でも `batch`）。
- `mode: "single"` の `results` は高々 1 要素（0 または 1）、`mode: "batch"` は 0 以上。対象 0 件の `batch`（`results: []`）は `status: "success"`（vacuous truth）。
- `status` は集約: `results` のいずれかの主体が `error`、または主体列挙・解決の前段で全体エラーがあれば `error`。
- `tool` / `command` / `mode` / `dryRun` / `specVersion` はトップレベルにのみ置く。これにより batch は単一 tool・単一 command であることが構造的に保証される。
- トップレベル `errors[]` は主体の列挙・解決の前段で起きる全体エラー（入力 parse 失敗・主体列挙自体の失敗等）のみ。解決済み主体に紐づくエラーは `subjectResult.errors` に置く（→ ADR-0006 の拡張）。
- `changes[].itemId` は同一 `result`（同一 subjectResult）内で解決する。

## 根拠

- 容器を `results[]` の 1 つに統一することで、schema と消費側の分岐を単一化し、single / batch で 2 種類の容器を二重管理することを避ける。
- `mode` を唯一の判別子にし、`result` / `results` の有無のような second discriminator を持たない。判別子は 1 つに閉じる。
- `mode` を起動の性質で決めることで、消費側はコマンド形式から結果の形を予測でき、同じ起動が常に同型になる。
- `tool` / `command` をトップレベルのみに置くことで、batch が単一ツール・単一コマンドであることが構造で担保され、複数ツールの束ね（ncompose の領分）と混ざらない。

## 影響

- `spec/v1/spec.md` §2（`results[]` / `mode` / SubjectResult）・§5（itemId のスコープ）。
- `schema/v1/envelope.schema.json`: 常時 `results[]`、`$defs/subjectResult`、`allOf` の `if/then`（`single` は `results` maxItems 1 / `batch` は各要素に `subject` 必須）。
- `testdata/v1/`: 既存 valid / invalid を `results` 形へ改変し、batch ベクタ（`batch-success` / `batch-partial-failure` / `batch-empty` / `batch-enumeration-error`）と mode 系 invalid を追加。
- `go/types.go`: `Envelope.Result`（単数）を `Envelope.Results []SubjectResult` に置換。

## 棄却した代替案

- **single は `result` 単数・batch は `results[]` の 2 容器（mode で result XOR results）**: 多数派の single が今の `.result` 形のまま簡潔になる利点はあるが、schema が result 枝と results 枝の 2 分岐を抱え容器の二重管理を招く。一様性と単一 schema を優先して棄却。
- **`result` / `results` の有無を判別子にする（`mode` を持たない）**: 判別子が暗黙になる。`mode` という明示子の方が起動の性質との対応が明快で、consumer の分岐も安定する。
- **per-subject 要素を完全な sub-envelope にする（specVersion / tool / command / mode を各要素に持たせ単独再流通可能に）**: single が要素にエンベロープ全体を二重化して冗長。単独再流通（ncompose の切り出し）は現時点で投機的価値であり、軽量な SubjectResult を優先して棄却。
- **read-only 列挙（list-generations の世代行・gitignore の行）を item に拡張する**: 世代・列挙は read-only で副作用がなく `changes` も cross-run 照合も持たないため、id 導出（§5）の機構が無用になる。列挙はツール固有インベントリとして `result.info` に置けば足りる（→ ADR-0007）。item は「処理単位の実行結果の記録」の定義を保つ。これに伴い §5 の「key に世代番号を含めてはならない」は書き換え不要となる（世代番号がどの key にも入らないため衝突しない）。
