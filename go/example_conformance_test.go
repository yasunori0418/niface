package niface_test

// 適合ガイド(docs/guides/conformance.md)の go module 経路の Go 例の正本。
// ガイド本文からはここへリンクだけ張り、コードの二重管理を避ける。
//
// 目的は Envelope の 4 型引数 / SubjectResult / Result / Item の型引数 / DeriveID を
// 1 コードパスで参照させ、型が動いたら CI(go test)が**コンパイル追随**で落ちる形に
// することに絞る。`// Output:` を持たない Example は go test ではコンパイルはされるが
// 実行はされない(godoc / testing の公式挙動)。したがって実行時挙動の回帰は
// この Example では担保せず、既存の id_test.go / envelope_test.go / conformance の
// テストが担う分業とする。

import (
	"encoding/json"
	"os"
	"time"

	niface "github.com/yasunori0418/niface/go"
)

// PutInfo はツール固有情報の型。エンベロープの Info スロットに載せる(規格型は info 配下のみ拡張可)。
type PutInfo struct {
	Target string `json:"target"`
}

func ExampleEnvelope() {
	// StartedAt は実行開始時に 1 度取り、FinishedAt は組み立て直前に取り直す。
	// 実装コードで duration が意味を持つことを写経読者に示すため、両者を同じ
	// 値には潰さない。
	started := time.Now().Format(time.RFC3339)

	// item id は identity({kind, key})から機械導出する。値域外(spec §5)は error。
	id, err := niface.DeriveID(niface.Identity{Kind: "file", Key: "/etc/hosts"})
	if err != nil {
		panic(err)
	}

	// Envelope は 4 つの型引数 [TItem, TChange, TInfo, TEnvInfo] を取る。
	// 自ツールの info 型でパラメータ化して型付きで組み立てる。
	// startedAt / finishedAt / dryRun は必須。startedAt / finishedAt は RFC 3339
	// 形式でないと適合検証で弾かれる(spec §2)。DryRun は zero value で出力に
	// 含まれるが、必須であることを写経読者に示すため明示する。
	env := niface.Envelope[PutInfo, struct{}, struct{}, struct{}]{
		SpecVersion: 1,
		Tool:        niface.Tool{Name: "nput", Version: "0.9.0"},
		Command:     "apply",
		Status:      niface.StatusSuccess,
		DryRun:      false,
		StartedAt:   started,
		FinishedAt:  time.Now().Format(time.RFC3339),
		Results: []niface.SubjectResult[PutInfo, struct{}, struct{}]{
			{
				Subject:    niface.Subject{Name: "home"},
				Status:     niface.StatusSuccess,
				StartedAt:  started,
				FinishedAt: time.Now().Format(time.RFC3339),
				Result: niface.Result[PutInfo, struct{}, struct{}]{
					Items: []niface.Item[PutInfo]{
						{ID: id, Kind: "file", Status: niface.ItemSuccess, Info: PutInfo{Target: "/etc/hosts"}},
					},
				},
			},
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(env); err != nil {
		panic(err)
	}
}
