{
  description = "niface development environment";

  inputs = {
    root.url = "path:../";
    nixpkgs.follows = "root/nixpkgs";
    flake-parts.follows = "root/flake-parts";

    # フェッチ済みリポジトリを任意パスへ配置する nput（→ https://github.com/yasunori0418/nput）。
    # niface の開発環境はこの nput を使って mattpocock/skills を .claude/skills/ へ配置する。
    nput = {
      url = "github:yasunori0418/nput";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-parts.follows = "flake-parts";
    };

    # Claude Code 用スキル集（mattpocock/skills）。nput の project mode で配置する
    # ため flake=false。flake.lock が rev を pin する。
    matt-skills = {
      url = "github:mattpocock/skills";
      flake = false;
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
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
        # nput の flake-parts module。perSystem.nput.<name> を
        # flake.nput.<system>.<name> へ転置し、CLI から addressable にする。
        inputs.nput.flakeModules.default
        # mattpocock/skills 配置 config（perSystem.nput.skills）を切り出す。
        ./nput.nix
      ];
      perSystem =
        { inputs', pkgs, ... }:
        {
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              # Nix
              nixd
              statix
              inputs'.root.formatter
              # 規格の参照実装・検証
              go
              gopls
              python3
              python3Packages.jsonschema
              # フェッチ済みリポジトリ配置（pin 版）
              inputs'.nput.packages.nput
            ];
            shellHook = ''
              export REPO_ROOT=$(git rev-parse --show-toplevel)
              # mattpocock/skills を .claude/skills/ へ dogfood 配置する（project mode）。
              # 競合時は待たず skip（--no-wait）し no-op。
              nput apply skills -f "$REPO_ROOT/dev" --no-wait
            '';
          };
        };
    };
}
