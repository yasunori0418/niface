# ADR-0026: リリースタグ運用(v1.N.P / go/v1.N.P)と v1 安定化基準を定める

- ステータス: 採用
- 日付: 2026-07-11
- 関連: `docs/design.md`, `docs/ecosystem/overview.md`, `go/`(サブディレクトリ module `github.com/yasunori0418/niface/go`), `flake.nix`, ADR-0010

## 背景

`docs/design.md` は「互換変更は v1 内 + git タグ」と定めるが、タグは 1 本も打たれておらず、命名規則・打つタイミング・「draft → 安定」宣言の基準が未定義だった。互換追加(ADR-0015 / 0018 / 0019 等)は既に複数回あるのに、消費側が「どの時点の v1 か」を参照する手段が commit hash しかない。

Go module はサブディレクトリ module(`github.com/yasunori0418/niface/go`)であり、`go/vX.Y.Z` 形式(prefix 付き)のタグが無いと pseudo-version 参照しかできない。prefix 無しの root タグは Go ツールチェーンから見えないため、規格スナップショット用のタグと Go module 用のタグを分けて対で打つ必要がある。

ADR-0010 はバージョニング規約を定めたが、その射程は `specVersion` フィールド(非互換世代を表す整数)であり、リリース単位のスナップショット識別子(タグ)は空白のままだった。本 ADR はその空白を埋める新規決定で、ADR-0010 の `specVersion` 運用は改訂しない。

## 決定

- **タグ体系**: 同一コミットに 2 本のタグを対で打つ。
  - `v1.N.P` — 規格スナップショット(flake input / 人間用)。`N` = 互換追加(spec / schema / testdata に触れる変更)ごと、`P` = 規格に影響しない修正(実装 fix・testdata 追補)。
  - `go/v1.N.P` — 同一コミットの Go module 用ミラー。サブディレクトリ module のため `go/` prefix を必須とする。
- **打つタイミング**: 互換変更 PR のマージごとに、上記 2 本を対で打つ。
- **v1 stable 宣言基準**: **2 ツール適合**(nput + 次ツール、M2 の nboot 想定)を宣言条件とする。単一実装(nput)の写像でないことを担保する needs 駆動の基準と同型。宣言時に README / overview の draft 表記を落とし、spec 英語版の要否を再検討する。
- **ADR-0010 との関係**: `v1.N.P` は semver 形式だが、ADR-0010 が棄却した「semver 文字列」は `specVersion` フィールドの粒度に関する判断であり、リリースタグとはレイヤが異なる。両者は矛盾しない(specVersion = 非互換世代の整数、タグ = リリーススナップショット識別子)。
- **実装バージョンとの独立**: `flake.nix` の `packages.validate` の `version = "0.1.0"` は Go 参照実装の実装バージョンであり、規格タグ `v1.N.P` とは独立して増減する。本 ADR は `flake.nix` を改訂しない。
- **初回タグ**: `v1.0.0` + `go/v1.0.0` を初回リリースとして打つ。ただしタグ作成は本 issue の worktree 内では行わず、並行 issue が全てマージされた後の main の HEAD に対して行う(worktree 内の作業は ADR + docs 同期のみ)。

## 根拠

- 2 本のタグを対で打つことで、人間 / flake input は `v1.N.P` を、Go ツールチェーンは `go/v1.N.P` を、同一コミットに対してそれぞれ自然に参照できる。サブディレクトリ module の制約(root タグが見えない)を、規格側のタグ命名を歪めずに吸収できる。
- `N` / `P` を「spec / schema / testdata に触れるか」で分けると、消費側は `N` の増分だけで互換追加の有無を判断でき、`P`(実装 fix)を無視できる。specVersion(非互換のみ増える)との二層で、粒度の異なる変化を別々に表現できる。
- stable 宣言を「2 ツール適合」に結び付けると、規格が単一実装の写像でないことを客観的に担保できる。2 つ目の適合ツールが出るまで draft を維持し、規格の一般性が実証されてから draft 表記を落とす。
- ADR-0010 の semver 棄却との非矛盾を明記することで、`v1.N.P` タグが ADR-0010 に反すると誤読されるのを防ぐ。棄却は specVersion の粒度に閉じており、タグは別レイヤ。
- 実装バージョン(`packages.validate` の `0.1.0`)と規格タグの独立を明記することで、実装の bump を規格タグの bump と誤認する等の混同を防ぐ。

## 影響

- `docs/design.md`: ADR 索引に本 ADR を追加。「バージョニングとリリースタグ」節を新設し、specVersion(非互換世代)/ リリースタグ(`v1.N.P`・`go/v1.N.P`)/ stable 宣言基準を記述。
- `docs/ecosystem/overview.md`: niface の draft → stable 基準(2 ツール適合)とリリースタグ体系を追記。
- タグ `v1.0.0` / `go/v1.0.0` は全レーンマージ後の main HEAD に対して親セッションが打つ(本 worktree では打たない)。
- `spec/` / `schema/` / `testdata/` / `go/` / `nix/` / `flake.nix` / `README` は無改訂。`specVersion` は 1 のまま(ADR-0010 の運用は不変)。

## 棄却した代替案

- **単一のタグ(`v1.N.P` のみ)で Go module も参照させる**: サブディレクトリ module のため root タグは Go ツールチェーンから見えず、pseudo-version 参照しか残らない。`go/` prefix 付きタグを対で打つ方が Go 側の参照が素直。
- **`specVersion` を semver 文字列化してタグと一体運用する**: ADR-0010 が棄却済み。非互換世代の粒度に semver は過剰で、整数 + ディレクトリ(`v1/`)の単純さを失う。タグ(スナップショット)と specVersion(世代)は別レイヤに保つ。
- **タグを打たず commit hash 参照のままにする**: 消費側が「どの時点の v1 か」を hash でしか指せず、互換追加の履歴が追えない。タグ(特に `N` の増分)で互換世代を可視化する方が参照コストが低い。
- **1 ツール(nput)適合で stable 宣言する**: 規格が単一実装の写像に留まり、一般性が未実証のまま安定を名乗ることになる。2 ツール適合を待つ方が規格の契約としての信頼性が高い。
