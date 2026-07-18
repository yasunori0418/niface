package conformance

import (
	"bytes"
	"testing"
)

// NewDefaultChecker が schema ファイル無しで valid testdata を全通過させることを
// 固定する。embed と正本のバイト一致は internal/spec 側で検査済みだが、default
// 経路のコンパイル(AssertFormat 等の設定)を独立に固定するため、ファイル
// schema 経路(conformance_test.go)と同じ検証を default checker でも通す。
func TestNewDefaultCheckerAcceptsValidTestdata(t *testing.T) {
	c, err := NewDefaultChecker()
	if err != nil {
		t.Fatalf("NewDefaultChecker: %v", err)
	}
	assertAcceptsValid(t, c)
}

// default checker 自身が invalid を全件弾くことも直接検証する。バイト一致 +
// 正本での invalid 拒否からの帰結に頼らず、将来 default 経路だけコンパイル設定
// (AssertFormat 等)が分岐した場合も検知できるようにする。schema 層由来
// (RFC 3339・id 形式等)と lint 層由来の invalid を両方含む。
func TestNewDefaultCheckerRejectsInvalidTestdata(t *testing.T) {
	c, err := NewDefaultChecker()
	if err != nil {
		t.Fatalf("NewDefaultChecker: %v", err)
	}
	assertRejectsInvalid(t, c)
}

// 防御的コピー(呼び出し側の mutate が次回呼び出しへ波及しない)は公開 API の
// 契約なので固定する。
func TestSchemaV1IsDefensiveCopy(t *testing.T) {
	s := SchemaV1()
	if len(s) == 0 {
		t.Fatal("SchemaV1 が空を返した")
	}
	s[0] = 'X'
	if bytes.Equal(s, SchemaV1()) {
		t.Error("SchemaV1 の返り値の変更が次回呼び出しへ波及した")
	}
}
