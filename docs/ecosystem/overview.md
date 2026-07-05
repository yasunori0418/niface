# n-tools エコシステム: オーバービュー

n プレフィックスのツール群と niface 規格からなるエコシステムの中央索引。北極星は、NixOS とは異なる「Nix 版 Arch / Gentoo」——最小のコアとユーザーによる組み立てを旨とし、優れた Handbook を持つディストリビューション。

NixOS がモジュールシステムという単一の大きな抽象で全体を統合するのに対し、このエコシステムは**単一責務のツールを niface 規格のパイプで合成する**ことで同じ問題を解く。

## 全体構成

```
build(Nix eval/build)
  → nput apply      # config / env の配置
  → nshadow         # ユーザー/グループ差分調停
  → nwrap           # setuid ラッパー生成
  → nherd           # サービス差分 restart/reload
  → nboot           # ブート世代エントリ同期
  ↑ ncompose        # 上記を固定順で呼び、失敗時は逆順 rollback を指揮
```

## ツール一覧

| ツール | 責務 | 状態 | リポジトリ |
|--------|------|------|-----------|
| **nput** | store パスの任意パスへの配置（symlink / copy）。世代管理・rollback | **active** | [yasunori0418/nput](https://github.com/yasunori0418/nput) |
| **nboot** | kernel / initrd の ESP 配置とブートエントリの世代同期 | planned | — |
| **nwrap** | setuid / capabilities ラッパーの生成（store の権限モデル上 symlink では原理的に不可能な領分） | planned | — |
| **nherd** | サービス差分の restart / reload / skip 適用。init 非依存の責務（初期実装は systemd バックエンド） | planned | — |
| **nshadow** | 宣言された users / groups と /etc/passwd・/etc/shadow の差分調停 | planned | — |
| **ncompose** | ツール群の固定順パイプライン実行と失敗時の逆順 rollback 指揮 | planned | — |
| **niface** | 上記全ツールが会話する JSON 規格（本リポジトリ） | **draft** | 本リポジトリ |

## 設計原則（要約）

1. **単一責務**: 各ツールは 1 つのことをする。nput のコンセプトに合わない責務は nput に入れず、独立したツールとして分離する
2. **standalone**: どのツールも単体で・どのディストロ上でも・JSON を手で与えて試せる（niface spec §8 の適合条件）
3. **規格が契約**: ツール間の会話は niface 規格のみに依存する。ツール固有の知識は `info` に隔離される
4. **オーケストレータは愚直に**: ncompose は固定順パイプライン + 失敗時逆順のみ。依存グラフエンジンにしない

## ドキュメント

| 文書 | 内容 |
|------|------|
| [distro-plan.md](distro-plan.md) | ディストリビューション構想の全体プラン（各ツールの詳細責務・standalone 契約・マイルストーン） |
| [../concept.md](../concept.md) | niface 規格のコンセプト |
| [../../spec/v1/spec.md](../../spec/v1/spec.md) | niface 規格の規範仕様 |

nput 固有の計画（mkEnv・plan / verify 等の機能拡張）は [nput リポジトリ](https://github.com/yasunori0418/nput)の issue / ADR が正。本リポジトリはエコシステム横断の文書のみを管理する。

## 未確定事項

- **傘 / ディストロ名**: 候補 nmanifold / nchord。北極星に近づいてから確定する
- 各 planned ツールの着手順序は distro-plan.md のマイルストーン（M0〜M4）を参照
