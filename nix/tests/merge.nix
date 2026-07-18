# nix-unit: アグリゲータ(nix/tests.nix)の重複キー検出 mergeDistinct を検証する。
# nix-unit の test attrset 内で実際にキーを重複させると評価全体が throw して自己
# 検証にならないため、ヘルパ(./tests/_lib.nix)を直接 import し tryEval で突く。
# leaf は { idLib } で import されるが本テストは idLib を使わないため `...` で無視する。
{ ... }:
let
  inherit (import ./_lib.nix) mergeDistinct;
  # foldl' の畳み込みと同じ経路でモジュール列をマージした成否(true=衝突なし)。
  merges = modules: (builtins.tryEval (builtins.foldl' mergeDistinct { } modules)).success;
in
{
  # 同名キーを持つモジュール同士は後勝ちで消さず throw する(検出が発火する)。
  testMergeRejectsDuplicate = {
    expr = merges [
      { testFoo = 1; }
      { testFoo = 2; }
    ];
    expected = false;
  };

  # キーが重ならなければマージは成功する(検出が誤発火しない)。
  testMergeAcceptsDistinct = {
    expr = merges [
      { testFoo = 1; }
      { testBar = 2; }
    ];
    expected = true;
  };

  # 非隣接の衝突(1 番目と 3 番目)も検出する。直前モジュールとだけ比較する
  # 誤実装ではなく、累積 acc 全体と照合していることを固定する。
  testMergeRejectsNonAdjacentDuplicate = {
    expr = merges [
      { testFoo = 1; }
      { testBar = 2; }
      { testFoo = 3; }
    ];
    expected = false;
  };

  # 単一モジュールは初期値 { } との畳み込み 1 回で成功する(下限直上で誤発火しない)。
  testMergeAcceptsSingle = {
    expr = merges [ { testFoo = 1; } ];
    expected = true;
  };

  # 空のモジュール列は初期値 { } を返して成功する(下限で誤発火しない)。
  testMergeAcceptsEmpty = {
    expr = merges [ ];
    expected = true;
  };
}
