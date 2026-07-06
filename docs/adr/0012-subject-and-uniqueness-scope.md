# ADR-0012: subject を弱い識別子として導入し §5 一意性を参照キー 3 層に再定義する

- ステータス: 採用
- 日付: 2026-07-07
- 関連: `spec/v1/spec.md`（§2, §5）, `schema/v1/envelope.schema.json`, ADR-0004, ADR-0011
- 改訂対象: ADR-0004（id 一意性の範囲を「ツール内」から `(tool.name, subject, id)` の 3 層参照キーに再定義）

## 背景

複数主体実行（→ ADR-0011）を入れると、「どの主体の結果か」を名指す手段が要る。加えて、操作の主体（例: nput の config 名）が引数に埋もれて出力に残らないという欠落が single モードにもある。一方で id 導出（`sha256(JCS(identity))`）に主体を混ぜると、identity が `{kind, key}` の 2 層から 3 層に変わり id-vectors・go / nix 実装の全面改訂（= 非互換）を招く。主体の表現と id 機構を分離する必要がある。

## 決定

- `subject`（`{ name }`）を「操作の主体を名指す弱い識別子」として導入する。`batch` の各 `results[]` 要素では必須、`single` では任意（トップレベルの主体名の欠落を埋める）。
- `subject` は id 導出に**関与しない**。id の入力は従来通り identity（`{kind, key}`）のみ。
- `subject.name` は 1 エンベロープ内で一意でなければならない（producer MUST）。world-wide 一意は求めない。
- §5 の一意性を参照キーの規約として 3 層に再定義する: id は 1 主体（`subject`）の `result` 内で一意、`subject` が `tool` 内で修飾、`tool.name` が複数ツール集約時に修飾。消費側は `(tool.name, subject, id)` の 3 つ組で参照を解決する。
- 同一 identity の id は主体を跨いで衝突してよい（同じ論理対象を表すため当然）。

## 根拠

- `subject` を id 機構から切り離すことで、identity は `{kind, key}` の 2 層のまま保て、id-vectors・go / nix の id 導出は無改訂で済む。specVersion 据え置きの前提を崩さない。
- 主体の同一性は「1 エンベロープ内で一意な name」という軽い契約で足りる。強い識別子（派生 id）は現時点で需要が無く、必要になったとき optional に足せる。
- 参照解決を 3 層キーで明文化することで、複数主体・複数ツールの集約時に id 衝突を主体・ツールで正しく解きほぐせる。

## 影響

- `spec/v1/spec.md` §2（SubjectResult の `subject`）・§5（参照キー 3 層）。
- `schema/v1/envelope.schema.json`: `$defs/subject`、SubjectResult の `subject`、`batch` で `subject` 必須の `if/then`。
- `go/types.go`: `Subject` 型と `SubjectResult.Subject`。
- id 導出系（`testdata/v1/id-vectors.json`・`go/id.go`・`nix/lib.nix`）は無改訂（subject 非関与のため）。

## 棄却した代替案

- **subject を id 導出に組み込む（identity を 3 層化）**: id が主体内で真に一意な値になるが、id-vectors 全面改訂・go / nix 実装改訂・既存 testdata 非互換 = specVersion 2 を招く。運用未開始でも、参照キー規約で解ける問題に非互換変更は過剰。
- **subject に optional な派生 id（`sha256(JCS(subject identity))`）を持たせる**: item との対称性は得られるが、世代跨ぎの安定 subject id は現時点で需要が無い。弱い name 識別子に留め、必要時に足す余地だけ残す。
- **subject を持たず主体を info に押し込む**: 主体は規格が型安全に扱うべき第一級の概念であり、ツール固有の `info` に隠すと consumer が主体で結果を引けない。
