package niface

// testdata/v1/valid の全エンベロープを Envelope 型へ decode する適合テスト。
// 規格型が valid ベクタの全フィールドを表現できること(未知フィールド禁止で decode)と、
// decode 結果の健全性(status の 2 値・必須フィールド非ゼロ・results 各要素の subject 存在等)を確認する。
// ツール固有の info には立ち入らないため、型パラメータは json.RawMessage で受ける。

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

type rawEnvelope = Envelope[json.RawMessage, json.RawMessage, json.RawMessage]

// id は sha256(JCS(identity)) の 64 文字 lowercase hex。
var idPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func TestDecodeValidEnvelopes(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "testdata", "v1", "valid", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no testdata found in ../testdata/v1/valid")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			dec := json.NewDecoder(bytes.NewReader(raw))
			dec.DisallowUnknownFields()
			var env rawEnvelope
			if err := dec.Decode(&env); err != nil {
				t.Fatalf("decode: %v", err)
			}
			checkEnvelope(t, &env)
		})
	}
}

func checkEnvelope(t *testing.T, env *rawEnvelope) {
	t.Helper()
	if env.SpecVersion != 1 {
		t.Errorf("specVersion: got %d want 1", env.SpecVersion)
	}
	if env.Tool.Name == "" {
		t.Error("tool.name is empty")
	}
	if env.Tool.Version == "" {
		t.Error("tool.version is empty")
	}
	if env.Command == "" {
		t.Error("command is empty")
	}
	if env.Status != StatusSuccess && env.Status != StatusError {
		t.Errorf("status: got %q want success or error", env.Status)
	}
	if env.StartedAt == "" {
		t.Error("startedAt is empty")
	}
	if env.FinishedAt == "" {
		t.Error("finishedAt is empty")
	}
	// results は必須。JSON の [] は非 nil スライスに decode されるので nil は欠落を意味する。
	if env.Results == nil {
		t.Error("results is missing")
	}
	checkErrors(t, "errors", env.Errors)
	for i, sr := range env.Results {
		checkSubjectResult(t, i, &sr)
	}
}

func checkSubjectResult(t *testing.T, i int, sr *SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]) {
	t.Helper()
	if sr.Subject.Name == "" {
		t.Errorf("results[%d]: subject.name is empty", i)
	}
	if sr.Status != StatusSuccess && sr.Status != StatusError {
		t.Errorf("results[%d]: status: got %q want success or error", i, sr.Status)
	}
	if sr.Generation != nil && sr.Generation.Profile == "" {
		t.Errorf("results[%d]: generation.profile is empty", i)
	}
	if sr.StartedAt == "" {
		t.Errorf("results[%d]: startedAt is empty", i)
	}
	if sr.FinishedAt == "" {
		t.Errorf("results[%d]: finishedAt is empty", i)
	}
	checkErrors(t, "errors", sr.Errors)
	if sr.Result.Items == nil {
		t.Errorf("results[%d]: result.items is missing", i)
	}
	for j, item := range sr.Result.Items {
		checkItem(t, i, j, &item)
	}
	for j, change := range sr.Result.Changes {
		checkChange(t, i, j, &change)
	}
}

func checkItem(t *testing.T, i, j int, item *Item[json.RawMessage]) {
	t.Helper()
	if !idPattern.MatchString(item.ID) {
		t.Errorf("results[%d].result.items[%d]: id %q is not 64-char lowercase hex", i, j, item.ID)
	}
	if item.Kind == "" {
		t.Errorf("results[%d].result.items[%d]: kind is empty", i, j)
	}
	switch item.Status {
	case ItemSuccess, ItemFailed, ItemSkipped:
	default:
		t.Errorf("results[%d].result.items[%d]: status: got %q want success, failed or skipped", i, j, item.Status)
	}
	if item.Status == ItemFailed && item.Error == nil {
		t.Errorf("results[%d].result.items[%d]: failed item has no error", i, j)
	}
	if item.Error != nil && item.Error.Code == "" {
		t.Errorf("results[%d].result.items[%d]: error.code is empty", i, j)
	}
	if item.Error != nil && item.Error.Message == "" {
		t.Errorf("results[%d].result.items[%d]: error.message is empty", i, j)
	}
}

func checkChange(t *testing.T, i, j int, change *Change[json.RawMessage]) {
	t.Helper()
	switch change.Kind {
	case ChangeAdd, ChangeRemove, ChangeModify:
	default:
		t.Errorf("results[%d].result.changes[%d]: kind: got %q want add, remove or modify", i, j, change.Kind)
	}
	if !idPattern.MatchString(change.ItemID) {
		t.Errorf("results[%d].result.changes[%d]: itemId %q is not 64-char lowercase hex", i, j, change.ItemID)
	}
}

func checkErrors(t *testing.T, label string, errs []Error) {
	t.Helper()
	for i, e := range errs {
		if e.Code == "" {
			t.Errorf("%s[%d]: code is empty", label, i)
		}
		if e.Message == "" {
			t.Errorf("%s[%d]: message is empty", label, i)
		}
	}
}
