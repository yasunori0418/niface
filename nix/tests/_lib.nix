# nix-unit アグリゲータ(nix/tests.nix)と、その検証テスト(nix/tests/merge.nix)が
# 共有するヘルパ。`_` prefix によりアグリゲータの readDir から除外され、leaf テスト
# としては読み込まれない(→ nix/tests.nix のフィルタ)。lib 非依存にするため
# builtins のみで実装する。
{
  # 2 つの attrset を `//` マージする。既存キーと衝突したら後勝ちで黙って消さず、
  # 重複キー名を挙げて throw する。foldl' の畳み込み関数として使う。
  mergeDistinct = acc: m:
    let dup = builtins.intersectAttrs acc m;
    in if dup != { } then
      throw "niface tests: duplicate test name(s) across nix/tests/*.nix: ${builtins.concatStringsSep ", " (builtins.attrNames dup)}"
    else acc // m;
}
