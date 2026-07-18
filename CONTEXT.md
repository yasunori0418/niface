# niface

n プレフィックスツール群（nput / nboot / nwrap / nherd / nshadow / ncompose）の実行結果を単一の構造化 JSON（結果エンベロープ）として stdout に出力するための共通規格の用語集。

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

**item**: `result.items[]` の要素。**処理単位の実行結果の記録**。規格が共通型を強制する。`id` / `kind` / `status` を必須に持つ（→ ADR-0003）。status の `skipped` は**前段の失敗による未実行にのみ**使い、方針による不作為（restart 不要判定・削除をしない方針等）は判定処理の成功なので `success` + 必要なら `warnings` / `info` で表す（→ ADR-0020）。
_Avoid_: item に差分（可逆性）を持たせること（差分は change の領分）; 方針による不作為を `skipped` で表すこと（`success` が正）。

**change**: `result.changes[]` の要素。**状態遷移（差分）の宣言**。差分のある項目のみを列挙し、`reversible` を要素に必須で持つ。plan / apply の両方で出力する（→ ADR-0003）。
apply では実際に生じた状態遷移の観測記録であり、item が `failed` でも生じた差分は列挙する。`status` が `error` で終わる実行でも、changes は適用済み差分を全て含める（producer MUST・→ ADR-0016）。
_Avoid_: noop（差分の無い項目）を列挙すること。

**reversible**: change の属性。その差分単位で巻き戻し可能か。可逆性は**行為の属性**なので item ではなく change に置く（→ ADR-0003）。

**info**: ツール固有情報の唯一の置き場所。エンベロープ最上位 / result / item / change の 4 箇所に同じパターンで存在する。最上位には主体に紐づかない実行全体の情報、result には主体ごとの情報を置く。規格型は `additionalProperties: false` で閉じ、ツール固有は必ず `info` 配下に隔離する（→ ADR-0007, ADR-0018）。
_Avoid_: ツール固有フィールドを規格フィールドと同一階層に混ぜること; 機微値を info / detail に生で載せること（マスク/ダイジェスト/省略で表す → ADR-0022）。

### 複数主体（batch）と subject

**batch**: 複数主体を 1 実行で扱う実行形態の**非公式呼称**（schema 上の判別子値ではない。判別子フィールド `mode` は廃止 → ADR-0013）。`results[]` に主体ごとの subjectResult を並べる。同一 tool・同一 command が構造的に保証される（複数ツールの束ねは ncompose の領分で別形状）（→ ADR-0011）。
_Avoid_: batch を schema の判別子値として扱うこと; NDJSON / 複数文書で主体を並べること; batch に複数ツールを混載すること。

**subjectResult**: `results[]` の要素。1 主体分の実行結果（`subject` / `status` / `startedAt` / `finishedAt` / `errors?` / `result`）。`subject` は常時必須（→ ADR-0013）。specVersion / tool / command / dryRun は持たず最上位にのみ置く（→ ADR-0011）。
_Avoid_: subjectResult を単独の完全なエンベロープにすること（tool / command 等を要素に重複させない）。

**subject**: 操作の主体を名指す弱い識別子（`{ name }`）。各 `results[]` 要素で必須（→ ADR-0013）。`name` は 1 エンベロープ内で一意（producer MUST）。id 導出には関与しない。参照解決は `(tool.name, subject, id)` の 3 層で行う（→ ADR-0012）。
_Avoid_: subject を id 導出（identity）に混ぜること; world-wide 一意を要求すること。

**generation**: subjectResult の任意フィールド。profile 世代遷移の**観測記録**（`{ profile, before, after }`）。`before` / `after` は実行開始 / 終了時点で profile が指していた世代番号で、観測できない場合はそれぞれ省略する。plan / dry-run では `after` = `before`。新世代の作成は `before` ≠ `after` で判定する。id 導出には関与しない（→ ADR-0015）。
_Avoid_: 「作成した世代」の宣言として扱うこと（観測の記録である）; 世代番号を identity / key に入れること（→ ADR-0004）。

### identity と id

