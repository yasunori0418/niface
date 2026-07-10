{
  description = "niface — n-tools interface: JSON pipe specification for the Nix system-tools ecosystem";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
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

      perSystem =
        { pkgs, ... }:
        let
          # niface の Go 参照実装。cmd/validate(CLI niface-validate)をビルドし、
          # build 時に go test ./...(id-vectors + testdata 適合検証)を走らせる。
          # 依存は vendorHash で pin した FOD が取得する(vendor はコミットしない)。
          niface-go = pkgs.buildGoModule {
            pname = "niface-validate";
            version = "0.1.0";
            src = ./.;
            modRoot = "go";
            vendorHash = "sha256-qVoj03LNLbdoCUAOydK7oEHsuZ1BZ6Z2jwYB3gPOfrw=";
            subPackages = [ "cmd/validate" ];
            doCheck = true;
            checkPhase = ''
              runHook preCheck
              go test ./...
              runHook postCheck
            '';
          };
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

          formatter = pkgs.nixfmt;
        };

      flake = {
        # niface 規格の参照実装(id 導出)。specVersion 1 の identity → id を導く。
        lib = import ./nix/lib.nix { inherit (nixpkgs) lib; };
      };
    };
}
