package niface

import (
	"bytes"
	"encoding/json"
	"testing"
)

// IDVectorsV1 の返り値を入力に据えて、consumer が辿る経路(embed bytes →
// UseNumber デコード → DeriveID → 期待値照合)を 1 本で固定する。正本を読む
// id_test.go と対で、公開 API 経由でも全ベクタ・全 rejected が同じ結果になる
// ことを直接検証する(vectorFile 型は id_test.go で定義)。
func TestIDVectorsV1PassesAllVectors(t *testing.T) {
	dec := json.NewDecoder(bytes.NewReader(IDVectorsV1()))
	dec.UseNumber()
	var vf vectorFile
	if err := dec.Decode(&vf); err != nil {
		t.Fatalf("id-vectors デコード: %v", err)
	}
	if len(vf.Vectors) == 0 || len(vf.Rejected) == 0 {
		t.Fatalf("vectors=%d rejected=%d — 空のベクタ表", len(vf.Vectors), len(vf.Rejected))
	}
	for i, v := range vf.Vectors {
		got, err := DeriveID(Identity{Kind: v.Identity.Kind, Key: v.Identity.Key})
		if err != nil {
			t.Errorf("vector %d: %v", i, err)
			continue
		}
		if got != v.Expected {
			t.Errorf("vector %d: got %s want %s (canonical: %s)", i, got, v.Expected, v.Canonical)
		}
	}
	for i, r := range vf.Rejected {
		if _, err := DeriveID(Identity{Kind: r.Identity.Kind, Key: r.Identity.Key}); err == nil {
			t.Errorf("rejected vector %d: expected error but got none (reason: %s)", i, r.Reason)
		}
	}
}

// 防御的コピー(呼び出し側の mutate が次回呼び出しへ波及しない)は公開 API の
// 契約なので固定する。
func TestIDVectorsV1IsDefensiveCopy(t *testing.T) {
	v := IDVectorsV1()
	if len(v) == 0 {
		t.Fatal("IDVectorsV1 が空を返した")
	}
	v[0] = 'X'
	if bytes.Equal(v, IDVectorsV1()) {
		t.Error("IDVectorsV1 の返り値の変更が次回呼び出しへ波及した")
	}
}
