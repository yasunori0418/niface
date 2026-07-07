# ADR-0019: エンベロープと subjectResult に warnings[] を追加し warning 型を error から分離する

- ステータス: 採用
- 日付: 2026-07-07
- 関連: `spec/v1/spec.md`（§2, §3, §6）, `schema/v1/envelope.schema.json`, `go/types.go`, ADR-0005, ADR-0006, ADR-0013

## 背景

警告の置き場は現状 `item.warnings` のみで、主体に紐づくが item に紐づかない警告（例: その主体の profile の移行注意）や、実行全体に関わる警告（例: 非推奨フラグの使用）を置ける場所が無い。エラーは全体 / 主体 / item の 3 階層に責務分離済み（→ ADR-0006）なのに、警告だけ item 階層に閉じている非対称がある。また schema 上 `warnings[]` は `$defs/error` を共用しており、code pattern が `^[EW]_` のため、error に `W_` code・warning に `E_` code を入れた出力も schema を通過してしまう。二層命名（→ ADR-0005）の `E_` / `W_` 区別は意図された意味論だが、producer 適合検証がそれを強制していない。（tracking #8 / #10）

## 決定

- エンベロープ直下と `subjectResult` に optional な `warnings[]` を追加する（互換追加・specVersion 据え置き）。置き場の責務は `errors` と同型とする: トップレベル = 実行全体（主体に紐づかない）、subjectResult = 主体（その主体の item に紐づかない）、item = 処理単位。
- `$defs/warning` を新設し、code pattern を `^W_[A-Z][A-Z0-9_]*$` に限定する。`item.warnings` を含む全 `warnings[]` は `$defs/warning` を参照する。
- `$defs/error` の code pattern を `^E_[A-Z][A-Z0-9_]*$` に限定する。
- pattern 限定は意図されていた意味論の強制であり、未リリース下書き段階の修正として specVersion は 1 のまま据え置く（ADR-0013 と同じ扱い）。
- `go/types.go` は `Error` と同形の `Warning` 型を新設し、`Envelope.Warnings` / `SubjectResult.Warnings` を追加、`Item.Warnings` を `[]Error` から `[]Warning` へ変更する（go module は未消費のため破壊許容）。

## 根拠

- 警告の階層構造をエラーと同型にすることで、置き場の判断基準（どの階層に紐づくか）を `errors[]` と共通化でき、新しい規約を増やさない。warning は status に影響しない（1 件でも failed → error の集約規則は不変）点だけが errors と異なる。
- 型分離により「error の code は `E_`、warning の code は `W_`」という二層命名の意味論が producer 適合検証で機械的に強制され、取り違え出力を schema の段階で検出できる。
- validation 強化（従来 valid だった出力を invalid にする）は文面上非互換だが、`W_` code の error / `E_` code の warning は意図に反する出力であり、consumer / producer とも未リリースの下書き段階で正すのが最も安い（ADR-0013 の先例に従う）。

## 影響

- `spec/v1/spec.md` §2（トップレベル / SubjectResult の jsonc 例と規範文）。§6 の二層命名は無改訂（意味論は元から `E_` / `W_` 区別）。
- `schema/v1/envelope.schema.json`: トップレベル / subjectResult の `properties.warnings`、`$defs/warning` 新設、`$defs/error` の pattern 限定、`item.warnings` の参照先を `$defs/warning` へ変更。
- `go/types.go`: `Warning` 型、`Envelope.Warnings` / `SubjectResult.Warnings` 追加、`Item.Warnings` の型変更。
- `testdata/v1/`: valid に上位 warnings 付きベクタ、invalid に `error-code-with-w-prefix.json` / `warning-code-with-e-prefix.json` を追加。
- `CONTEXT.md` の warning / エラーコード項と `docs/design.md` の索引。
- id 導出系（`testdata/v1/id-vectors.json`・`go/id.go`・`nix/lib.nix`）は無改訂。

## 棄却した代替案

- **$defs/error 共用の継続**: warnings[] の置き場だけ増やし型は error と共用し続ける案。`W_` / `E_` の取り違えを schema が検出できないままで、意味論の強制という #10 の動機を満たさない。
- **warning 型分離を #15（status⇔errors 整合の schema 表現）送り**: 上位 warnings[] を `$defs/error` 参照で先に入れると、後から参照先を差し替える二度手間になり、その間に取り違え出力が testdata・実装に紛れる余地が生じる。追加と同時に分離する方が安い。
- **specVersion increment**: pattern 限定は validation 強化だが、invalid 化されるのは意図に反する出力のみで、下書きを consume した実装は存在しない。v2 ディレクトリ分岐は過剰（ADR-0013 と同判断）。
