# mattpocock/skills を niface repo root の .claude/skills/ へ配置する nput の
# project mode config を flake-parts module として切り出す。
# nput の flakeModules.default（dev/flake.nix の imports が読む）を前提に、
# perSystem.nput.skills へ manifest を宣言する。
#
# `nput apply skills -f <dev flake>` でビルドし、各 skill を .claude/skills/<name>
# へ store-symlink 配置する。root = projectRoot（git toplevel）なので配置先は
# repo root 配下。配置物は .gitignore 済み（.claude/skills/*）の ephemeral。
{ inputs, ... }:
let
  nputLib = inputs.nput.lib;

  # 展開する skill を明示列挙する（mattpocock/skills の skills/ 配下の相対パス）。
  # nput のドッグフーディング構成と同じ集合を配置する。
  skillSubpaths = [
    "engineering/grill-with-docs"
    "engineering/improve-codebase-architecture"
    "engineering/prototype"
    "engineering/setup-matt-pocock-skills"
    "engineering/tdd"
    "engineering/to-issues"
    "engineering/to-prd"
    "engineering/triage"
    "productivity/grilling"
    "productivity/handoff"
  ];

  # skill ごとに { ".claude/skills/<name>" = entry; } を組む。
  # target = .claude/skills/<skill 名>、配置元は skills/<category>/<name> の subpath。
  skillEntries = builtins.listToAttrs (
    map (p: {
      name = ".claude/skills/${baseNameOf p}";
      value = {
        src = inputs.matt-skills;
        subpath = "skills/${p}";
      };
    }) skillSubpaths
  );
in
{
  perSystem =
    { pkgs, ... }:
    {
      # perSystem.nput.skills → flake.nput.<system>.skills へ自動転置される（nput flakeModule）。
      nput.skills = nputLib.mkManifest {
        inherit pkgs;
        root = nputLib.projectRoot;
        entries = skillEntries;
      };
    };
}
