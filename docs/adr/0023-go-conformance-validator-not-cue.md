# ADR-0023: 適合検証を Go 実装に集約し、単一ソース IDL(CUE)は採用しない

- ステータス: 採用
- 日付: 2026-07-10
- 関連: `spec/v1/spec.md`(§8), `schema/v1/envelope.schema.json`, `scripts/validate.py`, `go/`, ADR-0004, ADR-0014, ADR-0021
- 改訂対象: ADR-0021(§8 の「参照検証は `scripts/validate.py`」を Go 実装へ置換。強制する MUST・schema/lint の分担は不変)

## 背景

ADR-0021 で適合検査は「schema 検証 + validate.py リント + id-vectors」に拡張された。ここで 2 つの問いが出た。

1. **参照 linter の環境管理**: `scripts/validate.py` は `jsonschema`(Python)への依存を持つ「規格のリファレンス linter」になった。依存は Nix(`flake.nix` checks・`dev/flake.nix` devShell)が既に供給しているが、Python を第一級の検証器として持つなら管理を明示したい。
2. **スキーマ駆動化**: 現状は `schema/v1/envelope.schema.json`(正)と `go/types.go`(手書き)の二重管理。gRPC のように単一ソースから型・検証・データ生成を導けないか。

2 について CUE(単一ソース IDL 候補)を spike で実測した(cue v0.17.0、現 schema.json を import して 4 測定)。

- **(a) CUE→JSON Schema 書き出し**: if/then 整合は残るが、open な `info` が `const:{}`(空限定)に誤変換・`format:date-time` 脱落・213→830 行に膨張。
- **(b) CUE→Go 型(`cue exp gengotypes`)**: `Item` が if/then で型生成破綻(`any` + TODO)、generic `info` が `map[string]any` 化、enum の定数消失。実験機能。**現 types.go より質が低い**。
- **(c) `cue vet`**: valid 14/14 受理・invalid は schema レベル 12/12 拒否で JSON Schema と判定一致。
- **(d) 一意性・参照整合**: keyed-struct トリックで subject.name/item.id 一意・itemId 参照整合を CUE で表現でき、lint-invalid 3 件を正しく reject。CUE は schema+lint を単一ソースに統合できる。

## 決定

- **正は JSON Schema のまま。CUE は採用しない**(ADR-0014 の領分限定・「正は JSON Schema」を維持)。
- **適合検証の参照実装を Python から Go へ移す**。`santhosh-tekuri/jsonschema`(Draft 2020-12)で `schema/v1/envelope.schema.json` を検証し、ADR-0021 のリント 3 検査(`changes[].itemId` 参照整合・`subject.name` 一意・`item.id` 一意)を Go で実装する。強制する MUST・schema/lint の分担は ADR-0021 のまま不変。
- testdata harness は Go テストに載せ替え(`go test ./...` が適合検証を兼ねる)、producer が自分の出力 1 件を検査できる CLI(`niface-validate`)も提供する。`scripts/validate.py` と flake の python check は撤去する。
- id 導出の参照実装(`go/` の型・id パッケージ)は外部依存ゼロを保ち、検証器の依存(`santhosh-tekuri`)は検証パッケージ側に閉じる。

## 根拠

- 求めていた「型情報もデータもスキーマから(gRPC 的)」は **CUE では綺麗に実現しない**。gengotypes は if/then で退行し(b)、公開契約は JSON Schema 維持が要る(a の膨張・誤変換)。CUE の唯一の実質的勝ちは「検証の単一ソース化」(d)だが、**niface は nput/nboot 等が消費する共有スペック**で、検証成果物の lingua franca は JSON Schema。CUE は niche で、内部ソースにしても公開 JSON Schema は lint MUST を運べず下流に部分検証が残る。新言語導入の対価に見合わない。
- Go はこの repo の既存の参照言語(id 導出・testdata decode)。適合検証を Go に寄せると、Python の環境管理問題を「別の env manager(uv)を足す」ではなく「Python を無くす」で根治でき、参照スタックが Go + Nix に一本化される。
- santhosh-tekuri は本物の Draft 2020-12 validator(自作の JSON Schema 再実装ではない)。schema を正として読む点は Python 版と同じで、適合の意味論は変わらない。
- Go 実装は producer 向け単一文書 CLI を自然に生む(ADR-0021 で保留した単一文書モード)。

## 影響

- `go/`: 検証パッケージ(schema 検証 + lint 3 検査)、適合 harness の Go テスト化、`niface-validate` CLI を追加。go.mod に `santhosh-tekuri/jsonschema` を追加(id 導出パッケージからは import しない)。
- `scripts/validate.py` を撤去。`flake.nix` の `checks.schema`(python derivation)を撤去し `checks.go` が適合検証を兼ねる。`dev/flake.nix` から `python3` / `python3Packages.jsonschema` を撤去。
- `spec/v1/spec.md` §8 と `CONTEXT.md` の「適合」項: 参照検証を「`scripts/validate.py`」から Go 実装(`go/` の検証器・`niface-validate`)へ差し替え。強制対象 MUST の記述は不変。
- `docs/design.md`: 適合戦略・ディレクトリ注釈(`scripts/` 撤去)・ADR 索引。
- `CLAUDE.md` の検証コマンド(`python3 scripts/validate.py ...` を Go の口へ)。
- schema/spec/testdata の規範は無改訂(検証の**実装**が変わるだけ)。id 導出系も無改訂。

## 棄却した代替案

- **CUE を単一ソース IDL に採用**: spike が示す通り、型 codegen は退行(gengotypes)、公開 JSON Schema 生成は要手当て(explicitopen/format/膨張)、共有スペックへの新言語導入コストが高い。得られる勝ち(検証の単一ソース化)が対価に見合わない。再検討条件: エコシステムが CUE 採用を許容 / gengotypes が安定し if/then を扱える。
- **Python を uv(+uv2nix)で管理**: cryoflow 流。1 スクリプト・1 依存に対し pyproject/uv.lock/uv2nix + flake input 3 個は不釣り合いで、Nix が既に供給する再現性を uv で二重管理することになる。Python 面が増える計画も無い。
- **JSON Schema → Go 型を codegen して types.go を生成**: 手作りのジェネリクス `info` 設計(`Item[T]`/`Envelope[...]`)を再現できず `map[string]any` に潰れ、if/then 整合は型で表現できず lint は残る。現 types.go の方が設計が良く、二重管理解消の利より設計退行の害が大きい。
- **Python 参照 linter を維持**: 動いてはいるが「reference linter の環境管理」問題が残り、参照スタックが Go/Nix/Python の 3 言語に散る。Go 集約で Python を消す方が総管理量が減る。
