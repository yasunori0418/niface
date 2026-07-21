package conformance_test

// 適合ガイド(docs/guides/conformance.md)の「自ツールのテストで適合検証と id
// 整合を固定する」節の Go 例の正本。ガイドからはここへリンクだけ張る。
//
// 目的は NewDefaultChecker / (*Checker).Check のシグネチャに CI が**コンパイル
// 追随**することに絞る。`// Output:` を持たない Example は go test で実行はされない
// ため、envelopeJSON リテラルの schema 適合そのものはここでは担保しない。schema
// による全件検査は既存の conformance_test.go / default_test.go(TestNewDefaultChecker*)
// が担う。
//
// Example は Example 慣例に沿って `panic(err)` を使うが、自ツールのテストへ組み
// 込むときは `t.Fatal(err)` / `t.Errorf("niface 不適合: %v", findings)` に置き換える。

import (
	"fmt"

	"github.com/yasunori0418/niface/go/conformance"
)

func ExampleChecker_Check() {
	// 適合検証: 自ツールの出力(適合しているべき正のサンプル)が
	// 規格 schema + lint を通ることをテストで固定する。
	chk, err := conformance.NewDefaultChecker()
	if err != nil {
		panic(err)
	}
	envelopeJSON := []byte(`{
  "specVersion": 1,
  "tool": {"name": "nput", "version": "0.9.0"},
  "command": "apply",
  "status": "success",
  "dryRun": false,
  "startedAt": "2026-07-05T12:34:56+09:00",
  "finishedAt": "2026-07-05T12:34:56+09:00",
  "results": []
}`)
	if findings := chk.Check(envelopeJSON); len(findings) > 0 {
		fmt.Printf("niface 不適合: %v\n", findings)
	}
}
