package spec

import (
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
	if SchemaV1JSON != string(source) {
		t.Error("embed した schema が正本 schema/v1/envelope.schema.json とバイト一致しない。go generate ./... で同期すること")
	}
}

func TestIDVectorsV1MatchesSource(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/v1/id-vectors.json")
	if err != nil {
		t.Fatalf("正本 id-vectors 読み込み: %v", err)
	}
	if IDVectorsV1JSON != string(source) {
		t.Error("embed した id-vectors が正本 testdata/v1/id-vectors.json とバイト一致しない。go generate ./... で同期すること")
	}
}
