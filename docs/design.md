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
│   └── ecosystem/    # エコシステム中央ドキュメント
├── spec/v1/          # 規範仕様（RFC 2119 語彙）
├── schema/v1/        # JSON Schema（機械可読の正）
├── testdata/v1/      # valid / invalid サンプル + id-vectors.json
├── go/               # Go 参照実装（型 + id 導出 + ベクタテスト）
├── nix/              # Nix lib（id 導出・ベクタ検証）
├── scripts/          # schema 検証スクリプト（flake checks から使用）
├── dev/              # 開発環境（devShell + mattpocock/skills 配置）
└── flake.nix         # checks: id-vectors / schema / go
```

- **正は JSON Schema**。文書とコードはそれに従う。
- **id-vectors.json が最重要資産**: identity → 期待 id の対応表。JCS の罠（非 ASCII キー・数値表現・ネスト）を突くベクタを含め、全言語実装が CI でこれを通すことで導出の互換を証明する（→ ADR-0004）。
- 参照方法は二経路: Go ツールは go module で型を共有（コンパイル時）、各ツールの flake が本リポジトリを input に取り schema 検証 + id-vectors 適合を checks で回す（CI 時）。
- バージョニング: specVersion 整数とディレクトリ（v1/）を一致させ、互換変更は v1 内 + git タグ、非互換変更のみ v2/ を新設する（→ ADR-0010）。
- ツール固有の info schema は各ツールのリポジトリが管理する（規格側で抱えない・→ ADR-0007）。
