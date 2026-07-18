package niface

import (
	"bytes"
	"testing"
)

// IDVectorsV1 の返り値を入力に据えて、consumer が辿る経路(embed bytes →
// UseNumber デコード → DeriveID → 期待値照合)を固定する。正本を読む
// id_test.go と入力源だけが違い、照合ロジックは assertVectorFile
// (id_test.go)を共有する。
func TestIDVectorsV1PassesAllVectors(t *testing.T) {
	assertVectorFile(t, IDVectorsV1())
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
