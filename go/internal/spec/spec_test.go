package spec

import (
	"bytes"
	"os"
	"testing"
)

// embed コピーと正本のバイト完全一致を検査する。意味的一致ではなくバイト一致に
// するのは、consumer に届くのは bytes そのものであり、字面の乖離($id 欠落・
// 空白差等)も見逃さないため。不一致は go generate ./... の叩き忘れを意味する。

func TestSchemaV1MatchesSource(t *testing.T) {
	source, err := os.ReadFile("../../../schema/v1/envelope.schema.json")
	if err != nil {
		t.Fatalf("正本 schema 読み込み: %v", err)
	}
	if !bytes.Equal(SchemaV1(), source) {
		t.Error("embed した schema が正本 schema/v1/envelope.schema.json とバイト一致しない。go generate ./... で同期すること")
	}
}

func TestIDVectorsV1MatchesSource(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/v1/id-vectors.json")
	if err != nil {
		t.Fatalf("正本 id-vectors 読み込み: %v", err)
	}
	if !bytes.Equal(IDVectorsV1(), source) {
		t.Error("embed した id-vectors が正本 testdata/v1/id-vectors.json とバイト一致しない。go generate ./... で同期すること")
	}
}

// 返り値が防御的コピーであること(呼び出し側の mutate が以降の呼び出しに
// 波及しないこと)は公開 API の契約なので固定する。
func TestReturnedBytesAreCopies(t *testing.T) {
	s := SchemaV1()
	s[0] = 'X'
	if bytes.Equal(s, SchemaV1()) {
		t.Error("SchemaV1 の返り値の変更が次回呼び出しへ波及した(防御的コピーになっていない)")
	}
	v := IDVectorsV1()
	v[0] = 'X'
	if bytes.Equal(v, IDVectorsV1()) {
		t.Error("IDVectorsV1 の返り値の変更が次回呼び出しへ波及した(防御的コピーになっていない)")
	}
}
