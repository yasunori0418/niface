# nix-unit アグリゲータ: `nix/tests/` 配下の全 leaf `*.nix`(`_` prefix を除く)を
# `{ idLib }` で import し `//` マージした単一の test attrset を返す。leaf なテスト
# ファイルを追加するだけで自動的に nix-unit に載る(集約ファイル・flake.nix の編集不要)。
# `_` prefix のファイル(例: `_lib.nix`)はテストではなく共有ヘルパで、readDir 対象から
# 除外する。テスト名はファイル横断で一意であること。重複した場合は `//` の後勝ちで黙って
# 消えないよう、マージ時に重複キーを検出して throw する(mergeDistinct、→ ./tests/_lib.nix)。
# idLib = import ./lib.nix { inherit lib; }。
{ lib, idLib }:
let
  dir = ./tests;
  inherit (import (dir + "/_lib.nix")) mergeDistinct;
  # leaf テストファイルのみ拾う(`_` prefix のヘルパ・非 .nix は除外)。
  testFiles = lib.filterAttrs
    (name: type: type == "regular" && lib.hasSuffix ".nix" name && !lib.hasPrefix "_" name)
    (builtins.readDir dir);
  modules = lib.mapAttrsToList (name: _type: import (dir + "/${name}") { inherit idLib; }) testFiles;
in
lib.foldl' mergeDistinct { } modules
