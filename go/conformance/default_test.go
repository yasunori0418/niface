package conformance

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// NewDefaultChecker が schema ファイル無しで valid testdata を全通過させることを
// 固定する。embed と正本のバイト一致は internal/spec 側で検査済み。
func TestNewDefaultCheckerAcceptsValidTestdata(t *testing.T) {
	c, err := NewDefaultChecker()
	if err != nil {
		t.Fatalf("NewDefaultChecker: %v", err)
	}
	files, _ := filepath.Glob(filepath.Join(testdataGlob, "valid", "*.json"))
	if len(files) == 0 {
		t.Fatal("valid testdata が見つからない")
	}
	for _, p := range files {
		doc, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("%s 読み込み: %v", p, err)
		}
		if fs := c.Check(doc); len(fs) > 0 {
			t.Errorf("valid %s が違反を出した: %v", filepath.Base(p), fs)
		}
	}
}

// default checker 自身が invalid を弾くことも 1 件直接検証する。バイト一致 +
// 正本での invalid 拒否からの帰結に頼らず、将来 default 経路だけコンパイル設定
// (AssertFormat 等)が分岐した場合も検知できるようにする。
func TestNewDefaultCheckerRejectsInvalidTestdata(t *testing.T) {
	c, err := NewDefaultChecker()
	if err != nil {
		t.Fatalf("NewDefaultChecker: %v", err)
	}
	files, _ := filepath.Glob(filepath.Join(testdataGlob, "invalid", "*.json"))
	if len(files) == 0 {
		t.Fatal("invalid testdata が見つからない")
	}
	doc, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("%s 読み込み: %v", files[0], err)
	}
	if fs := c.Check(doc); len(fs) == 0 {
		t.Errorf("invalid %s が違反ゼロで通過した", filepath.Base(files[0]))
	}
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
