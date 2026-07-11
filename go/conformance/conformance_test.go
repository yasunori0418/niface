package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// startedAt / finishedAt の RFC 3339 は format assertion で強制する（§2, §8, ADR-0025）。
// Draft 2020-12 の既定（format = 注釈扱い）では素通りしてしまう不正日時が、schema 層で
// 弾かれることを保証する。他フィールドは valid に保ち、日時 1 つだけを壊す。
func TestFormatAssertionRejectsBadDateTime(t *testing.T) {
	c := newChecker(t)
	base := `{
  "specVersion": 1,
  "tool": { "name": "nput", "version": "0.9.0" },
  "command": "apply",
  "status": "success",
  "dryRun": false,
  "startedAt": %q,
  "finishedAt": "2026-07-05T12:34:56+09:00",
  "results": []
}`
	cases := []struct {
		name      string
		startedAt string
	}{
		{"非日時文字列", "not-a-date"},
		{"月域外", "2026-13-05T12:34:56+09:00"},
		{"区切り文字が空白", "2026-07-05 12:34:56+09:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := fmt.Appendf(nil, base, tc.startedAt)
			fs := c.Check(doc)
			if len(fs) == 0 {
				t.Fatalf("不正な startedAt %q が弾かれなかった", tc.startedAt)
			}
			if !strings.Contains(strings.Join(fs, "\n"), "startedAt") {
				t.Errorf("違反が startedAt を指していない: %v", fs)
			}
		})
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
