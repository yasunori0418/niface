# nix-unit: deriveId / checkDomain の値域(spec §5)を検証する。
# 拒否系は builtins.tryEval の success で判定する(域外は throw)。
{ idLib }:
let
  inherit (idLib) deriveId;
  # identity { kind = "n"; key = <key>; } を導出したときの成否(true=受理)。
  accepts = key: (builtins.tryEval (deriveId { kind = "n"; inherit key; })).success;
in
{
  # 境界 2^53−1 の受理。既知 hash による言語間互換の固定は testdata
  # (verifyVectors / checks.id-vectors)へ一本化し、ここは受理判定に留める。
  testAcceptMaxSafeInt = {
    expr = accepts 9007199254740991;
    expected = true;
  };

  # 有効な整数キーは受理する。
  testAcceptSmallInt = {
    expr = accepts 1;
    expected = true;
  };
  testAcceptZero = {
    expr = accepts 0;
    expected = true;
  };
  testAcceptNegMaxSafeInt = {
    expr = accepts (-9007199254740991);
    expected = true;
  };

  # 浮動小数点(float 型)は表記を問わず域外。
  testRejectFractional = {
    expr = accepts 1.5;
    expected = false;
  };
  testRejectIntegerValuedFloat = {
    expr = accepts 1.0;
    expected = false;
  };
  testRejectExponent = {
    expr = accepts 1.0e3;
    expected = false;
  };

  # 安全整数域外の整数は拒否。
  testRejectAboveMax = {
    expr = accepts 9007199254740992;
    expected = false;
  };
  testRejectBelowMin = {
    expr = accepts (-9007199254740992);
    expected = false;
  };

  # 非 ASCII のメンバー名(BMP / BMP 外)は域外。値の非 ASCII は許す。
  testRejectNonAsciiKeyBmp = {
    expr = accepts { "キー" = "v"; };
    expected = false;
  };
  testRejectNonAsciiKeyAstral = {
    expr = accepts {
      items = [ { "🔑" = 1; } ];
    };
    expected = false;
  };
  testAcceptNonAsciiValue = {
    expr = accepts { target = "メモ/日記.md"; };
    expected = true;
  };

  # 規格の値域にない型(関数など)は域外。実 JSON(fromJSON 由来)からは
  # 到達しないが、checkDomain の `unsupported type` 分岐を突く。
  testRejectUnsupportedType = {
    expr = accepts (x: x);
    expected = false;
  };
}
