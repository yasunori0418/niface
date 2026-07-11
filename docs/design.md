# niface: 設計

設計判断は 1 件 = 1 ADR として `docs/adr/` に個別記録する。本書はその概要と索引。規範的な定義は `spec/v1/spec.md`、用語は `CONTEXT.md` を正とする。

## ADR 索引

| ADR | 論点 |
|-----|------|
| [0001](adr/0001-nested-result-envelope.md) | ペイロードを `result` 配下に入れ子にする |
| [0002](adr/0002-status-two-values-and-partial-failure.md) | `status` を 2 値にし部分失敗を error へ集約する |
| [0003](adr/0003-items-changes-separation.md) | items（実行結果）と changes（差分宣言）を分離する |
| [0004](adr/0004-id-derivation-jcs-sha256.md) | item id を `sha256(JCS(identity))` で機械導出し安定性を保証する |
| [0005](adr/0005-error-code-two-layer-registry.md) | エラーコードを二層命名にし共通レジストリを凍結する |
| [0006](adr/0006-error-responsibility-separation.md) | item 起因エラーと全体エラーの置き場所を分ける |
| [0007](adr/0007-info-isolation.md) | ツール固有フィールドを `info` 配下に隔離する |
| [0008](adr/0008-posix-exit-code.md) | exit code は POSIX 慣行のまま status と連動させる |
| [0009](adr/0009-dryrun-schema-parity.md) | dry-run は apply と同一スキーマで出力する |
| [0010](adr/0010-camelcase-versioning-extension.md) | 命名・バージョニング・拡張ポリシーを定める |
| [0011](adr/0011-batch-envelope.md) | エンベロープを常に results[] にし mode で単一/複数主体を表す |
| [0012](adr/0012-subject-and-uniqueness-scope.md) | subject を弱い識別子として導入し §5 一意性を参照キー 3 層に再定義する |
| [0013](adr/0013-remove-mode-and-require-subject.md) | mode 判別子を廃止し subject を常時必須にする |
| [0014](adr/0014-scope-result-envelope-only.md) | 規格の領分を結果エンベロープに限定し世代表現を独立スキーマにしない |
| [0015](adr/0015-generation-observation-slot.md) | subjectResult に観測型の世代遷移スロット generation を追加する |
| [0016](adr/0016-changes-completeness-on-partial-failure.md) | 部分失敗時も適用済み changes を必ず出力する |
| [0017](adr/0017-best-effort-envelope-on-interruption.md) | 中断時は best-effort でエンベロープを出力する |
| [0018](adr/0018-envelope-info-slot.md) | エンベロープ直下にツール固有 info の置き場を追加する |
| [0019](adr/0019-envelope-warnings-and-warning-type.md) | 上位 warnings[] を追加し warning 型を error から分離する |
| [0020](adr/0020-skipped-limited-to-preceding-failure.md) | skipped を前段の失敗による未実行に限定する |
| [0021](adr/0021-conformance-lint-for-schema-inexpressible-musts.md) | schema で表現しきれない MUST を適合検査（schema 強化 + リント）に取り込む |
| [0022](adr/0022-sensitive-info-masking-guidance.md) | 機微情報を info / detail に生で載せない SHOULD を記録する |
| [0023](adr/0023-go-conformance-validator-not-cue.md) | 適合検証を Go 実装に集約し単一ソース IDL（CUE）は採用しない |
| [0024](adr/0024-identity-value-domain.md) | item id の identity 値域を確定し域外を実装拒否とする |

ADR の書式と改訂注記の運用は [`docs/adr/README.md`](adr/README.md) を参照。

## リポジトリ構成と適合戦略

```
niface/
├── CONTEXT.md        # 用語集（glossary。規格の語彙の正名）
├── docs/
│   ├── concept.md    # コンセプト・原則・しないこと
│   ├── design.md     # 本書（設計概要 + ADR 索引）
│   ├── adr/          # 設計判断の個別記録（ADR）
│   ├── agents/       # エージェント向けの参照規約
│   ├── ecosystem/    # エコシステム中央ドキュメント
│   └── guides/       # 実装者向けの指針（規範ではない）
├── spec/v1/          # 規範仕様（RFC 2119 語彙）
├── schema/v1/        # JSON Schema（機械可読の正）
├── testdata/v1/      # valid / invalid サンプル + id-vectors.json
├── go/               # Go 参照実装（型 + id 導出 + 適合検証 + CLI niface-validate）
├── nix/              # Nix lib（id 導出・ベクタ検証）
├── dev/              # 開発環境（devShell + mattpocock/skills 配置）
└── flake.nix         # checks: id-vectors / go、apps: validate
```

- **正は JSON Schema**。文書とコードはそれに従う。
- **id-vectors.json が最重要資産**: identity → 期待 id の対応表。JCS の罠（非 ASCII の値・数値表記・ネスト）を突くベクタと、域外を突く `rejected`（非整数表記・範囲外整数・非 ASCII メンバー名）を含め、全言語実装が CI でこれを通すことで導出の互換を証明する（→ ADR-0004, ADR-0024）。
- 参照方法は二経路: Go ツールは go module で型を共有（コンパイル時）、各ツールの flake が本リポジトリを input に取り schema 検証 + MUST リント検査 + id-vectors 適合を checks で回す（CI 時）。schema で表現しきれない MUST（status 整合・itemId 参照整合・一意性）は `go/conformance` のリント層が担い、単一文書検証は CLI `niface-validate`（`nix run .#validate`）が提供する（→ ADR-0021, ADR-0023）。
- バージョニング: specVersion 整数とディレクトリ（v1/）を一致させ、互換変更は v1 内 + git タグ、非互換変更のみ v2/ を新設する（→ ADR-0010）。
- ツール固有の info schema は各ツールのリポジトリが管理する（規格側で抱えない・→ ADR-0007）。
