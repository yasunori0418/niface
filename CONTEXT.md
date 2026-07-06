# niface

n プレフィックスツール群（nput / nboot / nwrap / nherd / nshadow / ncompose）が stdout / stdin で会話するための共通 JSON 規格の用語集。

ここは glossary であり仕様書ではない。規範は `spec/v1/spec.md`、設計判断は `docs/adr/` に置く。正名と、避けるべき同義語を固定する。

## 用語

### エンベロープと最上位

**エンベロープ (envelope)**: 全ツールが stdout に出す**単一の valid JSON 文書**。メタ情報を最上位に、実行ペイロードを `results[]`（主体ごとの subjectResult）配下に入れ子で持つ。single / batch を問わず容器は常に `results[]`（→ ADR-0001, ADR-0011）。stdout に出すのはこれ 1 つだけで、進捗・ログ・診断は stderr に逃がす。
_Avoid_: 「レスポンス」「複数行 JSON / NDJSON」（stdout は単一文書）; `result` 単数の容器（`results[]` に一般化済み）。

**specVersion**: エンベロープの規格バージョン。整数。互換変更（フィールド追加）では上げず、削除・意味変更・必須化でのみ上げる（→ ADR-0010）。
_Avoid_: semver 文字列。

**status**: 実行全体の成否。`success` / `error` の**2 値のみ**。1 件でも item が failed なら `error`（→ ADR-0002）。exit code と連動する（`success` ⇔ 0、`error` ⇔ 非 0・→ ADR-0008）。
_Avoid_: `partial` などの中間値。

**dryRun**: plan / dry-run を表す最上位の boolean。`true` の出力は値以外において apply と同一スキーマ（→ ADR-0009）。
_Avoid_: plan 専用の別スキーマ・別 result 型。

**errors[]**: 最上位のエラー配列。**主体の列挙・解決の前段で起きる全体エラーのみ**（入力 parse 失敗・specVersion 不能・主体列挙自体の失敗等）を置く。解決済み主体に紐づくエラーは `subjectResult.errors` に置く（→ ADR-0006, ADR-0011）。
_Avoid_: 全 item エラーの集約ビューをここに持たせること; 解決済み主体の lock 取得失敗等をここに置くこと（subjectResult.errors が正）。

### ペイロード

**result**: 1 主体分の実行ペイロードのコンテナ。`items` / `changes` / `info` を持ち、各 subjectResult の下に 1 つ置かれる（→ ADR-0001, ADR-0011）。
_Avoid_: items / changes を最上位に平置きすること; `results[]`（主体の配列）と混同すること。

**item**: `result.items[]` の要素。**処理単位の実行結果の記録**。規格が共通型を強制する。`id` / `kind` / `status` を必須に持つ（→ ADR-0003）。
_Avoid_: item に差分（可逆性）を持たせること（差分は change の領分）。

**change**: `result.changes[]` の要素。**状態遷移（差分）の宣言**。差分のある項目のみを列挙し、`reversible` を要素に必須で持つ。plan / apply の両方で出力する（→ ADR-0003）。
_Avoid_: noop（差分の無い項目）を列挙すること。

**reversible**: change の属性。その差分単位で巻き戻し可能か。可逆性は**行為の属性**なので item ではなく change に置く（→ ADR-0003）。

**info**: ツール固有情報の唯一の置き場所。item / change / result の 3 箇所に同じパターンで存在する。規格型は `additionalProperties: false` で閉じ、ツール固有は必ず `info` 配下に隔離する（→ ADR-0007）。
_Avoid_: ツール固有フィールドを規格フィールドと同一階層に混ぜること。

### 複数主体（batch）と subject

**batch**: 複数主体を 1 実行で扱う実行形態の**非公式呼称**（schema 上の判別子値ではない。判別子フィールド `mode` は廃止 → ADR-0013）。`results[]` に主体ごとの subjectResult を並べる。同一 tool・同一 command が構造的に保証される（複数ツールの束ねは ncompose の領分で別形状）（→ ADR-0011）。
_Avoid_: batch を schema の判別子値として扱うこと; NDJSON / 複数文書で主体を並べること; batch に複数ツールを混載すること。

**subjectResult**: `results[]` の要素。1 主体分の実行結果（`subject` / `status` / `startedAt` / `finishedAt` / `errors?` / `result`）。`subject` は常時必須（→ ADR-0013）。specVersion / tool / command / dryRun は持たず最上位にのみ置く（→ ADR-0011）。
_Avoid_: subjectResult を単独の完全なエンベロープにすること（tool / command 等を要素に重複させない）。

**subject**: 操作の主体を名指す弱い識別子（`{ name }`）。各 `results[]` 要素で必須（→ ADR-0013）。`name` は 1 エンベロープ内で一意（producer MUST）。id 導出には関与しない。参照解決は `(tool.name, subject, id)` の 3 層で行う（→ ADR-0012）。
_Avoid_: subject を id 導出（identity）に混ぜること; world-wide 一意を要求すること。

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
