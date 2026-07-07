# niface: 仕様（specVersion 1）

本文書は規範である。キーワード MUST / MUST NOT / SHOULD / MAY は RFC 2119 に従う。設計判断の根拠は design.md、目的と原則は concept.md を参照。

## 1. 出力チャネル

- ツールは stdout に**単一の valid JSON 文書（エンベロープ）のみ**を出力しなければならない（MUST）。複数文書・NDJSON・非 JSON の混在を禁止する
- 進捗・ログ・診断は stderr に出力する（形式は自由）
- exit code は POSIX 慣行に従う: 0 = 成功、非 0 = 失敗。トップレベル `status` と連動しなければならない（MUST）: `success` ⇔ 0、`error` ⇔ 非 0。消費側が依存してよいのはこの対応のみ
- シグナル等により実行が中断された場合でも、ツールは可能な限り、中断時点までの `items` と適用済みの `changes` を含む valid なエンベロープを stdout に出力すべきである（SHOULD）。この場合 `status` は `error`、exit code は非 0 とする

## 2. エンベロープ

トップレベルは常に `results[]`（主体ごとの結果の配列）を持ち、各要素は主体を `subject` で名指す。

```jsonc
{
  "specVersion": 1,
  "tool": { "name": "nput", "version": "0.9.0" },
  "command": "apply",
  "status": "success",                        // "success" | "error"（集約）
  "dryRun": false,
  "startedAt": "2026-07-05T12:34:56+09:00",   // RFC 3339
  "finishedAt": "2026-07-05T12:34:58+09:00",
  "errors": [ Error, ... ],                   // 主体列挙・解決の前段エラーのみ
  "warnings": [ Warning, ... ],               // 実行全体の警告（任意）
  "info": { },                                // ツール固有（実行全体。任意）
  "results": [ SubjectResult, ... ]           // 主体ごとの結果
}
```

- フィールド命名は camelCase とする
- `specVersion` は整数。互換変更（フィールド追加）では増やさない。フィールドの削除・意味変更・必須化で増やす
- コンテナは single / batch（`apply --all` 等の複数主体実行）を問わず常に `results[]` であり、各要素の主体を `subject` で名指す。実行形態を切り替える判別子フィールドは持たない。`results` の要素数は 0 以上で、起動の性質（単一主体か複数主体か）によらない
- `status` は 2 値。**`results` のいずれかの主体が `error`、または主体列挙・解決の前段で全体エラーがあれば `error`** としなければならない（MUST）。`results` が空でエラーも無ければ `success`（対象 0 件の実行は success）
- `dryRun` はエンベロープ全体の性質で、`results` の全主体で均一とする（MUST）。`tool` / `command` / `specVersion` もトップレベルにのみ置き、主体ごとに変えない
- `dryRun: true` の出力は、`dryRun` の値以外において apply と同一スキーマでなければならない（MUST）
- トップレベル `errors[]` には**主体の列挙・解決の前段で起きる全体エラーのみ**を置く（入力 parse 失敗・specVersion 不能・主体列挙自体の失敗等）。解決済み主体に紐づくエラーを置いてはならない（MUST NOT）
- トップレベル `warnings[]` には**実行全体に関わる警告**（主体に紐づかないもの）を置いてよい（MAY）。解決済み主体に紐づく警告は `subjectResult.warnings` に、処理単位に紐づく警告は `item.warnings` に置く。warning は `status` の集約に影響しない
- warning の `code` は `W_` prefix、error の `code` は `E_` prefix でなければならない（MUST・§6 の二層命名）。schema は `$defs/warning` / `$defs/error` の pattern でこれを強制する
- トップレベル `info` はツール固有情報の置き場（任意）。**主体に紐づかない実行全体の情報**を置いてよい（MAY）。主体ごとのツール固有情報は `result.info` に置く。他の `info` と同様、ツール固有フィールドを規格フィールドと同一階層に追加してはならない（MUST）
- 消費側は未知フィールドを無視しなければならない（MUST / must-ignore）

### SubjectResult

`results[]` の要素。1 主体分の実行結果。

```jsonc
{
  "subject": { "name": "home" },              // 主体。常に必須
  "status": "success",                        // "success" | "error"
  "generation": {                             // profile 世代遷移の観測記録（任意）
    "profile": "/nix/var/nix/profiles/per-user/alice/home",
    "before": 41,
    "after": 42
  },
  "startedAt": "2026-07-05T12:34:56+09:00",   // この主体の実行時刻
  "finishedAt": "2026-07-05T12:34:58+09:00",
  "errors": [ Error, ... ],                   // この主体に紐づく全体エラー（任意）
  "warnings": [ Warning, ... ],               // この主体に紐づく警告（任意）
  "result": {
    "items":   [ Item, ... ],
    "changes": [ Change, ... ],
    "info":    { }                            // ツール固有
  }
}
```

