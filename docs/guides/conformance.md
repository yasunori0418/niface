# 適合ガイド

自作ツールを niface 適合にするための最小手順をまとめる。本書は指針であり規範ではない。適合の定義は `README.md` の Conformance 節(schema 検証 / id-vectors 通過 / standalone の 3 点)を、standalone の規律は `spec/v1/spec.md` §8 を正とする。

参照方法は 2 経路ある。両方を使うことも、片方だけを使うこともできる。

- **flake 経路**: niface を flake input に取り、CI(`nix flake check`)で自ツールの出力サンプルを規格 schema で検証し、id 導出を id-vectors で固定する。
- **go module 経路**: Go 製ツールが `github.com/yasunori0418/niface/go` を import し、エンベロープ型を共有(コンパイル時)して item id を `DeriveID` で導出する。

リリースの参照点はタグで指す。同一コミットに `v1.N.P`(規格スナップショット・flake input / 人間用)と `go/v1.N.P`(同一コミットの Go module 用ミラー)を対で打つ。Go module はサブディレクトリ(`go/`)にあり prefix 無しの root タグが Go ツールチェーンから見えないため、2 系統のタグが要る。詳細は `docs/design.md` の「バージョニングとリリースタグ」節(→ ADR-0026)。

## flake 経路

niface を input に追加し、`lib.mkSchemaCheck` と `lib.verifyVectors` を checks に組み込む。

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    # リリースタグで固定する(未指定なら追従)。例: github:yasunori0418/niface/v1.0.0
    niface.url = "github:yasunori0418/niface/v1.0.0";
  };

  outputs =
    { nixpkgs, niface, ... }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      checks.${system} = {
        # 自ツールの「適合しているべき正のサンプル」を規格 schema で検証する。
        # 検証器(niface-validate)の呼び出し規約・依存は niface 側に閉じているため、
        # 呼び出しは pkgs と testdataDir を渡すだけで済む。testdataDir 配下の *.json は
        # 全件が適合前提なので、不適合サンプルも持つなら valid 側だけを渡す。
        schema = niface.lib.mkSchemaCheck {
          inherit pkgs;
          testdataDir = ./testdata/valid;
        };

        # id 導出を Nix 側で niface.lib.deriveId に委ねる場合、その導出が
        # id-vectors 全ベクタで期待値と一致することを固定する(不一致は throw)。
        # ベクタ表は niface input に同梱されたものを参照する。
        id-vectors =
          pkgs.runCommand "id-vectors"
            {
              result = builtins.toJSON (
                niface.lib.verifyVectors "${niface}/testdata/v1/id-vectors.json"
              );
            }
            ''
              [ "$result" = "true" ] && touch $out
            '';
      };
    };
}
```

`mkSchemaCheck` は `testdataDir` 配下の `*.json` を再帰的に集め、全て規格 schema を通過すれば check が成功する。1 件でも違反すると derivation が失敗し、違反内容が build ログに出る。したがって渡すディレクトリには「規格に適合しているべき正のサンプル」だけを置く(`README.md` の valid サンプルと同じ位置付け)。niface 自身のように valid / invalid を分けて持つ場合は、上の例のように valid 側のディレクトリだけを `testdataDir` に渡す。invalid を混ぜると再帰収集で拾われ check が常に失敗する。

id 導出を Go 側(go module 経路)で行うツールは、id-vectors の検証は go module のテストが担うため、`verifyVectors` の check は不要。

## go module 経路

Go 製ツールは go module を取り込み、エンベロープ型を共有して item id を導出する。

```sh
# go/vX.Y.Z タグを参照する。Go はサブディレクトリ module のタグを @vX.Y.Z で解決する
# (内部的に go/vX.Y.Z を探す。prefix は Go ツールチェーンの制約 → ADR-0026)。
go get github.com/yasunori0418/niface/go@v1.0.0
```

構築例のコードは testable Example として go/ 配下に置き、go test でコンパイル追随を CI に強制する(型が動けば `nix flake check` が落ちる)。ドキュメントとコードの二重管理を避けるため、写経元はこちらを一次参照とする。

- `go/example_conformance_test.go` の `ExampleEnvelope`([source](../../go/example_conformance_test.go) / [pkg.go.dev](https://pkg.go.dev/github.com/yasunori0418/niface/go#example-Envelope))

`Envelope` は 4 つの型引数 `[TItem, TChange, TInfo, TEnvInfo]` を取り、自ツールの info 型でパラメータ化して型付きで組み立てる。`startedAt` / `finishedAt` / `dryRun` は必須(spec §2)で、`startedAt` / `finishedAt` は RFC 3339 形式でないと適合検証で弾かれる。`dryRun` は bool の zero value でも出力に含まれる必要があるため、Go 側で明示せず zero value に頼るなら `omitempty` を付けないこと。item id は identity(`{kind, key}`)から `DeriveID` で機械導出する。値域外(spec §5)は error。

ツールを知らない消費側は info を `json.RawMessage` でパラメータ化(`niface.Envelope[json.RawMessage, json.RawMessage, json.RawMessage, json.RawMessage]`)し、規格部分だけを型安全に扱える。

### 自ツールのテストで適合検証と id 整合を固定する

schema と id-vectors は go module に embed されている(正本とのバイト完全一致を niface 側の CI が保証)。consumer は module 依存だけで適合検証・id 整合検証をテストへ組み込め、schema や id-vectors を手動で vendored copy する必要はない。

検証コードの写経元は同様に go/ 配下の Example にある。

- `go/conformance/example_test.go` の `ExampleChecker_Check`([source](../../go/conformance/example_test.go) / [pkg.go.dev](https://pkg.go.dev/github.com/yasunori0418/niface/go/conformance#example-Checker.Check))

`NewDefaultChecker` は embed 済みの正本 schema をコンパイルした `Checker` を返す。`Check(envelopeJSON)` は違反メッセージのスライスを返し、空スライスなら適合。Example は Example 慣例に沿って `panic(err)` を使うため、自ツールのテスト関数へ組み込むときは `t.Fatal(err)` / `t.Errorf("niface 不適合: %v", findings)` に置き換える。

schema の生 bytes が要る場合は `conformance.SchemaV1()`、id-vectors の生 bytes は `niface.IDVectorsV1()`(補助 API)。id 導出を `DeriveID` に委ねるツールは id-vectors の検証が module のテストで担われているため不要で、`IDVectorsV1` は id 導出を自前実装する場合の整合固定に使う。デコード時の注意(UseNumber 必須)は godoc を参照。

## 単発検証

エンベロープ 1 件を手元で検証する口として、niface が提供する `validate` app を使える。ツール側リポジトリに何も足さずに CI ログや手元で使える。

```sh
# ファイルを渡す
nix run github:yasunori0418/niface#validate -- path/to/envelope.json

# パイプで stdin から渡す(自ツールの出力をそのまま流す)
your-tool apply | nix run github:yasunori0418/niface#validate

# niface リポジトリ内なら flake ref を省略できる
nix run .#validate -- path/to/envelope.json
```

正の schema はこの app が store パスから既定注入するため、`-schema` の指定は不要(上書きしたいときのみ Go flag として渡せる)。

## ツール固有 info schema は各ツールで管理する

規格が閉じるのは結果エンベロープの規格型のみで、`info` 配下の中身(ツール固有フィールド)は各ツールが自リポジトリで公開・管理する(→ ADR-0007 / ADR-0014)。niface 側に info schema を持ち込まない。これにより規格が個々のツールのリリースサイクルに引きずられず、ツールは規格型を `additionalProperties: false` で閉じたまま固有情報を `info` に載せられる。
