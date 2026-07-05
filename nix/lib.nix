# niface 規格の Nix 実装。
#
# deriveId の正しさは builtins.toJSON が JCS (RFC 8785) サブセットと一致する
# ことに依存する:
#   - Nix の attrset はキー名でソート済み(JCS の UTF-16 順とは ASCII 範囲で一致。
#     非 ASCII キーを使う場合は要検証)
#   - builtins.toJSON はコンパクト出力(空白なし)
#   - 文字列は UTF-8 のまま出力(最小エスケープ)
#   - 浮動小数は使用しないこと(整数のみ)
# 適合確認: flake の checks が testdata/v1/id-vectors.json との一致を検証する。
{ lib ? null }:
rec {
  # identity ({ kind, key }) から item id を導出する
  deriveId = identity:
    builtins.hashString "sha256" (builtins.toJSON {
      inherit (identity) kind key;
    });

  # 全ベクタとの一致を検証する(true / throw)
  verifyVectors = vectorsFile:
    let
      data = builtins.fromJSON (builtins.readFile vectorsFile);
      check = v:
        let got = deriveId v.identity;
        in if got == v.expected then true
           else throw "niface id vector mismatch: got ${got} want ${v.expected} (canonical: ${v.canonical})";
    in builtins.all check data.vectors;
}
