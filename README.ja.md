# niface

*この文書は英語版 [`README.md`](README.md) の日本語訳。仕様・用語の一次参照は英語版とし、両者に差異があれば英語版が優先する。*

**n**-tools **i**nter**face** — Nix ベースのシステム管理ツール群（nput / nboot / nwrap / nherd / nshadow / ncompose）が stdout / stdin を通じて会話するための共通 JSON 規格。

各ツールが単一の責務を持ち、パイプで組み合わさってディストリビューションを形成する UNIX 哲学的なエコシステムにおいて、niface は「配管の口径」を定義する。規格に適合しさえすれば、どの言語で書かれたどのツールでもパイプラインに乗れる。

本リポジトリは 2 つの役割を持つ:

1. **niface 規格**: 仕様・schema・適合テストデータ・参照実装
2. **エコシステム中央ドキュメント**: n プレフィックスツール群の全体構想・ディストリビューション化プラン・命名記録（[docs/ecosystem/](docs/ecosystem/)）

> **ステータス: draft (specVersion 1)** — 初期論点は確定済み。実装ツールからのフィードバックにより互換変更（フィールド追加）を積む段階。

## ドキュメント

規格:

| 文書 | 内容 |
|------|------|
| [docs/concept.md](docs/concept.md) | なぜ存在するか・原則・しないこと |
| [docs/design.md](docs/design.md) | 全設計判断（19 項目）とその根拠 |
| [spec/v1/spec.md](spec/v1/spec.md) | **規範**（RFC 2119 語彙による仕様） |

エコシステム:

| 文書 | 内容 |
|------|------|
| [docs/ecosystem/overview.md](docs/ecosystem/overview.md) | 全体索引: 北極星・ツール一覧・設計原則 |
| [docs/ecosystem/distro-plan.md](docs/ecosystem/distro-plan.md) | ディストリビューション構想と各ツールの責務・マイルストーン |
| [docs/ecosystem/naming.md](docs/ecosystem/naming.md) | 命名の決定記録と検討経緯 |

## リポジトリ構成

```
niface/
├── docs/             # 規格のコンセプト・設計
│   └── ecosystem/    # エコシステム中央ドキュメント（構想・命名・全体索引）
├── spec/v1/          # 規範仕様
├── schema/v1/        # JSON Schema（機械可読の正）
├── testdata/v1/      # 適合テストデータ
│   ├── valid/        #   schema に適合すべきサンプル
│   ├── invalid/      #   schema が拒否すべきサンプル
│   └── id-vectors.json  # identity → 期待 id の対応表（言語間互換の要）
├── go/               # Go 参照実装（エンベロープ型 + id 導出）
├── nix/              # Nix 実装（id 導出）
├── scripts/          # 検証スクリプト
├── dev/              # 開発環境（devShell + mattpocock/skills 配置）
└── flake.nix
```

## 30 秒でわかる規格

全ツールは stdout に単一の JSON 文書（エンベロープ）だけを出力する:

```jsonc
{
  "specVersion": 1,
  "tool": { "name": "nput", "version": "0.9.0" },
  "command": "apply",
  "status": "success",            // "success" | "error"（2 値のみ）
  "dryRun": false,                // plan でもスキーマは apply と同一
  "startedAt": "...", "finishedAt": "...",
  "errors": [],                   // item に紐づかない全体エラーのみ
  "result": {
    "items":   [ ... ],           // 処理単位の実行結果（1 件でも failed → error）
    "changes": [ ... ],           // 差分の宣言（reversible を必ず持つ）
    "info": { }                   // ツール固有はここだけ（items/changes 内も同様）
  }
}
```

item id は `sha256(JCS(identity))` で機械的に導出する。ツールが決めるのは identity（`{kind, key}`）の中身だけで、形式・エスケープ・一意性の問題は規格が吸収する。

## 適合

ツールが niface 適合であるとは:

1. **schema 検証**: 出力が `schema/v1/envelope.schema.json` を通過する
2. **id-vectors 通過**: id 導出実装が `testdata/v1/id-vectors.json` の全ベクタで期待値と一致する
3. **standalone**: 入力は stdin の JSON か明示引数のみ。状態・設定の暗黙探索をしない（詳細は spec §8）

## 検証

```sh
nix flake check   # id-vectors(Nix 実装) + schema(testdata) + go test を実行
```

個別に:

```sh
# schema 検証
python3 scripts/validate.py schema/v1/envelope.schema.json testdata/v1

# Go 実装のベクタテスト
cd go && go test ./...
```

> **注意**: Go / Nix 実装は Go・Nix ツールチェーンのない環境で作成されたため未実行です。初回 push 前に `nix flake check` の通過を必ず確認してください。id-vectors.json 自体は Python 実装（JCS サブセット）で計算済みの実値です。

## 開発環境

テンプレートを `.envrc` にコピーして direnv を許可する。`.envrc` は `dev/` の flake を読み込み（`use flake ./dev`）、開発ツールを入れて mattpocock/skills を nput の project mode で `.claude/skills/` へ配置する。

```sh
cp example.envrc .envrc && direnv allow    # または: nix develop ./dev
```

## 関連

- [nput](https://github.com/yasunori0418/nput) — 配置プリミティブ（エコシステムの起点・active）
- 他ツール（nboot / nwrap / nherd / nshadow / ncompose）は planned。[docs/ecosystem/overview.md](docs/ecosystem/overview.md) を参照
