# HANDOFF — claude.ai セッションからの引き継ぎ(2026-07-05)

このリポジトリの全内容は claude.ai 上の設計セッションで作成された。
本文書はそのセッションの到達点と未完了事項の引き継ぎ。恒常的な規約は
CLAUDE.md、設計判断の全記録は docs/design.md にある。

## セッションで確定したこと

### 命名(docs/ecosystem/naming.md に経緯と実測)

- 規格 = **niface**(n + interface。GitHub 衝突 0 件を実測済み)
- ツール = nput(既存)/ nboot / nwrap / **nherd**(旧 nsvc。systemd 非依存が
  改名理由)/ **nshadow**(旧 nuser)/ ncompose
- 傘/ディストロ名 = **未確定**(候補: nmanifold / nchord。北極星に近づいてから)

### 規格(specVersion 1 の初期論点は全確定)

19 の設計判断(docs/design.md の表)。要点: nested result・2 値 status・
items/changes 分離・id の hash 導出(JCS+sha256)・info 隔離・POSIX exit code・
共通エラーコード 9 個で凍結。

### 成果物の検証状態

| 成果物 | 状態 |
|--------|------|
| schema/v1/envelope.schema.json | testdata と相互検証済み(valid 5 全通過・invalid 4 全拒否) |
| testdata/v1/id-vectors.json | Python(JCS サブセット)で**実計算済み**の 8 ベクタ |
| go/(型 + id 導出 + テスト) | **未実行**(作成環境に Go なし) |
| nix/lib.nix + flake.nix | **未実行**(作成環境に Nix なし) |

## 完了済み(2026-07-05 追記・別セッション)

- `git init`(main ブランチ)実施済み。**初回コミットは未実施**
- README を英語 `README.md` + 日本語 `README.ja.md` に整理(不自然な 80 字折返し除去)
- root `flake.nix` を `genAttrs` → flake-parts(`flake-parts.lib.mkFlake`)へ変更。
  `lib`・`checks`(id-vectors / schema / go)・`formatter`(pkgs.nixfmt)を提供
- `CLAUDE.md` を簡潔なプロジェクト概要へ更新
- 開発環境 `dev/`(flake-parts)を新設。nput ＋ mattpocock/skills を主軸にした
  devShell。`example.envrc`(`use flake ./dev`)をコミット、`.envrc` はローカル
- **`nix flake check` 全 3 checks 通過を確認済み**。HANDOFF が懸念していた
  Nix builtins.toJSON と JCS の非 ASCII エスケープ不一致・go sandbox 問題は
  顕在化せず、実装は健全
- dev devShell の E2E 確認済み: `nix develop ./dev` の shellHook で
  `nput apply skills` が走り、10 skill が `.claude/skills/` へ配置される

## 次のアクション(優先順)

1. 初回コミット → GitHub リポジトリ作成(yasunori0418/niface 想定)
2. CI(GitHub Actions or flake check ベース)の追加

## 関連する nput 側の未完了(nput リポジトリで管理)

- issue #162(mkEnv)に A 案/B 案のコメント投稿
  (claude.ai セッションでドラフト作成済み: planA = mkEnv 実装、
  planB = entry 直接指定は docs idiom のみ・関数は実装しない)
- nput の #130/#131(--json)実装時、エンベロープを niface 規格に合わせる。
  「nput 専用か niface 共通か」の分岐点がここ
- nput 側の未計画機能(plan/diff・verify・初回衝突ポリシー・copy 権限・
  atomicity 境界・世代 UX)は nput リポジトリへ issue 化する

## 未解決の設計論点(急がない)

- 傘/ディストロ名の確定(nmanifold vs nchord。nchord は「エヌコード」が
  なろう N コード・encode と同音になる問題あり)
- niface: 進捗イベントの stderr NDJSON 形式(needs 駆動で将来定義)
- distro-plan の M0 残り: 状態境界ポリシー文書(宣言的/命令的の線引き)は未着手
- ncompose のパイプライン定義形式の詳細(固定順 + onFailure のみ、の方針は確定)

## このセッションの経緯(1 段落)

nput の仕組みの議論から始まり、mkEnv(buildEnv 合成環境の配置)の設計 →
nput issue #162 起票 → ディストロ北極星の宣言 → UNIX 哲学によるツール分割
(nboot/nwrap/nherd/nshadow/ncompose)→ パイプ規格の grilling による策定
(エンベロープ 19 判断)→ 命名確定 → 本リポジトリの構築、と進んだ。
判断の根拠が必要になったら docs/design.md と docs/ecosystem/naming.md が
一次記録。それ以上の背景は元セッションにしかないため、疑問が残る判断は
ユーザーに直接確認すること。