- `subject` は操作の主体を名指す弱い識別子で、各 `results[]` 要素で必須（MUST）。`subject.name` は 1 エンベロープ内で一意でなければならない（producer MUST）。`subject` は id 導出に関与しない（§5）
- `subject.status` が `error` になるのは、その主体の item が 1 件でも `failed`、またはその主体に紐づく全体エラー（例: その主体の lock 取得失敗）があるときである（MUST）
- `errors[]` にはその主体に紐づく全体エラー（その主体の item に紐づかないもの）を置く。item 起因のエラーを置いてはならない（MUST NOT）
- `warnings[]` にはその主体に紐づく警告（その主体の item に紐づかないもの）を置いてよい（MAY）。item 起因の警告は `item.warnings` に置く
- `generation` は profile 世代遷移の**観測記録**（任意）。`profile` は実際に使用した profile のパスで、`generation` を出力する場合は必須。`before` / `after` は実行開始時点 / 終了時点で profile が指していた世代番号であり、観測できない場合はそれぞれ省略する（初回実行に `before` は無い。profile 未作成の plan には `after` も無い）。profile を管理するツールは `generation` を出力すべきである（SHOULD）
- `generation` は作成の宣言ではなく観測の記録である。plan / dry-run では切替が起きないため `after` = `before` となる。新世代の作成は `before` ≠ `after` で判定する。世代番号は item id の導出に関与せず、key に含めてはならない原則（§5）は不変

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

- `skipped` は「前段の失敗による未実行」等に使う。dry-run では使わない（dry-run の items は「実行したら何をするか」を表し status は原則 success）
- ツール固有フィールドは `info` 配下にのみ置く（MUST）。規格フィールドと同一階層に追加してはならない

### Error / Warning

```jsonc
{ "code": "E_NPUT_COLLISION",       "message": "...", "detail": { } }
{ "code": "W_NPUT_FOREIGN_SYMLINK", "message": "...", "detail": { } }
```

- `code` は §6 の命名規則に従う文字列。`detail` は任意のツール固有構造
- 可逆性は warning で運ばない。差分が巻き戻し可能かは `change.reversible` が運ぶ（§4）。消費側はそれを警告として扱ってよい

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
- `reversible: false` の差分を含む実行の巻き戻しは不完全となる。消費側はこれを警告として扱うべきである（SHOULD）
- apply（`dryRun: false`）の changes は**実際に生じた状態遷移の観測記録**である。change を出すかどうかは item の status と独立に「現実が変化したか」で決まり、item が `failed` でも失敗までに現実が変化したなら対応する change を出力する。その `kind` は意図した遷移ではなく実際に生じた遷移を表し、`reversible` はその生じた遷移についてツールが判断する
- apply では、`status` が `error` で終わる実行でも、`changes[]` は失敗時点までに実際に生じた差分を**全て**含めなければならない（MUST）。plan（`dryRun: true`）が error で終わる場合の changes の完全性は要求しない。check は副作用を持たないため対象外（§7）

## 5. Item id の導出

```
identity = { "kind": <string>, "key": <JSON value> }
id       = lowercase-hex( sha256( JCS( identity ) ) )
```

- JCS は RFC 8785（JSON Canonicalization Scheme）
- ツールが定めるのは identity の中身のみ。key は対象の**宣言上の同一性を表す最小の値集合**としなければならない（MUST）
- key にビルド毎・実行毎に変わる値（store パス・タイムスタンプ・世代番号）を含めてはならない（MUST NOT）
- 同一対象の id は世代・実行・plan/apply を跨いで不変でなければならない（MUST）。違反はツールの欠陥として扱う
- `subject` は id 導出に関与しない。id の入力は identity（`{kind, key}`）のみである
- 一意性と参照解決は 3 層で修飾する: id は **1 主体（`subject`）の `result` 内**で一意、`subject` が `tool` 内で修飾、`tool.name` が複数ツール集約時に修飾する。消費側は `(tool.name, subject, id)` の 3 つ組で参照を解決する。`subject` は全 `results[]` 要素に常在するため 3 つ組は常に揃う。同一 identity の id は主体を跨いで衝突してよい（同じ論理対象を表すため当然である）
- `changes[].itemId` は**同じ `result`（同一 subjectResult）内**の item を参照しなければならない（MUST）。`result` を跨いで参照してはならない（MUST NOT）

## 6. エラーコード

- 二層命名: 共通コード `E_<NAME>`、ツール別コード `E_<TOOL>_<NAME>`。警告（`W_`）も同様
- 判断基準: 他のツールでも同じ意味で出しうるなら共通、ツールの概念を知らないと意味が取れないならツール別
- 共通コードの追加は「2 つ以上のツールで同じ意味の code が必要になった」ことを条件とする（needs 駆動）。追加は互換変更

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
- 1 件でも不成立ならトップレベル `status: "error"`・exit code 非 0（消費側は check を実行可否の gate として `&&` 接続できる）
- check は副作用を持たない。`dryRun` フィールドは出力し、値は `true` を推奨

## 8. 適合

- 本仕様への適合は、schema/v1/ の JSON Schema 検証と testdata/v1/id-vectors.json の全ベクタ通過をもって判定する
- 適合するツールは standalone で動作しなければならない（MUST）: 入力は stdin の JSON か明示引数のみとし、状態・設定の暗黙探索、特定ディストリビューション・特定フレームワークへの依存をしない
