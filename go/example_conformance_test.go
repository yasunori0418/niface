package niface_test

// 適合ガイド(docs/guides/conformance.md)の go module 経路の Go 例の正本。
// ガイド本文からはここへリンクだけ張り、コードの二重管理を避ける。
//
// 目的は Envelope の 4 型引数 / SubjectResult / Result / Item の型引数 / DeriveID を
// 1 コードパスで参照させ、型が動いたら CI(go test)が**コンパイル追随**で落ちる形に
// することに絞る。`// Output:` を持つ Example は go test で出力比較まで走るため、
// 未使用の info スロット(TInfo / TEnvInfo)を `*struct{}` にすることで `"info":{}` が
// envelope 直下・result 内の双方から消える形も同時に固定する。

import (
	"encoding/json"
	"os"

	niface "github.com/yasunori0418/niface/go"
)

// PutInfo はツール固有情報の型。エンベロープの Info スロットに載せる(規格型は info 配下のみ拡張可)。
type PutInfo struct {
	Target string `json:"target"`
}

// GenerationsInfo は list-generations 相当の Envelope.Info。
// current は omitempty を付けないため、zero value(0)でも出力に含まれる。
type GenerationsInfo struct {
	Current int `json:"current"`
}

// InitInfo は init 相当の Envelope.Info。
type InitInfo struct {
	Profile string `json:"profile"`
}

