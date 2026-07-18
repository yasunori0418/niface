# nix-unit アグリゲータ: `nix/tests/` 配下の全 `*.nix` を `{ lib, idLib }` で import し
# `//` マージした単一の test attrset を返す。leaf なテストファイルを追加するだけで
# 自動的に nix-unit に載る(集約ファイル・flake.nix の編集不要)。テスト名はファイル
# 横断で一意であること。重複した場合は `//` の後勝ちで黙って消えないよう、マージ時に
# 重複キーを検出して throw する。idLib = import ./lib.nix { inherit lib; }。
{ lib, idLib }:
let
  dir = ./tests;
  testFiles = lib.filterAttrs (name: type: type == "regular" && lib.hasSuffix ".nix" name) (
    builtins.readDir dir
  );
  modules = lib.mapAttrsToList (name: _type: import (dir + "/${name}") { inherit lib idLib; }) testFiles;
  # 各モジュールを順にマージする。既存キーと衝突したら後勝ちで消さず throw する。
  mergeDistinct = acc: m:
    let dup = builtins.intersectAttrs acc m;
    in if dup != { } then
      throw "niface tests: duplicate test name(s) across nix/tests/*.nix: ${lib.concatStringsSep ", " (builtins.attrNames dup)}"
    else acc // m;
in
lib.foldl' mergeDistinct { } modules
