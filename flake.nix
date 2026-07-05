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
        {
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

            # testdata が schema に適合することを検証。
            schema = pkgs.runCommand "niface-schema-check"
              { nativeBuildInputs = [ pkgs.python3 pkgs.python3Packages.jsonschema ]; } ''
              python3 ${./scripts/validate.py} ${./schema/v1/envelope.schema.json} ${./testdata/v1}
              touch $out
            '';

            # Go 参照実装のテスト(id-vectors 通過)。
            go = pkgs.runCommand "niface-go-test"
              { nativeBuildInputs = [ pkgs.go ]; } ''
              export HOME=$TMPDIR GOFLAGS=-mod=mod
              cp -r ${./.} src && chmod -R +w src && cd src/go
              go test ./...
              touch $out
            '';
          };

          formatter = pkgs.nixfmt;
        };

      flake = {
        # niface 規格の参照実装(id 導出)。specVersion 1 の identity → id を導く。
        lib = import ./nix/lib.nix { inherit (nixpkgs) lib; };
      };
    };
}
