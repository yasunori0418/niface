# ADR-0024: item id の identity 値域を確定し、域外を実装拒否とする

- ステータス: 採用
- 日付: 2026-07-11
- 関連: `spec/v1/spec.md`(§5・§2), `testdata/v1/id-vectors.json`, `go/id.go`, `go/id_test.go`, `nix/lib.nix`, `flake.nix`, ADR-0004
- 改訂対象: ADR-0010(specVersion を上げる条件に「相互運用が成立していなかった域外の縮小・厳格化は互換扱い」の例外を追加)

## 背景

ADR-0004 は id を `sha256(JCS(identity))` で導出し、JCS は RFC 8785 と定めた。RFC 8785 は全数値・非 ASCII に決定的な正準形を与えるため、規範としては全値域を RFC 8785 に委譲していた。しかし参照実装はそのフル仕様を実装しておらず、「RFC 8785 の全域で言語間互換」は**相互運用が一度も成立していない空手形**だった。この乖離が言語間の id 分裂を招く余地を残していた。

- Go(`go/id.go`)は非整数数値をエラーで拒否するサブセットだが、`int` / `int64` 経路には範囲ガードが無く、`float64` 経路にだけ非対称な `1e15` ガードがあった。2^53 を超える整数は、JSON の数値を double で読む実装との間で表現が割れ得る。加えて `encoding/json` は `1` と `1.0` を共に float64 にするため、整数値 float(`1.0`)を Go は受理し Nix は拒否する非対称もあった。
- Nix(`nix/lib.nix`)は float を拒否せず、`builtins.toJSON` の出力を**黙って**規格外の id に落とし込んでいた。
- attrset のキー順(Nix はバイト順)は JCS のキー順(UTF-16 code unit 順)と BMP 外の文字で食い違う。key 内オブジェクトのメンバー名に絵文字等を使うと Go と Nix で id が割れる。既存の id-vectors は非 ASCII を**値**でしか突いておらず、メンバー名は全ベクタ ASCII だった。

spec 上は第三者ツールがこれらの値を合法に使えたため、互換が構造的に保証されていなかった。本 ADR は、相互運用が成立していなかったこの領域を規範から締め出して値域を確定する。ADR-0004 の導出方式・安定性契約・key の最小性は改訂しない。

## 決定

- 位置づけを「JCS のサブセット実装」ではなく「**identity の値域制約**」とする。制約された値域の上では、各実装はフル JCS と完全一致する。
- identity(`kind` / `key`)を構成する JSON 値の値域を spec §5 に明文化する: 文字列(値は全 Unicode)/ 整数(−(2^53−1) 〜 +(2^53−1))/ bool / null / 配列 / オブジェクト(**メンバー名は ASCII に限る**)。**数値は値ではなく表記で判定し**、小数点・指数表記(`1.0`・`1e3` 等、値が整数でも)は域外とする。域外 identity を実装は id 導出時に**拒否しなければならない**(MUST)。
- Go: `encoding/json` は `1` と `1.0` を共に float64 にするため、表記を保持する `json.Number`(`UseNumber` デコード)でのみ数値を受理し `float64` 型は拒否する。整数表記のみ ±(2^53−1) で受理し、小数点・指数表記は拒否する。オブジェクトメンバー名の ASCII 検査を追加する。値域違反は sentinel error で分類する。
- Nix: `deriveId` に域外(float 型・範囲外整数・非 ASCII メンバー名)で throw する検査を追加する。`builtins.fromJSON` は小数・指数表記を float 型にするため、float 型を拒否すれば表記判定が Go と一致する。ASCII 検査は `builtins.match "[[:cntrl:][:print:]]*"`(= ASCII 全域)で行う。
- `testdata/v1/id-vectors.json` に `rejected: [{identity, reason}]`(非整数表記 `1.0`/`1e3`・範囲外整数・非 ASCII メンバー名)と境界 2^53−1 の受理ベクタを追加する(既存ハーネスは `vectors` しか読まないため後方互換)。Go テストは `UseNumber` デコードで、Nix は既存 `verifyVectors`(`builtins.tryEval`)で rejected を検証し、加えて nix-unit を導入して `deriveId`/`checkDomain`/`isAscii` を評価テストする。
- 本値域確定は相互運用が成立していなかった域外の締め出しであり、それに安全に依存できた消費側が存在しないため互換扱いとし specVersion を上げない。この例外を spec §2 の互換ポリシーに明文化する(ADR-0010 の specVersion 条件を精緻化)。