func ExampleEnvelope() {
	// 未使用の型パラメータ(TChange / TInfo / TEnvInfo)は非ポインタ `struct{}` ではなく
	// nil ポインタ型 `*struct{}` を使う。TInfo / TEnvInfo は `info` フィールドの型で
	// `omitempty` が付くが、非ポインタ struct には omitempty が効かず、`struct{}` を
	// 使うと envelope 直下と result 内の双方に `"info":{}` が出力されて意図しない
	// 出力契約になる。schema 違反ではないが「未使用スロットは出力に現れない」自然な
	// 期待を裏切るため、nil ポインタ型で塞ぐ。TChange は `Result.Changes` 自体が
	// omitempty のため出力には現れないが、他スロットと揃えて `*struct{}` にする。

	// item id は identity({kind, key})から機械導出する。値域外(spec §5)は error。
	id, err := niface.DeriveID(niface.Identity{Kind: "file", Key: "/etc/hosts"})
	if err != nil {
		panic(err)
	}

	// startedAt / finishedAt は Example 出力の再現性を保つため固定値にしている。
	// 実装コードでは実行開始時と組み立て直前の 2 回 time.Now().Format(time.RFC3339)
	// を呼び、duration を保つ。RFC 3339 形式でないと適合検証で弾かれる(spec §2)。
	// DryRun は zero value で出力に含まれるが、必須であることを写経読者に示すため
	// 明示する。
	const startedAt = "2026-07-22T09:00:00+09:00"
	const finishedAt = "2026-07-22T09:00:01+09:00"

	// Envelope は 4 つの型引数 [TItem, TChange, TInfo, TEnvInfo] を取る。
	// 自ツールの info 型でパラメータ化して型付きで組み立てる。
	env := niface.Envelope[PutInfo, *struct{}, *struct{}, *struct{}]{
		SpecVersion: 1,
		Tool:        niface.Tool{Name: "nput", Version: "0.9.0"},
		Command:     "apply",
		Status:      niface.StatusSuccess,
		DryRun:      false,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		Results: []niface.SubjectResult[PutInfo, *struct{}, *struct{}]{
			{
				Subject:    niface.Subject{Name: "home"},
				Status:     niface.StatusSuccess,
				StartedAt:  startedAt,
				FinishedAt: finishedAt,
				Result: niface.Result[PutInfo, *struct{}, *struct{}]{
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
	// Output の `id` は DeriveID(sha256(JCS(identity))) の実測値。ここが失敗した場合、
	// id 導出経路(JCS / sha256 / identity 値域)の変更を疑う。本命の担保は id_test.go
	// / vectors_test.go(id-vectors 経由)にある。
	// Output:
	// {"specVersion":1,"tool":{"name":"nput","version":"0.9.0"},"command":"apply","status":"success","dryRun":false,"startedAt":"2026-07-22T09:00:00+09:00","finishedAt":"2026-07-22T09:00:01+09:00","results":[{"subject":{"name":"home"},"status":"success","startedAt":"2026-07-22T09:00:00+09:00","finishedAt":"2026-07-22T09:00:01+09:00","result":{"items":[{"id":"c5ec86d4056d73123e43df3257b6e36ba12e396ec4b0a8a63a7ba74429705ed6","kind":"file","status":"success","info":{"target":"/etc/hosts"}}]}}]}
}

// ExampleEnvelope_multiCommand は multi-command producer(1 tool 内で command ごとに
// info 型が異なるケース)の型消去パターンを示す。1 回の実行では command が 1 つに
// 定まり info 型も静的に決まる。ただし複数 command の結果をプロセス全域で 1 つの
// 変数に貯めようとすると、command ごとに Envelope の型パラメータが異なるため
// 単一の静的 generic 型で宣言できない、という producer 実装上の課題がある。
//
// niface に新 API を足す必要はない。`Envelope` は既に `MarshalJSON` を実装しており、
// 型パラメータの異なる Envelope 値も Go 標準の `json.Marshaler` として一様に扱える。
//
// TItem は `Items` が非 omitempty で常に出るスロットのため、item を持たない command
// では `*struct{}` にする必要はないが、この Example では他スロットと揃えて統一する。
// 加えて list-generations 側は Item[*struct{}] + DeriveID の型追随を新 Example でも
// 1 コードパスで踏むため、item を 1 件入れている。
func ExampleEnvelope_multiCommand() {
	const startedAt = "2026-07-22T09:00:00+09:00"
	const finishedAt = "2026-07-22T09:00:01+09:00"

	// list-generations: Envelope.Info を GenerationsInfo で型付け、他スロットは *struct{}。
	// 型追随のため Item[*struct{}] を 1 件入れる(kind=generation, key=home/42)。
	genID, err := niface.DeriveID(niface.Identity{Kind: "generation", Key: "home/42"})
	if err != nil {
		panic(err)
	}
	listGens := niface.Envelope[*struct{}, *struct{}, *struct{}, GenerationsInfo]{
		SpecVersion: 1,
		Tool:        niface.Tool{Name: "nput", Version: "0.9.0"},
		Command:     "list-generations",
		Status:      niface.StatusSuccess,
		DryRun:      false,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		Info:        GenerationsInfo{Current: 42},
		Results: []niface.SubjectResult[*struct{}, *struct{}, *struct{}]{
			{
				Subject:    niface.Subject{Name: "home"},
				Status:     niface.StatusSuccess,
				StartedAt:  startedAt,
				FinishedAt: finishedAt,
				Result: niface.Result[*struct{}, *struct{}, *struct{}]{
					Items: []niface.Item[*struct{}]{
						{ID: genID, Kind: "generation", Status: niface.ItemSuccess},
					},
				},
			},
		},
	}

	// init: Envelope.Info を InitInfo で型付け、他スロットは *struct{}。item は無し。
	initCmd := niface.Envelope[*struct{}, *struct{}, *struct{}, InitInfo]{
		SpecVersion: 1,
		Tool:        niface.Tool{Name: "nput", Version: "0.9.0"},
		Command:     "init",
		Status:      niface.StatusSuccess,
		DryRun:      false,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		Info:        InitInfo{Profile: "home"},
		Results: []niface.SubjectResult[*struct{}, *struct{}, *struct{}]{
			{
				Subject:    niface.Subject{Name: "home"},
				Status:     niface.StatusSuccess,
				StartedAt:  startedAt,
				FinishedAt: finishedAt,
				Result:     niface.Result[*struct{}, *struct{}, *struct{}]{Items: []niface.Item[*struct{}]{}},
			},
		},
	}

	// process 全域の変数は Envelope の型パラメータが command ごとに異なるため単一
	// 静的 generic 型で宣言できない。`Envelope.MarshalJSON` があるので、両者を
	// `json.Marshaler`(= 非 generic な interface)へ束ねると型消去できる。
	// 順序を固定するためスライスに詰めて 1 件ずつ marshal する。
	envelopes := []json.Marshaler{listGens, initCmd}
	for _, env := range envelopes {
		if err := json.NewEncoder(os.Stdout).Encode(env); err != nil {
			panic(err)
		}
	}
	// Output:
	// {"specVersion":1,"tool":{"name":"nput","version":"0.9.0"},"command":"list-generations","status":"success","dryRun":false,"startedAt":"2026-07-22T09:00:00+09:00","finishedAt":"2026-07-22T09:00:01+09:00","info":{"current":42},"results":[{"subject":{"name":"home"},"status":"success","startedAt":"2026-07-22T09:00:00+09:00","finishedAt":"2026-07-22T09:00:01+09:00","result":{"items":[{"id":"956954835debb292ce0463552f17d91098b9003b48af2b11ca5c7c46041d1225","kind":"generation","status":"success"}]}}]}
	// {"specVersion":1,"tool":{"name":"nput","version":"0.9.0"},"command":"init","status":"success","dryRun":false,"startedAt":"2026-07-22T09:00:00+09:00","finishedAt":"2026-07-22T09:00:01+09:00","info":{"profile":"home"},"results":[{"subject":{"name":"home"},"status":"success","startedAt":"2026-07-22T09:00:00+09:00","finishedAt":"2026-07-22T09:00:01+09:00","result":{"items":[]}}]}
}
