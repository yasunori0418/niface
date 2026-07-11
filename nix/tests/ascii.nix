# nix-unit: isAscii の判定を検証する(nix/lib.nix のロケール前提を担保する)。
# 高位バイト(≥0x80)を非 ASCII として false にすることが deriveId の
# メンバー名検査の正しさ・Go(バイト単位検査)との対称性の土台。
{ lib, idLib }:
let
  inherit (idLib) isAscii;
in
{
  testAsciiPlain = {
    expr = isAscii "abc_123.service";
    expected = true;
  };
  testAsciiEmpty = {
    expr = isAscii "";
    expected = true;
  };
  # 制御文字(tab = 0x09)は ASCII 域。
  testAsciiControl = {
    expr = isAscii "a\tb";
    expected = true;
  };
  # 非 ASCII: BMP(アクセント・日本語)/ BMP 外(絵文字)。
  testRejectAccent = {
    expr = isAscii "café";
    expected = false;
  };
  testRejectBmp = {
    expr = isAscii "キー";
    expected = false;
  };
  testRejectAstral = {
    expr = isAscii "🔑";
    expected = false;
  };
}
