{
  description = "niface — n-tools interface: JSON pipe specification for the Nix system-tools ecosystem";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    # id 導出(nix/lib.nix)の評価テスト。flake-parts モジュールが checks に載せる。
    nix-unit = {
      url = "github:nix-community/nix-unit";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      inherit systems;

      imports = [
        # nix-unit の評価テストを checks に載せる flake-parts モジュール。
        inputs.nix-unit.modules.flake.default
      ];

      perSystem =
        { pkgs, ... }:
        let
          # niface の Go 参照実装(CLI niface-validate)。定義は nix/package.nix に
          # 括り出し、packages.validate / flake.lib.mkSchemaCheck と共有する。
          niface-go = import ./nix/package.nix { inherit pkgs; };
        in
        {
          packages.validate = niface-go;

          # nix run .#validate -- envelope.json … 正の schema を store パスから既定注入。
          # 利用側は -schema で上書きできる(Go flag は後勝ち)。
          apps.validate = {
            type = "app";
            program = "${pkgs.writeShellScript "niface-validate" ''
              exec ${niface-go}/bin/validate -schema ${./schema/v1/envelope.schema.json} "$@"
            ''}";
          };

          checks = {
            # id 導出の Nix 実装がテストベクタと一致することを検証。
            id-vectors = pkgs.runCommand "niface-id-vectors"
              {
                passAsFile = [ "result" ];
                result = builtins.toJSON
                  ((import ./nix/lib.nix { }).verifyVectors ./testdata/v1/id-vectors.json);
              } ''
              [ "$(cat $resultPath)" = "true" ] && touch $out
            '';

            # Go 参照実装のビルド + テスト(id-vectors 通過 + testdata 適合検証)。
            # niface-go の build/checkPhase で go test ./... が走る。依存は
            # vendorHash で pin した FOD が取得するため vendor をコミットしない。
            go = niface-go;
          };

          # nix-unit: id 導出(nix/lib.nix)の値域・isAscii を評価テストで検証する。
          # sandbox 内で flake を再 import するため direct input を渡しオフライン評価。
          nix-unit.inputs = {
            inherit (inputs) nixpkgs flake-parts nix-unit;
          };
          nix-unit.tests = import ./nix/tests.nix {
            inherit (pkgs) lib;
            idLib = import ./nix/lib.nix { inherit (pkgs) lib; };
          };

          formatter = pkgs.nixfmt;
        };

      flake = {
        # niface 規格の参照実装(id 導出)。specVersion 1 の identity → id を導く。
        lib = import ./nix/lib.nix { inherit (nixpkgs) lib; };
      };
    };
}
