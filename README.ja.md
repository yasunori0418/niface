# niface

*この文書は英語版 [`README.md`](README.md) の日本語訳。仕様・用語の一次参照は英語版とし、両者に差異があれば英語版が優先する。*

**n**-tools **i**nter**face** — Nix ベースのシステム管理ツール群（nput / nboot / nwrap / nherd / nshadow / ncompose）の実行結果を、単一の構造化 JSON 文書（結果エンベロープ）として stdout に出力するための共通規格。

各ツールが単一の責務を持ち、パイプで組み合わさってディストリビューションを形成する UNIX 哲学的なエコシステムにおいて、niface は結果報告の「配管の口径」を定義する。規格に適合しさえすれば、どの言語で書かれたどのツールでもパイプラインに乗れる。各ツールの*入力* JSON の形は本規格ではなく各ツールが公開・管理する（ADR-0014 参照）。

本リポジトリは 2 つの役割を持つ:

1. **niface 規格**: 仕様・schema・適合テストデータ・参照実装
2. **エコシステム中央ドキュメント**: n プレフィックスツール群の全体構想・ディストリビューション化プラン（[docs/ecosystem/](docs/ecosystem/)）

> **ステータス: draft (specVersion 1)** — 初期論点は確定済み。実装ツールからのフィードバックにより互換変更（フィールド追加）を積む段階。

## ドキュメント

規格:

| 文書 | 内容 |
|------|------|
| [CONTEXT.md](CONTEXT.md) | 用語集（規格の語彙の正名） |
| [docs/concept.md](docs/concept.md) | なぜ存在するか・原則・しないこと |
| [docs/design.md](docs/design.md) | 設計概要と ADR 索引 |
| [docs/adr/](docs/adr/) | 設計判断の個別記録（背景・根拠・棄却案） |
| [spec/v1/spec.md](spec/v1/spec.md) | **規範**（RFC 2119 語彙による仕様） |

エコシステム:

| 文書 | 内容 |
|------|------|
| [docs/ecosystem/overview.md](docs/ecosystem/overview.md) | 全体索引: 北極星・ツール一覧・設計原則 |
| [docs/ecosystem/distro-plan.md](docs/ecosystem/distro-plan.md) | ディストリビューション構想と各ツールの責務・マイルストーン |

## リポジトリ構成

```
niface/
├── CONTEXT.md        # 用語集（規格の語彙の正名）
├── docs/             # 規格のコンセプト・設計
│   ├── adr/          # 設計判断の個別記録（ADR）
│   ├── agents/       # エージェント向けの参照規約
│   └── ecosystem/    # エコシステム中央ドキュメント（構想・全体索引）
├── spec/v1/          # 規範仕様
├── schema/v1/        # JSON Schema（機械可読の正）
├── testdata/v1/      # 適合テストデータ
│   ├── valid/        #   schema に適合すべきサンプル
│   ├── invalid/      #   schema が拒否すべきサンプル
│   └── id-vectors.json  # identity → 期待 id の対応表（言語間互換の要）
├── go/               # Go 参照実装（エンベロープ型 + id 導出 + 適合検証 + niface-validate CLI）
├── nix/              # Nix 実装（id 導出）
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
  "errors": [],                   // 主体列挙・解決の前段の全体エラーのみ
  "results": [                    // 主体ごとに subjectResult 1 つ（0..N。--all 等で複数になる）
    {
      "subject": { "name": "home" },  // 操作の主体（常に必須）
      "status": "success",
      "startedAt": "...", "finishedAt": "...",
      "result": {
        "items":   [ ... ],       // 処理単位の実行結果（1 件でも failed → error）
        "changes": [ ... ],       // 差分の宣言（reversible を必ず持つ）
        "info": { }               // ツール固有はここだけ（items/changes 内も同様）
      }
    }
  ]
}
```

item id は `sha256(JCS(identity))` で機械的に導出する。ツールが決めるのは identity（`{kind, key}`）の中身だけで、形式・エスケープ・一意性の問題は規格が吸収する。

## 適合

ツールが niface 適合であるとは:

1. **schema 検証**: 出力が `schema/v1/envelope.schema.json` を通過する
2. **id-vectors 通過**: id 導出実装が `testdata/v1/id-vectors.json` の全ベクタで期待値と一致する
3. **standalone**: 入力は stdin の JSON か明示引数のみ。状態・設定の暗黙探索をしない（詳細は spec §8）

自作ツールへの組み込み手順（flake input / go module）は[適合ガイド](docs/guides/conformance.md)を参照。

## 検証

```sh
nix flake check   # id-vectors(Nix 実装) + go test(適合検証 + ベクタ)を実行
```

個別に:

```sh
# Go 実装のベクタテスト + testdata 適合検証
cd go && go test ./...

# 単一エンベロープの適合検証
nix run .#validate -- path/to/envelope.json
```

## 開発環境

テンプレートを `.envrc` にコピーして direnv を許可する。`.envrc` は `dev/` の flake を読み込み（`use flake ./dev`）、開発ツールを入れて mattpocock/skills を nput の project mode で `.claude/skills/` へ配置する。

```sh
cp example.envrc .envrc && direnv allow    # または: nix develop ./dev
```

## 関連

- [nput](https://github.com/yasunori0418/nput) — 配置プリミティブ（エコシステムの起点・active）
- 他ツール（nboot / nwrap / nherd / nshadow / ncompose）は planned。[docs/ecosystem/overview.md](docs/ecosystem/overview.md) を参照
