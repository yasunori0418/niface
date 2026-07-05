# niface

n プレフィックスツール群（nput / nboot / nwrap / nherd / nshadow / ncompose）が stdout / stdin で会話するための共通 JSON 規格の用語集。

ここは glossary であり仕様書ではない。規範は `spec/v1/spec.md`、設計判断は `docs/adr/` に置く。正名と、避けるべき同義語を固定する。

## 用語

### エンベロープと最上位

**エンベロープ (envelope)**: 全ツールが stdout に出す**単一の valid JSON 文書**。メタ情報を最上位に、実行ペイロードを `result` 配下に入れ子で持つ（→ ADR-0001）。stdout に出すのはこれ 1 つだけで、進捗・ログ・診断は stderr に逃がす。
_Avoid_: 「レスポンス」「複数行 JSON / NDJSON」（stdout は単一文書）。

**specVersion**: エンベロープの規格バージョン。整数。互換変更（フィールド追加）では上げず、削除・意味変更・必須化でのみ上げる（→ ADR-0010）。
_Avoid_: semver 文字列。

**status**: 実行全体の成否。`success` / `error` の**2 値のみ**。1 件でも item が failed なら `error`（→ ADR-0002）。exit code と連動する（`success` ⇔ 0、`error` ⇔ 非 0・→ ADR-0008）。
_Avoid_: `partial` などの中間値。

**dryRun**: plan / dry-run を表す最上位の boolean。`true` の出力は値以外において apply と同一スキーマ（→ ADR-0009）。
_Avoid_: plan 専用の別スキーマ・別 result 型。

**errors[]**: 最上位のエラー配列。**item に紐づかない全体エラーのみ**（入力 parse 失敗・lock 取得失敗等）を置く。item 起因のエラーは置かない（→ ADR-0006）。
_Avoid_: 全 item エラーの集約ビューをここに持たせること。

### ペイロード

**result**: 実行ペイロードのコンテナ。`items` / `changes` / `info` を持つ（→ ADR-0001）。
_Avoid_: items / changes を最上位に平置きすること。

**item**: `result.items[]` の要素。**処理単位の実行結果の記録**。規格が共通型を強制する。`id` / `kind` / `status` を必須に持つ（→ ADR-0003）。
_Avoid_: item に差分（可逆性）を持たせること（差分は change の領分）。

**change**: `result.changes[]` の要素。**状態遷移（差分）の宣言**。差分のある項目のみを列挙し、`reversible` を要素に必須で持つ。plan / apply の両方で出力する（→ ADR-0003）。
_Avoid_: noop（差分の無い項目）を列挙すること。

**reversible**: change の属性。その差分単位で巻き戻し可能か。可逆性は**行為の属性**なので item ではなく change に置く（→ ADR-0003）。

**info**: ツール固有情報の唯一の置き場所。item / change / result の 3 箇所に同じパターンで存在する。規格型は `additionalProperties: false` で閉じ、ツール固有は必ず `info` 配下に隔離する（→ ADR-0007）。
_Avoid_: ツール固有フィールドを規格フィールドと同一階層に混ぜること。

### identity と id

**identity**: item の**宣言上の同一性**を表す最小の値集合。`{ "kind": <string>, "key": <JSON value> }`。ツールが定めるのはこの中身のみ（→ ADR-0004）。

**key**: identity の一部。対象を一意に指す最小の値。ビルド毎・実行毎に変わる値（store パス・タイムスタンプ・世代番号）を含めない（→ ADR-0004）。

**id**: `lowercase-hex( sha256( JCS( identity ) ) )` で機械導出される 64 文字 hex。世代・実行・plan/apply を跨いで不変（→ ADR-0004）。
_Avoid_: 「`種別:キー`」形式の可読文字列 id。

**JCS**: RFC 8785（JSON Canonicalization Scheme）。id 導出の正準化に使う。niface の実装はサブセット（文字列 / 整数 / bool / null / 配列 / オブジェクト。浮動小数非対応）。

### エラーコード

**共通エラーコード**: 複数ツールで同じ意味で出しうるエラーの code。`E_<NAME>` 形式。初期 9 個で凍結し、追加は「2 つ以上のツールで必要」を条件とする（needs 駆動・→ ADR-0005）。

**ツール別エラーコード**: そのツールの概念を知らないと意味が取れないエラーの code。`E_<TOOL>_<NAME>` 形式。警告は `W_` prefix（→ ADR-0005）。

### 適合

**standalone**: 適合ツールの要件。入力は stdin の JSON か明示引数のみとし、状態・設定の暗黙探索や特定ディストリビューション・フレームワークへの依存をしない（→ `spec/v1/spec.md` §8）。

**適合 (conformance)**: `schema/v1/` の JSON Schema 検証と `testdata/v1/id-vectors.json` の全ベクタ通過をもって判定する（→ `spec/v1/spec.md` §8）。
