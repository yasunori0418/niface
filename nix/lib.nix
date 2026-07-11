# niface 規格の Nix 実装。
#
# deriveId は identity を JCS (RFC 8785) で正準化して sha256 する。正しさは
# builtins.toJSON が JCS と一致することに依存する:
#   - Nix の attrset はキー名でソート済み(JCS の UTF-16 順と ASCII 範囲で一致)
#   - builtins.toJSON はコンパクト出力(空白なし)・文字列は UTF-8 のまま(最小エスケープ)
# identity の値域は spec §5 に従う(文字列 / 整数 ±(2^53−1) / bool / null / 配列 /
# オブジェクト(メンバー名 ASCII))。域外(非整数数値・範囲外整数・非 ASCII の
# メンバー名)は deriveId が throw で拒否する。メンバー名を ASCII に限ることで
# attrset のバイト順と JCS の UTF-16 順の食い違い(BMP 外)を値域外へ排除する。
# 適合確認: flake の checks が testdata/v1/id-vectors.json との一致(vectors)と
# 域外拒否(rejected)を検証する。
{ lib ? null }:
rec {
  # identity 整数値の値域(spec §5): −(2^53−1) 〜 +(2^53−1)
  maxSafeInt = 9007199254740991;
  minSafeInt = -9007199254740991;

  # s の全バイトが ASCII(U+0000–U+007F)か。cntrl(0x00–0x1F, 0x7F)と
  # print(0x20–0x7E)の和が ASCII 全域に一致し、非 ASCII バイト(≥0x80)を除く。
  isAscii = s: builtins.match "[[:cntrl:][:print:]]*" s != null;

  # identity 値を spec §5 の値域で検証し、域内ならそのまま返す(域外は throw)。
  # 有効な値では元と同一の構造を返すため toJSON 出力(= 導出 id)は不変。
  checkDomain = v:
    let t = builtins.typeOf v;
    in if t == "int" then
      (if v > maxSafeInt || v < minSafeInt then
        throw "niface: integer ${toString v} is out of identity domain ±(2^53-1)"
      else v)
    else if t == "float" then
      throw "niface: non-integral number ${toString v} is unsupported in identity domain"
    else if t == "string" || t == "bool" || t == "null" then v
    else if t == "list" then map checkDomain v
    else if t == "set" then
      builtins.mapAttrs
        (name: value:
          if isAscii name then checkDomain value
          else throw "niface: object member name '${name}' must be ASCII in identity domain")
        v
    else throw "niface: unsupported type ${t} in identity domain";

  # identity ({ kind, key }) から item id を導出する(域外は throw)
  deriveId = identity:
    builtins.hashString "sha256"
      (builtins.toJSON (checkDomain { inherit (identity) kind key; }));

  # 全ベクタとの一致(vectors)と域外拒否(rejected)を検証する(true / throw)
  verifyVectors = vectorsFile:
    let
      data = builtins.fromJSON (builtins.readFile vectorsFile);
      checkVector = v:
        let got = deriveId v.identity;
        in if got == v.expected then true
           else throw "niface id vector mismatch: got ${got} want ${v.expected} (canonical: ${v.canonical})";
      checkRejected = v:
        let r = builtins.tryEval (deriveId v.identity);
        in if !r.success then true
           else throw "niface id vector: expected rejection but derived ${r.value} (reason: ${v.reason})";
      rejected = data.rejected or [ ];
    in builtins.all (x: x)
      ((map checkVector data.vectors) ++ (map checkRejected rejected));
}
