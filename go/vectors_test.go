package niface

import (
	"bytes"
	"encoding/json"
	"testing"
)

// IDVectorsV1 が UseNumber デコードで vectors / rejected を持つ表として読める
// ことを固定する。正本とのバイト一致は internal/spec 側で検査済み。
func TestIDVectorsV1Decodes(t *testing.T) {
	dec := json.NewDecoder(bytes.NewReader(IDVectorsV1()))
	dec.UseNumber()
	var vf vectorFile
	if err := dec.Decode(&vf); err != nil {
		t.Fatalf("id-vectors デコード: %v", err)
	}
	if len(vf.Vectors) == 0 || len(vf.Rejected) == 0 {
		t.Fatalf("vectors=%d rejected=%d — 空のベクタ表", len(vf.Vectors), len(vf.Rejected))
	}
}

// 防御的コピー(呼び出し側の mutate が次回呼び出しへ波及しない)は公開 API の
// 契約なので固定する。
func TestIDVectorsV1IsDefensiveCopy(t *testing.T) {
	v := IDVectorsV1()
	v[0] = 'X'
	if bytes.Equal(v, IDVectorsV1()) {
		t.Error("IDVectorsV1 の返り値の変更が次回呼び出しへ波及した")
	}
}
