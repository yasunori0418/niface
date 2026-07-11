# nix-unit アグリゲータ: `nix/tests/` 配下の全 `*.nix` を `{ lib, idLib }` で import し
# `//` マージした単一の test attrset を返す。leaf なテストファイルを追加するだけで
# 自動的に nix-unit に載る(集約ファイル・flake.nix の編集不要)。テスト名はファイル
# 横断で一意であること(`//` は後勝ち)。idLib = import ./lib.nix { inherit lib; }。
{ lib, idLib }:
let
  dir = ./tests;
  testFiles = lib.filterAttrs (name: type: type == "regular" && lib.hasSuffix ".nix" name) (
    builtins.readDir dir
  );
  modules = lib.mapAttrsToList (name: _type: import (dir + "/${name}") { inherit lib idLib; }) testFiles;
in
lib.foldl' (acc: m: acc // m) { } modules
