# niface: 仕様（specVersion 1）

本文書は規範である。キーワード MUST / MUST NOT / SHOULD / MAY は RFC 2119 に従う。
設計判断の根拠は design.md、目的と原則は concept.md を参照。

## 1. 出力チャネル

- ツールは stdout に**単一の valid JSON 文書（エンベロープ）のみ**を出力しなければ
  ならない（MUST）。複数文書・NDJSON・非 JSON の混在を禁止する
- 進捗・ログ・診断は stderr に出力する（形式は自由）
- exit code は POSIX 慣行に従う: 0 = 成功、非 0 = 失敗。
  トップレベル `status` と連動しなければならない（MUST）:
  `success` ⇔ 0、`error` ⇔ 非 0。消費側が依存してよいのはこの対応のみ

## 2. エンベロープ

```jsonc
{
  "specVersion": 1,
  "tool": { "name": "nput", "version": "0.9.0" },
  "command": "apply",
  "status": "success",                        // "success" | "error"
  "dryRun": false,
  "startedAt": "2026-07-05T12:34:56+09:00",   // RFC 3339
  "finishedAt": "2026-07-05T12:34:58+09:00",
  "errors": [ Error, ... ],
  "result": {
    "items":   [ Item, ... ],
    "changes": [ Change, ... ],
    "info":    { }                            // ツール固有
  }
}
```

- フィールド命名は camelCase とする
- `specVersion` は整数。互換変更（フィールド追加）では増やさない。
  フィールドの削除・意味変更・必須化で増やす
- `status` は 2 値。**1 件でも item が failed であれば `error`** としなければ
  ならない（MUST）
- `dryRun: true` の出力は、`dryRun` の値以外において apply と同一スキーマで
  なければならない（MUST）
- トップレベル `errors[]` には **item に紐づかない全体エラーのみ**を置く
  （入力 parse 失敗・lock 取得失敗等）。item 起因のエラーを置いてはならない
  （MUST NOT）
- 消費側は未知フィールドを無視しなければならない（MUST / must-ignore）

## 3. Item

`result.items[]` の要素。処理単位の実行結果の記録。

```jsonc
{
  "id": "<64 文字 lowercase hex>",   // 必須。§5 の導出による
  "kind": "entry",                   // 必須。種別（語彙はツール管理）
  "label": ".zshrc",                 // 推奨。人間向け表示名
  "status": "success",               // 必須。"success" | "failed" | "skipped"
  "error": Error,                    // status: "failed" のとき必須
  "warnings": [ Warning, ... ],      // 任意
  "info": { }                        // 任意。ツール固有
}
```

- `skipped` は「前段の失敗による未実行」等に使う。dry-run では使わない
  （dry-run の items は「実行したら何をするか」を表し status は原則 success）
- ツール固有フィールドは `info` 配下にのみ置く（MUST）。
  規格フィールドと同一階層に追加してはならない

### Error / Warning

```jsonc
{ "code": "E_NPUT_COLLISION", "message": "...", "detail": { } }
{ "code": "W_IRREVERSIBLE",   "message": "...", "detail": { } }
```

- `code` は §6 の命名規則に従う文字列。`detail` は任意のツール固有構造

## 4. Change

`result.changes[]` の要素。状態遷移（差分）の宣言。plan / apply の両方で出力する。

```jsonc
{
  "kind": "modify",             // 必須。"add" | "remove" | "modify"
  "itemId": "<対応 item の id>", // 必須
  "reversible": true,           // 必須。この差分単位で巻き戻し可能か
  "info": { }                   // 任意。差分の具体内容（old / new 等）
}
```

- **差分のある項目のみ**列挙する。noop を含めてはならない（MUST NOT）
- `reversible: false` の差分を含む実行の巻き戻しは不完全となる。
  消費側はこれを警告として扱うべきである（SHOULD）

## 5. Item id の導出

```
identity = { "kind": <string>, "key": <JSON value> }
id       = lowercase-hex( sha256( JCS( identity ) ) )
```

- JCS は RFC 8785（JSON Canonicalization Scheme）
- ツールが定めるのは identity の中身のみ。key は対象の
  **宣言上の同一性を表す最小の値集合**としなければならない（MUST）
- key にビルド毎・実行毎に変わる値（store パス・タイムスタンプ・世代番号）を
  含めてはならない（MUST NOT）
- 同一対象の id は世代・実行・plan/apply を跨いで不変でなければならない（MUST）。
  違反はツールの欠陥として扱う
- 一意性の範囲はツール内。複数ツールの集約時はエンベロープの `tool.name` で修飾する

## 6. エラーコード

- 二層命名: 共通コード `E_<NAME>`、ツール別コード `E_<TOOL>_<NAME>`。
  警告（`W_`）も同様
- 判断基準: 他のツールでも同じ意味で出しうるなら共通、
  ツールの概念を知らないと意味が取れないならツール別
- 共通コードの追加は「2 つ以上のツールで同じ意味の code が必要になった」ことを
  条件とする（needs 駆動）。追加は互換変更

### 共通コード初期レジストリ

| code | 意味 |
|------|------|
| `E_INPUT` | stdin / 引数の JSON が規格違反・parse 不能 |
| `E_SPEC_VERSION` | JSON は正しいが specVersion が扱えない |
| `E_PERMISSION` | 権限不足 |
| `E_IO` | ファイルシステム・外部 I/O の失敗 |
| `E_LOCK` | 排他制御の取得失敗 |
| `E_NOTFOUND` | 参照先の不在 |
| `E_PRECONDITION` | 前提条件の不成立（check 系） |
| `E_UNSUPPORTED` | 未対応の機能・環境 |
| `E_INTERNAL` | ツール内部の想定外状態（バグ） |

## 7. check サブコマンド

- 専用の result 型は設けない。前提条件 1 つ = Item 1 つで表現する
- 不成立は `status: "failed"` + `error.code: "E_PRECONDITION"`
- 1 件でも不成立ならトップレベル `status: "error"`・exit code 非 0
  （消費側は check を実行可否の gate として `&&` 接続できる）
- check は副作用を持たない。`dryRun` フィールドは出力し、値は `true` を推奨

## 8. 適合

- 本仕様への適合は、schema/v1/ の JSON Schema 検証と
  testdata/v1/id-vectors.json の全ベクタ通過をもって判定する
- 適合するツールは standalone で動作しなければならない（MUST）:
  入力は stdin の JSON か明示引数のみとし、状態・設定の暗黙探索、
  特定ディストリビューション・特定フレームワークへの依存をしない