**identity**: item の**宣言上の同一性**を表す最小の値集合。`{ "kind": <string>, "key": <JSON value> }`。ツールが定めるのはこの中身のみ。値は §5 の値域に限る（→ ADR-0004, ADR-0024）。

**key**: identity の一部。対象を一意に指す最小の値。ビルド毎・実行毎に変わる値（store パス・タイムスタンプ・世代番号）を含めない（→ ADR-0004）。

**id**: `lowercase-hex( sha256( JCS( identity ) ) )` で機械導出される 64 文字 hex。世代・実行・plan/apply を跨いで不変（→ ADR-0004）。
_Avoid_: 「`種別:キー`」形式の可読文字列 id。

**identity の値域**: identity を構成する JSON 値の許容域（spec §5・→ ADR-0024）。文字列（全 Unicode）/ 整数（±(2^53−1)、整数表記のみ）/ bool / null / 配列 / オブジェクト（メンバー名は ASCII）。数値は表記で判定し、小数点・指数表記（`1.0`・`1e3` 等）・NaN・Infinity は域外。負ゼロ表記（`-0`）は域内で `0` と同一 identity として扱い、実装が `0` に正規化する（→ ADR-0027）。域外は実装が id 導出時に拒否する。制約された値域の上で各実装はフル JCS と一致する。
_Note_: cross-language の巨大整数拒否を `testdata/v1/id-vectors.json` の rejected で固定するときは **≥2^64 の値**（例 `99999999999999999999`）を使う。`[2^63, 2^64)`（例 `9223372036854775808`）は Nix `builtins.fromJSON` が**パース時に throw** し、`verifyVectors` 冒頭の `fromJSON` が tryEval の外で失敗して checks.id-vectors 全体がクラッシュするため使わない。

**JCS**: RFC 8785（JSON Canonicalization Scheme）。id 導出の正準化に使う。niface は identity の値域を上記に制約し、その域上で実装は RFC 8785 と一致する（「サブセット実装」ではなく「値域制約」→ ADR-0024）。
_Avoid_: 「実装は JCS のサブセット」という言い回し（値域制約で捉える）。

### エラーコード

**共通エラーコード**: 複数ツールで同じ意味で出しうるエラーの code。`E_<NAME>` 形式（警告は `W_<NAME>`）。初期 9 個で凍結し、追加は「2 つ以上のツールで必要」を条件とする（needs 駆動・→ ADR-0005）。

**ツール別エラーコード**: そのツールの概念を知らないと意味が取れないエラーの code。`E_<TOOL>_<NAME>` 形式。警告は `W_` prefix（`W_<TOOL>_<NAME>`・→ ADR-0005）。

**warning**: 警告。構造は error と同形（`code` / `message` / `detail`）だが、`code` は `W_` prefix に限定され、schema は `$defs/warning` として error から型分離されている（→ ADR-0019）。エンベロープ最上位（実行全体）/ subjectResult（主体）/ item（処理単位）の 3 箇所に置け、責務分離は errors と同型。status の集約には影響しない。
_Avoid_: warning の code に `E_` prefix、error の code に `W_` prefix を入れること（schema の pattern で拒否される）; 可逆性を warning で運ぶこと（`change.reversible` の領分）。

### 適合

**standalone**: 適合ツールの要件。入力は stdin の JSON か明示引数のみとし、状態・設定の暗黙探索や特定ディストリビューション・フレームワークへの依存をしない（→ `spec/v1/spec.md` §8）。

**適合 (conformance)**: `schema/v1/` の JSON Schema 検証・schema で表現しきれない MUST のリント検査（status↔errors/item 整合・`changes[].itemId` の同一 result 内参照整合・`subject.name` / `item.id` の一意性）・`testdata/v1/id-vectors.json` の全ベクタ通過をもって判定する。schema 検証とリント検査の参照実装は Go（`go/conformance`、単一文書 CLI `niface-validate`）（→ `spec/v1/spec.md` §8, ADR-0021, ADR-0023）。
_Avoid_: schema で表現しきれない MUST を「schema を通れば適合」と見なして素通りさせること。
