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

// assertAcceptsValid は checker が valid testdata を全件違反ゼロで通すことを
// 検証する。検証ループはここに一本化し、ファイル schema 経路(本ファイル)と
// embed 経路(default_test.go)で共有する。
func assertAcceptsValid(t *testing.T, c *Checker) {
	t.Helper()
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

// assertRejectsInvalid は checker が invalid testdata を全件(schema か lint の
// いずれかで)弾くことを検証する。assertAcceptsValid と同じく両経路で共有する。
func assertRejectsInvalid(t *testing.T, c *Checker) {
	t.Helper()
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

// valid/ は schema + lint を全通過しなければならない。
func TestValidTestdataConforms(t *testing.T) {
	assertAcceptsValid(t, newChecker(t))
}

// startedAt / finishedAt の RFC 3339 は format assertion で強制する（§2, §8, ADR-0025）。
// Draft 2020-12 の既定（format = 注釈扱い）では素通りしてしまう不正日時が、schema 層で
// 弾かれることを保証する。top-level / subjectResult の 4 スロットそれぞれに不正値を入れ、
// 他 3 スロットは valid に保ち、finding が対応する JSON pointer を指すことまで確認する。
func TestFormatAssertionRejectsBadDateTime(t *testing.T) {
	c := newChecker(t)
	// 4 つの date-time スロットを個別に差し込むエンベロープテンプレート。
	envelope := func(topStart, topFinish, subStart, subFinish string) []byte {
		return fmt.Appendf(nil, `{
  "specVersion": 1,
  "tool": { "name": "nput", "version": "0.9.0" },
  "command": "apply",
  "status": "success",
  "dryRun": false,
  "startedAt": %q,
  "finishedAt": %q,
  "results": [
    {
      "subject": { "name": "home" },
      "status": "success",
      "startedAt": %q,
      "finishedAt": %q,
      "result": { "items": [] }
    }
  ]
}`, topStart, topFinish, subStart, subFinish)
	}
	const ok = "2026-07-05T12:34:56+09:00"
	// RFC 3339 の各違反を突く不正 date-time。オフセット欠落は §2 の MUST に対応する。
	bads := []struct {
		name, val string
	}{
		{"非日時文字列", "not-a-date"},
		{"月域外", "2026-13-05T12:34:56+09:00"},
		{"区切り文字が空白", "2026-07-05 12:34:56+09:00"},
		{"オフセット欠落", "2026-07-05T12:34:56"},
	}
	// 各スロットに不正値を注入し、finding がそのスロットの JSON pointer を指すことを検査する。
	slots := []struct {
		name, ptr string
		inject    func(bad string) []byte
	}{
		{"top-level startedAt", "/startedAt", func(b string) []byte { return envelope(b, ok, ok, ok) }},
		{"top-level finishedAt", "/finishedAt", func(b string) []byte { return envelope(ok, b, ok, ok) }},
		{"subjectResult startedAt", "/results/0/startedAt", func(b string) []byte { return envelope(ok, ok, b, ok) }},
		{"subjectResult finishedAt", "/results/0/finishedAt", func(b string) []byte { return envelope(ok, ok, ok, b) }},
	}
	for _, s := range slots {
		for _, bad := range bads {
			t.Run(s.name+"/"+bad.name, func(t *testing.T) {
				fs := c.Check(s.inject(bad.val))
				if len(fs) == 0 {
					t.Fatalf("不正な %s=%q が弾かれなかった", s.name, bad.val)
				}
				// santhosh は違反箇所を at '<pointer>' 形式で出す。引用符で囲んで
				// /startedAt と /results/0/startedAt の取り違えを防ぐ。
				if want := "'" + s.ptr + "'"; !strings.Contains(strings.Join(fs, "\n"), want) {
					t.Errorf("違反が %s を指していない: %v", s.ptr, fs)
				}
			})
		}
	}
}

// invalid/ は schema か lint のいずれかで必ず弾かれなければならない。
func TestInvalidTestdataRejected(t *testing.T) {
	assertRejectsInvalid(t, newChecker(t))
}
