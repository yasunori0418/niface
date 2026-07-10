package conformance

import (
	"os"
	"path/filepath"
	"testing"
)

// schema と testdata は repo ルート基準。テストは go/conformance/ を CWD に走る。
const (
	schemaPath   = "../../schema/v1/envelope.schema.json"
	testdataGlob = "../../testdata/v1"
)

func newChecker(t *testing.T) *Checker {
	t.Helper()
	schemaJSON, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("schema 読み込み: %v", err)
	}
	c, err := NewChecker(schemaJSON)
	if err != nil {
		t.Fatalf("schema コンパイル: %v", err)
	}
	return c
}

// valid/ は schema + lint を全通過しなければならない。
func TestValidTestdataConforms(t *testing.T) {
	c := newChecker(t)
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

// invalid/ は schema か lint のいずれかで必ず弾かれなければならない。
func TestInvalidTestdataRejected(t *testing.T) {
	c := newChecker(t)
	files, _ := filepath.Glob(filepath.Join(testdataGlob, "invalid", "*.json"))
	if len(files) == 0 {
		t.Fatal("invalid testdata が見つからない")
	}
	for _, p := range files {
		doc, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("%s 読み込み: %v", p, err)
		}
		if fs := c.Check(doc); len(fs) == 0 {
			t.Errorf("invalid %s が弾かれなかった", filepath.Base(p))
		}
	}
}
