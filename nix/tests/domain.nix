# nix-unit: deriveId / checkDomain の値域(spec §5)を検証する。
# 拒否系は builtins.tryEval の success で判定する(域外は throw)。
{ lib, idLib }:
let
  inherit (idLib) deriveId;
  # identity { kind = "n"; key = <key>; } を導出したときの成否(true=受理)。
  accepts = key: (builtins.tryEval (deriveId { kind = "n"; inherit key; })).success;
in
{
  # 境界 2^53−1 の受理と既知 hash(Go と一致・言語間互換の要)。
  testAcceptMaxSafeInt = {
    expr = deriveId {
      kind = "maxint";
      key = 9007199254740991;
    };
    expected = "6d43a7e0ef4ce0d93e02b79571ceb9165144a9383c96a8bcbaa7bb0c0f83103e";
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
}
