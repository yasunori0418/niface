package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run の schema 選択分岐(-schema 省略時は embed / 指定時はファイル)と
// exit code 分類(適合 0 / 違反 1 / I/O エラー 2)を固定する。
// schema と testdata は repo ルート基準(テストは go/cmd/validate/ を CWD に走る)。
const (
	schemaPath  = "../../../schema/v1/envelope.schema.json"
	validGlob   = "../../../testdata/v1/valid/*.json"
	invalidGlob = "../../../testdata/v1/invalid/*.json"
)

func firstMatch(t *testing.T, pattern string) string {
	t.Helper()
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		t.Fatalf("testdata が見つからない: %s", pattern)
	}
	return files[0]
}

func TestRunEmbeddedSchemaAcceptsValid(t *testing.T) {
	var stderr bytes.Buffer
	if got := run("", []string{firstMatch(t, validGlob)}, strings.NewReader(""), &stderr); got != 0 {
		t.Errorf("exit=%d want 0 (stderr: %s)", got, stderr.String())
	}
}

func TestRunEmbeddedSchemaRejectsInvalid(t *testing.T) {
	var stderr bytes.Buffer
	if got := run("", []string{firstMatch(t, invalidGlob)}, strings.NewReader(""), &stderr); got != 1 {
		t.Errorf("exit=%d want 1 (stderr: %s)", got, stderr.String())
	}
}

func TestRunExplicitSchemaAcceptsValid(t *testing.T) {
	var stderr bytes.Buffer
	if got := run(schemaPath, []string{firstMatch(t, validGlob)}, strings.NewReader(""), &stderr); got != 0 {
		t.Errorf("exit=%d want 0 (stderr: %s)", got, stderr.String())
	}
}

func TestRunStdinAcceptsValid(t *testing.T) {
	doc, err := os.ReadFile(firstMatch(t, validGlob))
	if err != nil {
		t.Fatalf("testdata 読み込み: %v", err)
	}
	var stderr bytes.Buffer
	if got := run("", nil, bytes.NewReader(doc), &stderr); got != 0 {
		t.Errorf("exit=%d want 0 (stderr: %s)", got, stderr.String())
	}
}

func TestRunSchemaReadErrorExits2(t *testing.T) {
	var stderr bytes.Buffer
	if got := run("no/such/schema.json", nil, strings.NewReader("{}"), &stderr); got != 2 {
		t.Errorf("exit=%d want 2 (stderr: %s)", got, stderr.String())
	}
}

// schema ファイルは読めるが JSON として parse できない経路も exit 2。
func TestRunSchemaParseErrorExits2(t *testing.T) {
	broken := filepath.Join(t.TempDir(), "broken.schema.json")
	if err := os.WriteFile(broken, []byte("{"), 0o644); err != nil {
		t.Fatalf("壊れた schema の書き込み: %v", err)
	}
	var stderr bytes.Buffer
	if got := run(broken, nil, strings.NewReader("{}"), &stderr); got != 2 {
		t.Errorf("exit=%d want 2 (stderr: %s)", got, stderr.String())
	}
}

// JSON としては妥当だが JSON Schema として compile に失敗する経路も exit 2
// (parse 失敗とは別の同値クラス)。解決できない $ref は compile 段階で落ちる。
func TestRunSchemaCompileErrorExits2(t *testing.T) {
	broken := filepath.Join(t.TempDir(), "uncompilable.schema.json")
	if err := os.WriteFile(broken, []byte(`{"$ref": "#/does/not/exist"}`), 0o644); err != nil {
		t.Fatalf("壊れた schema の書き込み: %v", err)
	}
	var stderr bytes.Buffer
	if got := run(broken, nil, strings.NewReader("{}"), &stderr); got != 2 {
		t.Errorf("exit=%d want 2 (stderr: %s)", got, stderr.String())
	}
}

// errReader は常に失敗する Reader。stdin 読み込みエラー経路を突く。
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

func TestRunStdinReadErrorExits2(t *testing.T) {
	var stderr bytes.Buffer
	if got := run("", nil, errReader{}, &stderr); got != 2 {
		t.Errorf("exit=%d want 2 (stderr: %s)", got, stderr.String())
	}
}

func TestRunInputReadErrorExits2(t *testing.T) {
	var stderr bytes.Buffer
	if got := run("", []string{"no/such/envelope.json"}, strings.NewReader(""), &stderr); got != 2 {
		t.Errorf("exit=%d want 2 (stderr: %s)", got, stderr.String())
	}
}