## 根拠

- 整数上限を安全整数(2^53−1)に揃えると、整数を double で読む実装(JavaScript 系・多くの JSON parser)との id 分裂を原理的に排除できる。
- **数値を表記で判定する**のが Go/Nix で機械的に一致する唯一の規則である。Nix の `builtins.toJSON` は integer-valued float を `1.0` と出すため、Nix で float を受理すると Go の `1`(整数正準形)と id が割れる。逆に Go は `encoding/json` が `1` と `1.0` を区別できないため、`json.Number` の字句(`.`/`e`/`E` の有無)でしか表記を判定できない。両者が保持できる情報の交点が「整数表記か否か」であり、そこに値域を揃える。
- メンバー名を ASCII に限れば、attrset のバイト順(Nix)と UTF-16 code unit 順(JCS)が一致する範囲に閉じ、キー順の食い違いによる分裂を値域外へ排除できる。文字列の**値**は正準化で順序に関与しないため、非 ASCII を許して良い。
- 「拒否(MUST)」を課すことで、黙って規格外 id を出す実装(旧 Nix)を適合検査で落とせる。`rejected` ベクタが Go / Nix 双方の拒否を言語横断で担保する。
- 値域は既存の「key は宣言上の同一性を表す最小の値集合」(ADR-0004)の下で実用上の制約にならない。
- specVersion を上げないのは、旧 spec が RFC 8785 に委譲していた域が**相互運用の成立しない空手形**で、そこに安全に依存できた消費側が存在しないため。機能の削除ではなく壊れていた領域の是正であり、互換破壊にならない(→ spec §2 の例外・ADR-0010)。

## 影響

- `spec/v1/spec.md`: §5 に identity の値域(表記ベース)、§2 に互換ポリシーの例外を追加。
- `go/id.go` / `go/id_test.go`: `json.Number` 表記判定・`float64` 型拒否・sentinel error・`UseNumber` デコード、境界/表記域外/整数域のテスト。
- `nix/lib.nix`: `deriveId` に値域検査(域外 throw)、`verifyVectors` を `rejected` 対応(`tryEval`)に拡張。`flake.nix` に nix-unit を導入し `nix/tests/` の評価テストを checks に載せる(既存 `verifyVectors` は維持し二層)。
- `testdata/v1/id-vectors.json`: `rejected`(非整数表記・範囲外整数・非 ASCII メンバー名)と境界受理ベクタを追加。
- `docs/design.md` の ADR 索引、`CONTEXT.md`(JCS / identity / key 項)、`CLAUDE.md` の JCS 記述を同期。
- id 導出のアルゴリズム(`sha256(JCS(identity))`)・安定性契約・key の MUST NOT(ADR-0004)は不変。

## 棄却した代替案

- **フル JCS を実装する**: 浮動小数の ECMAScript 数値表記・BMP 外文字の UTF-16 順ソートを全実装(Go / Nix / 将来の言語)で厳密一致させる負担が大きく、id の入力である key にそれらの値を使う実益が無い(key は最小の同一性集合)。値域を絞る方が互換の保証コストが低い。
- **メンバー名も含め非 ASCII を全面許可し、実装側で UTF-16 順ソートを徹底する**: Nix の `builtins.toJSON`(attrset はバイト順)を迂回する独自シリアライザが要り、`toJSON` に乗る現行実装の単純さを失う。
- **域外を拒否せず未定義動作とする**: 黙って割れる id を放置することになり、言語間互換という id-vectors の存在意義に反する。
