# niface の Go 参照実装(CLI niface-validate)を buildGoModule でビルドする関数。
#
# cmd/validate をビルドし、build 時に go test ./...(id-vectors + testdata 適合
# 検証)を走らせる。依存は vendorHash で pin した FOD が取得する(vendor は
# コミットしない)。src は repo 全体(../.. = nix/pkgs から見た repo 根)を渡し
# modRoot = "go" で module を指す。go test は ../testdata・../schema を参照するため
# src はサブディレクトリではなく repo 全体である必要がある。
#
# perSystem の packages.validate と flake.lib.mkSchemaCheck の両方がこの関数を
# 共有し、検証器の呼び出し規約・依存(vendorHash)を niface 側に閉じ込める。
{ pkgs }:
pkgs.buildGoModule {
  pname = "niface-validate";
  version = "0.1.0";
  src = ../..;
  modRoot = "go";
  vendorHash = "sha256-qVoj03LNLbdoCUAOydK7oEHsuZ1BZ6Z2jwYB3gPOfrw=";
  subPackages = [ "cmd/validate" ];
  doCheck = true;
  checkPhase = ''
    runHook preCheck
    go test ./...
    runHook postCheck
  '';
}
