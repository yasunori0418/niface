# ADR-0024: item id の identity 値域を確定し、域外を実装拒否とする

- ステータス: 採用
- 日付: 2026-07-11
- 関連: `spec/v1/spec.md`(§5), `testdata/v1/id-vectors.json`, `go/id.go`, `nix/lib.nix`, ADR-0004

## 背景

ADR-0004 は id を `sha256(JCS(identity))` で導出すると決めたが、identity を構成する JSON 値の**値域**は未規定のままだった。JCS は RFC 8785 のフル仕様を指す一方、参照実装はその一部しか実装しておらず、両者のずれが言語間の id 分裂を招く余地を残していた。

- Go(`go/id.go`)は非整数数値をエラーで拒否するサブセットだが、`int` / `int64` 経路には範囲ガードが無く、`float64` 経路にだけ非対称な `1e15` ガードがあった。2^53 を超える整数は、JSON の数値を double で読む実装との間で表現が割れ得る。
- Nix(`nix/lib.nix`)は float を拒否せず、`builtins.toJSON` の出力を**黙って**規格外の id に落とし込んでいた。
- attrset のキー順(Nix はバイト順)は JCS のキー順(UTF-16 code unit 順)と BMP 外の文字で食い違う。key 内オブジェクトのメンバー名に絵文字等を使うと Go と Nix で id が割れる。既存の id-vectors は非 ASCII を**値**でしか突いておらず、メンバー名は全ベクタ ASCII だった。

spec 上は第三者ツールがこれらの値を合法に使えたため、互換が構造的に保証されていなかった。ADR-0004 は導出方式と key の最小性を定めたが値域は空白で、その空白がこの分裂リスクの源だった。本 ADR はその空白を埋める新規決定であり、ADR-0004 の導出方式・安定性契約は改訂しない。

## 決定

- 位置づけを「JCS のサブセット実装」ではなく「**identity の値域制約**」とする。制約された値域の上では、各実装はフル JCS と完全一致する。
- identity(`kind` / `key`)を構成する JSON 値の値域を spec §5 に明文化する: 文字列(値は全 Unicode)/ 整数(−(2^53−1) 〜 +(2^53−1))/ bool / null / 配列 / オブジェクト(**メンバー名は ASCII に限る**)。非整数の数値は MUST NOT。域外 identity を実装は id 導出時に**拒否しなければならない**(MUST)。
- Go: `int` / `int64` 経路に ±(2^53−1) ガードを追加し、`float64` 経路のガードを同値(`1e15` → 2^53−1)に揃える。オブジェクトメンバー名の ASCII 検査を追加する。
- Nix: `deriveId` に域外(非整数数値・範囲外整数・非 ASCII メンバー名)で throw する検査を追加する。ASCII 検査は `builtins.match "[[:cntrl:][:print:]]*"`(= ASCII 全域)で行う。
- `testdata/v1/id-vectors.json` に `rejected: [{identity, reason}]` を互換追加する(既存ハーネスは `vectors` しか読まないため後方互換)。Go テストと Nix `verifyVectors`(`builtins.tryEval`)を rejected の拒否検証に対応させ、`flake.nix` は変更しない(`verifyVectors` 内で完結)。

## 根拠

- 整数上限を安全整数(2^53−1)に揃えると、整数を double で読む実装(JavaScript 系・多くの JSON parser)との id 分裂を原理的に排除できる。`int` / `int64` / `float64` の 3 経路に同一の範囲を課すことで Go 内部の非対称も消える。
- メンバー名を ASCII に限れば、attrset のバイト順(Nix)と UTF-16 code unit 順(JCS)が一致する範囲に閉じ、キー順の食い違いによる分裂を値域外へ排除できる。文字列の**値**は正準化で順序に関与しないため、非 ASCII を許して良い。
- 「拒否(MUST)」を課すことで、黙って規格外 id を出す実装(旧 Nix)を適合検査で落とせる。`rejected` ベクタが Go / Nix 双方の拒否を言語横断で担保する。
- 値域は既存の「key は宣言上の同一性を表す最小の値集合」(ADR-0004)の下で実用上の制約にならない。

## 影響

- `spec/v1/spec.md` §5 に「identity の値域」を追加。
- `go/id.go`: 範囲ガード(int / int64 追加・float64 を 2^53−1 に統一)とメンバー名 ASCII 検査。`go/id_test.go`: `rejected` 検証と整数域ユニットテスト。
- `nix/lib.nix`: `deriveId` に値域検査(域外 throw)、`verifyVectors` を `rejected` 対応(`tryEval`)に拡張。`flake.nix` は不変。
- `testdata/v1/id-vectors.json`: `rejected` セクション追加。
- `docs/design.md` の ADR 索引、`CLAUDE.md` の JCS 記述を同期。
- id 導出のアルゴリズム(`sha256(JCS(identity))`)・安定性契約・key の MUST NOT(ADR-0004)は不変。

## 棄却した代替案

- **フル JCS を実装する**: 浮動小数の ECMAScript 数値表記・BMP 外文字の UTF-16 順ソートを全実装(Go / Nix / 将来の言語)で厳密一致させる負担が大きく、id の入力である key にそれらの値を使う実益が無い(key は最小の同一性集合)。値域を絞る方が互換の保証コストが低い。
- **メンバー名も含め非 ASCII を全面許可し、実装側で UTF-16 順ソートを徹底する**: Nix の `builtins.toJSON`(attrset はバイト順)を迂回する独自シリアライザが要り、`toJSON` に乗る現行実装の単純さを失う。
- **域外を拒否せず未定義動作とする**: 黙って割れる id を放置することになり、言語間互換という id-vectors の存在意義に反する。
