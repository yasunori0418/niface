# ディストリビューション構想（傘名は未確定・候補: nmanifold / nchord） — 必要な機能群とツールのプラン

北極星: NixOS とは異なる、Nix 版 Arch / Gentoo のような「最小のコア +
ユーザーが組み立てる + 優れた Handbook」型ディストリビューション。

基本方針: **nput のコンセプトに合わない責務は nput に入れず、単機能ツールとして分離し、
JSON パイプで会話するオーケストレータが組み合わせる**（UNIX 哲学）。
NixOS との差別化は「モジュールシステムという大聖堂を作らない」こと。

## 全体構成

```
build(Nix eval/build)
  → nput apply      # config / env の配置（既存 + #129 system mode + #162 mkEnv）
  → nshadow           # ユーザー/グループ差分調停
  → nwrap           # setuid ラッパー生成
  → nherd            # サービス差分 restart/reload
  → nboot           # ブート世代エントリ同期
  ↑ ncompose        # 上記を固定順で呼び、失敗時は逆順 rollback を指揮
```

## コアスイート（ディストロ成立に必須）

### nboot — ブート世代マネージャ【新規実装・最重要】

- **責務**: kernel / initrd / cmdline の ESP への copy、
  `/boot/loader/entries/*.conf` の生成・削除、nput 世代とエントリの同期
- **なぜ nput に入れないか**: ESP は FAT32 で symlink 不可・容量有限。
  保持世代数ポリシーや世代 GC とエントリ削除の同期という独自の関心事を持つ
- **価値**: 「ブートメニューから旧世代を選べる」= Nix 系ディストロの殺し文句。
  これがないと rollback が稼働中システムに閉じる
- **I/F**: `nboot sync --json < generations.json`

### nwrap — setuid ラッパー生成【新規実装】

- **責務**: `{ source, owner, group, setuid, capabilities }` のリストから
  `/run/wrappers/bin` 相当（tmpfs 上・boot 毎再生成）へ C 実装ラッパーを配置
- **なぜ nput に入れないか**: store は常に 0444 / 0555 で setuid ビットが立たず、
  symlink 配置では**原理的に解決不能**。NixOS `security.wrappers` の単機能切り出し
- **対象例**: sudo / passwd / ping

### nherd — サービス差分適用【既存ツール採用が先】

- **責務**: サービス配置差分から変更サービスの restart / reload / skip 判定 → 結果報告。
  **責務は特定 init システムに依存しない**（改名の理由。旧名 nsvc / nsysd 案は
  systemd 固定の含意があるため棄却）
- **方針**: 初期実装は systemd バックエンド（`daemon-reload` → unit 差分適用）とし、
  sd-switch がほぼこれなので**新規実装せず採用検討が先**。
  niface 規格に合わせる薄いアダプタで済む可能性が高い。
  バックエンド差し替え（他 init 対応）はniface 規格の info 隔離と同じ構造で
  将来拡張できる形にする
- **入力**: nput の `--json` 出力 2 世代分、または plan の `changes[]`

### nshadow — ユーザー / グループ照合【薄い差分ツール】

- **責務**: 宣言された users / groups の JSON と `/etc/passwd` / `/etc/shadow` の
  差分検出、`useradd` / `groupmod` コマンド列の生成（`--dry-run` で列挙のみ、
  実行はオーケストレータに委ねる設計も可）
- **位置づけ**: 「変異する状態ファイルは配置しない」境界の向こう側を担う調停者

### ncompose — オーケストレータ【愚直に】

- **責務**: 各ツールを正しい順序で呼ぶことだけ。失敗時は各ツールの rollback を
  逆順に呼ぶ指揮
- **設計原則**: 賢くしない。依存グラフエンジンではなく
  **固定の短いパイプライン + 失敗時逆順**。パイプライン定義自体を Nix の eval 時に
  生成すれば「オーケストレータも config を生成しない」を貫ける

## ツールより先に決める横断規格

### 1. niface 規格（最初の成果物はコードではなくこの文書）

- 各ツールが読む / 吐く JSON スキーマ: 世代表現・差分表現・結果エンベロープ
- **nput #130 のエンベロープ設計をエコシステム共通規格に昇格**させる。
  #130 の設計時に「nput 専用か将来の共通規格か」を意識しておくこと

### 2. rollback プロトコル

- 契約: 「各ツールは自分の領分の rollback を持つ」
  「オーケストレータは逆順に呼ぶだけ」
- **可逆 / 不可逆の表明を規格に含める**: nherd（restart の巻き戻し）や
  nshadow（ユーザー削除はしない）は完全 rollback 不能

### 3. 特権境界

