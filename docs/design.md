# niface: 設計

規格の設計判断とその根拠を記録する。規範的な定義は spec.md を正とする。

## 判断一覧

| # | 論点 | 決定 |
|---|------|------|
| 1 | ペイロード配置 | `result` キー配下に分離（nested） |
| 2 | specVersion | 整数 |
| 3 | 命名規約 | camelCase（nput manifest.json と整合） |
| 4 | エラーコード | 文字列コード。共通（`E_*`）+ ツール別 prefix（`E_NBOOT_*`）の二層 |
| 5 | status 語彙 | `success` / `error` の 2 値 + item 単位 status |
| 6 | 部分失敗 | 1 件でも item 失敗 → トップレベル `error`・exit code 非 0 |
| 7 | dry-run | トップレベル `dryRun: boolean`。スキーマは apply と同一 |
| 8 | 時刻 | `startedAt` + `finishedAt`（RFC 3339） |
| 9 | item 共通型 | 規格で強制 |
| 10 | エラーの置き場所 | item 起因は items 内のみ、全体エラーはトップ `errors[]` のみ。重複させない |
| 11 | 拡張ポリシー | 未知フィールドは must-ignore。追加は互換変更 |
| 12 | exit code | POSIX 慣行のまま。独自体系を設けない |
| 13 | 共通コード初期セット | 9 個で凍結、以降 needs 駆動追加 |
| 14 | changes[] | items と分離。差分のみ列挙・`reversible` を要素に必須 |
| 15 | check の result 型 | 専用型を設けず items を再利用 |
| 16 | id の安定性 | 世代を跨いで不変を必須保証（違反はバグ） |
| 17 | id の導出 | `sha256(JCS(identity))` フル hex 64 文字。identity = `{kind, key}` |
| 18 | id の可読性担保 | `kind` を必須フィールド化、`label` を推奨 |
| 19 | ツール固有フィールド | `info` 配下に隔離（item / change / result の 3 箇所同パターン） |

## 主要な根拠

### stdout 単一 JSON 文書（進捗は stderr）

パイプ接続の根幹。stdout に複数文書や非 JSON が混ざると、消費側のパースが「行ごとの推測」になり全ツールが壊れる。進捗・診断は stderr に逃がすことで、`tool | jq` が常に成立する。

### status 2 値 + item 単位 status（決定 5・6）

全体 status に partial のような中間値を設けると、消費側が「partial のとき何をすべきか」の解釈を迫られる。2 値 + 「1 件でも失敗なら error」はオーケストレータの分岐を単純化し（非 0 なら止める / 巻き戻す）、 **何が適用済みか**という詳細は items が機械判定可能な形で運ぶ。これは nput の部分失敗集約（continue + aggregate）の意味論をエコシステム全体に広げて適用したもの。

### エラーの責務分離（決定 10）

item に紐づくエラーの一次情報源は item 自身とし、トップ `errors[]` は item に紐づかない全体エラー（入力 parse 失敗・lock 失敗）専用とする。集約ビューは `jq` で導出可能なため、エンベロープに冗長ビューを持つのは過剰。 items の責務を「宣言された処理単位の結果」に限定することで型と意味の両方が閉じる。

### exit code を POSIX のまま使う（決定 12）

ncompose がツールをシェルコマンドとして実行する以上、exit code の意味論はシェル文化に馴染むものが最良。nput の 0/1/2 も POSIX 準拠の範囲内で 1 / 2 の役割を明示定義しているだけであり、消費側契約は「0 か非 0 か」に限定する。

### items と changes の分離（決定 14）

items = 処理単位の**実行結果の記録**、changes = **状態遷移（差分）の宣言**。 dry-run スキーマ同一原則により plan / apply の両方が changes を出力し、 plan では「実行したら生じる差分」、apply では「実際に生じた差分」となる。可逆性（`reversible`）は行為の属性なので changes 要素に置く（item は結果の記録であり可逆性を持たない）。ncompose の rollback 指揮は changes の走査で完結する。

### id の機械的導出（決定 16〜18）

「種別:キー」の文字列規約は、エスケープ・文字種・区切り文字衝突といった形式上の問題をツール実装者に押し付ける。JCS（RFC 8785）正準化 + sha256 の導出を規格が定義することで、ツールに残る判断を **key の選定（宣言上の同一性を表す最小の値集合）だけ**に帰着させる。安定性契約（決定 16）の遵守も「key にビルド毎に変わる値を含めない」という検証可能な規則に単純化される。可読性の喪失は `kind` 必須 + `label` 推奨で補う。

### info への隔離（決定 19）

同一階層マージは Go の encoding/json で flatten 非対応のため二段デコードを強い、 JSON Schema も `additionalProperties` 頼みで規格型を閉じられない。 `info` 一段掘りにより規格型は `additionalProperties: false` で閉じ、 `Item[T any]` の Generics で型付けでき、ツールを知らない消費側は `Item[json.RawMessage]` で規格部分だけを安全に扱える。代償は jq パスの一段（`.info.target`）のみ。

## リポジトリ構成と適合戦略

```
niface/
├── docs/             # concept.md / design.md（本文書）
├── spec/v1/          # 人間向け規格文書（規範は spec.md）
├── schema/v1/        # JSON Schema（機械可読の正）
├── testdata/v1/      # valid / invalid サンプル + id-vectors.json
├── go/               # Go 参照実装（型 + id 導出 + ベクタテスト）
├── nix/              # Nix lib（id 導出・ベクタ検証）
├── scripts/          # schema 検証スクリプト（flake checks から使用）
└── flake.nix         # checks: id-vectors / schema / go
```

- **正は JSON Schema**。文書とコードはそれに従う
- **id-vectors.json が最重要資産**: identity → 期待 id の対応表。 JCS の罠（非 ASCII キー・数値表現・ネスト）を突くベクタを含め、全言語実装が CI でこれを通すことで導出の互換を証明する
- 参照方法は二経路: Go ツールは go module で型を共有（コンパイル時）、各ツールの flake が本リポジトリを input に取り schema 検証 + id-vectors 適合を checks で回す（CI 時）
- バージョニング: specVersion 整数とディレクトリ（v1/）を一致させ、互換変更は v1 内 + git タグ、非互換変更のみ v2/ を新設
- ツール固有の info schema は各ツールのリポジトリが管理する（規格側で抱えない）