- root 必須: nboot / nwrap / nshadow。nput は user 権限のまま
  （system mode #129 の euid gate 設計と揃える）

## 状態ファイルの境界ポリシー（設計判断・文書）

- 配置できないもの: `/etc/passwd`・`/etc/shadow`（nshadow の領分）、
  `/etc/machine-id`、DHCP が書く `resolv.conf` 等の実行時変異ファイル
- Arch 的に「命令的管理」と割り切る領域を**明文化した境界ポリシー文書**が必須

## 周辺（後続でよい）

- **base system manifest**: ディストロ最小構成のリファレンス flake
  （kernel + systemd + shell + nix + nput）。`/bin/sh`・`/usr/bin/env` の
  FHS シムもここの通常 root 配置 entry で表現。ツールではなくキュレーション
- **The Handbook**: Arch Wiki / Gentoo Handbook がディストロの本体。
  #161 の合成 idiom 集を「Handbook」へ格上げする発想。
  インストールは「パーティション作業 + chroot 内で ncompose を 1 回」の手順書
  （Gentoo 流。インストーラ自動化は後でよい）
- **GC 調停**: nput 世代・nboot エントリ・nix store GC root を跨ぐ削除可否の合意。
  当初は `nix-collect-garbage` + 各ツールの prune を ncompose が順に呼ぶだけで足りる
- **カーネルモジュール / firmware / udev**: 配置は nput で可能。
  reload idiom（`udevadm control --reload` 等）と `/run/booted-system` 相当
  （起動時カーネルと稼働モジュールの一致保証）の概念が nboot / nherd 側に必要
- **nixpkgs との関係の宣言**: 独自パッケージセットは持たず、pin / update の
  運用規範を定める（Arch の rolling release 相当の位置づけ）

## standalone 実行可能性（全ツール共通契約）

nput が standalone（NixOS / home-manager / flakes 非依存）で成立しているのと同じ性質を、
**全ツールの受け入れ条件**としてniface 規格に含める。
「nix がインストールされ nix 言語が実行可能なシステムなら、どのディストロ上でも
各ツールを単体で試せる」ことを保証する。

### 4 原則

1. **入力は stdin の JSON か明示引数のみ**: 状態ファイルや設定の暗黙探索をしない。
   JSON を生成するのが Nix であることすら要求しない（nput が flakes に
   依存しないのと同型）
2. **全ツールに `--dry-run` / `plan`**: 副作用なしで「何をするか」を出力できること。
   単体試行可能性の実体はこれ
3. **前提条件の自己申告（`check` サブコマンド）**: 「このホストで動く条件
   （systemd-boot の有無・root の有無等）」を検査・報告する口を規格化。
   ncompose は実行前にこれを束ねて呼ぶだけ
4. **配置先の注入可能性**: `/boot` や `/run/wrappers` をハードコードせず引数で受ける。
   テスト容易性と standalone 性は同じ設計から出る

### ツール別の注意点

- **nherd / nshadow / ncompose**: 本質的に「JSON を読み既存システムコマンドを呼ぶ」だけ。
  Nix すら必須でなく、既存ディストロ + systemd / shadow-utils で単体動作する。
  nshadow の `--dry-run`（コマンド列の列挙のみ）は root 不要で安全に試せる
- **nwrap**: root 必須なのは `activate` のみ。`generate --output ./dir --dry-run`
  （setuid ビットなしで生成内容のみ確認）を持たせ、root 無し単体試行を可能にする
- **nboot**: 既存ディストロでも「動いてしまう」ため最も注意深く設計する。
  `--esp` は**必須明示引数**（自動検出でホストのブートを触る事故を防ぐ）。
  `--esp ./fake-esp` のように任意ディレクトリを指せることで、
  **VM も root も無しで完全な単体テストが可能**になる

### 副産物: 開発順序

standalone 性を貫くと「ディストロを作る前に、既存ディストロ（例: NixOS）上で
部品を 1 つずつ本番検証できる」。nboot は fake ESP 相手に、nherd は実 systemd 相手に
独立して成熟させ、全部品が単体で信頼できるようになってから ncompose で束ねる。

## マイルストーン案

1. **M0**: niface 規格文書（#130 と同時期に着手）+ 状態境界ポリシー文書
2. **M1**: nput 側前提の完成 — system mode（#129）・mkEnv（#162）・
   plan / verify（nput リポジトリの機能計画を参照。`changes[]` が nherd の入力になる）
3. **M2**: nboot（ブート世代連携 ADR から着手）+ nwrap
4. **M3**: nherd（sd-switch 評価 → アダプタ）+ nshadow + ncompose
5. **M4**: base system manifest + Handbook 初版 + chroot インストール手順
