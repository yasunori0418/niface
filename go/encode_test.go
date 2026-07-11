package niface

// golden encode 適合テスト。
// testdata/v1/valid の各エンベロープと同一内容を Go コードから構築して json.Marshal し、
// testdata と JSON 等価であることを検証する。decode → re-encode の round-trip では
// [] を非 nil の空 slice に decode してしまうため nil slice 事故を検出できないが、
// 構築ベースの本テストは必須配列(Results / Items)を nil のまま組み立てるケースを含み、
// MarshalJSON による [] 正規化が退行すると null が出力されて即座に失敗する。
//
// ツール固有 info には立ち入らないため、型パラメータは envelope_test.go の rawEnvelope
// (全 info を json.RawMessage で受ける)を再利用し、info は生 JSON バイト列として与える。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// testdata に現れる item id(sha256(JCS(identity)) の 64 文字 hex)。
const (
	idZshrc   = "2ba0b24fb64819dfe4bab60dd153fe5a636bdd5a7535d32039c1fe24736f97f8"
	idInitLua = "4f63987615f37911e2dbb1741f2d88ae288acf3ada4d5b625b54cd94f39429e3"
	idEsp     = "895da5a1a51d63456a4b5b88b0b0ba52646718d87c36edddec0c359ccf80286f"
)

// 複数ベクタで再利用する info 生 JSON。
const (
	infoZshrc   = `{"target":".zshrc","method":"symlink"}`
	infoInitLua = `{"target":".config/nvim/init.lua","method":"symlink"}`
	changeZshrc = `{"oldSrc":"/nix/store/aaa-zshrc","newSrc":"/nix/store/bbb-zshrc"}`
	infoGen42   = `{"generation":42}`
	infoGen7    = `{"generation":7}`
	profileHome = "/nix/var/nix/profiles/per-user/alice/home"
	profileWork = "/nix/var/nix/profiles/per-user/alice/work"
)

func rawMsg(s string) json.RawMessage { return json.RawMessage(s) }
func intPtr(i int) *int               { return &i }

// goldenEnvelopes は testdata/v1/valid の各ファイル名に対応する構築済みエンベロープ。
func goldenEnvelopes() map[string]rawEnvelope {
	return map[string]rawEnvelope{
		"apply-partial-failure-applied-changes.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusError,
			DryRun:      false,
			StartedAt:   "2026-07-07T09:15:00+09:00",
			FinishedAt:  "2026-07-07T09:15:03+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusError,
					StartedAt:  "2026-07-07T09:15:00+09:00",
					FinishedAt: "2026-07-07T09:15:03+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Info: rawMsg(infoZshrc)},
							{ID: idInitLua, Kind: "entry", Label: "init.lua", Status: ItemFailed, Error: &Error{Code: "E_IO", Message: "failed to create symlink after removing old target"}},
						},
						Changes: []Change[json.RawMessage]{
							{Kind: ChangeModify, ItemID: idZshrc, Reversible: true, Info: rawMsg(changeZshrc)},
							{Kind: ChangeRemove, ItemID: idInitLua, Reversible: true, Info: rawMsg(`{"oldSrc":"/nix/store/aaa-init.lua"}`)},
						},
					},
				},
			},
		},

		"apply-partial-failure.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusError,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:34:58+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusError,
					StartedAt:  "2026-07-05T12:34:56+09:00",
					FinishedAt: "2026-07-05T12:34:58+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Info: rawMsg(infoZshrc)},
							{ID: idInitLua, Kind: "entry", Label: "init.lua", Status: ItemFailed, Error: &Error{Code: "E_NPUT_COLLISION", Message: "target already exists"}},
						},
						Changes: []Change[json.RawMessage]{
							{Kind: ChangeModify, ItemID: idZshrc, Reversible: true, Info: rawMsg(changeZshrc)},
						},
						Info: rawMsg(infoGen42),
					},
				},
			},
		},

		"apply-success.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusSuccess,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:34:58+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-05T12:34:56+09:00",
					FinishedAt: "2026-07-05T12:34:58+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Info: rawMsg(infoZshrc)},
							{ID: idInitLua, Kind: "entry", Label: "init.lua", Status: ItemSuccess, Info: rawMsg(infoInitLua)},
						},
						Changes: []Change[json.RawMessage]{
							{Kind: ChangeModify, ItemID: idZshrc, Reversible: true, Info: rawMsg(changeZshrc)},
						},
						Info: rawMsg(infoGen42),
					},
				},
			},
		},

		"check-success.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nboot", Version: "0.1.0"},
			Command:     "check",
			Status:      StatusSuccess,
			DryRun:      true,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:34:58+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "system"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-05T12:34:56+09:00",
					FinishedAt: "2026-07-05T12:34:58+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idEsp, Kind: "requirement", Label: "esp-mounted", Status: ItemSuccess},
						},
					},
				},
			},
		},

		// Results を nil のまま構築する。MarshalJSON が [] へ正規化しなければ null になり失敗する。
		"empty-results.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusSuccess,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:34:56+09:00",
		},

		"enumeration-error.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusError,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:34:56+09:00",
			Errors:      []Error{{Code: "E_IO", Message: "cannot read the profile base directory"}},
		},

		"envelope-info.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusSuccess,
			StartedAt:   "2026-07-07T22:00:00+09:00",
			FinishedAt:  "2026-07-07T22:00:03+09:00",
			Info:        rawMsg(`{"configSource":"git+https://example.com/dotfiles.git","cacheHits":12}`),
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-07T22:00:00+09:00",
					FinishedAt: "2026-07-07T22:00:03+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Info: rawMsg(infoZshrc)},
						},
						Info: rawMsg(infoGen42),
					},
				},
			},
		},

		"envelope-warnings.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusSuccess,
			StartedAt:   "2026-07-07T22:10:00+09:00",
			FinishedAt:  "2026-07-07T22:10:02+09:00",
			Warnings: []Warning{
				{Code: "W_NPUT_DEPRECATED_FLAG", Message: "--force is deprecated; use --overwrite", Detail: map[string]any{"flag": "--force"}},
			},
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-07T22:10:00+09:00",
					FinishedAt: "2026-07-07T22:10:02+09:00",
					Warnings:   []Warning{{Code: "W_NPUT_PROFILE_LEGACY_LAYOUT", Message: "profile uses legacy layout"}},
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Warnings: []Warning{{Code: "W_NPUT_FOREIGN_SYMLINK", Message: "existing symlink was not created by nput"}}},
						},
					},
				},
			},
		},

		"generation-apply.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusSuccess,
			StartedAt:   "2026-07-07T21:00:00+09:00",
			FinishedAt:  "2026-07-07T21:00:02+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					Generation: &Generation{Profile: profileHome, Before: intPtr(41), After: intPtr(42)},
					StartedAt:  "2026-07-07T21:00:00+09:00",
					FinishedAt: "2026-07-07T21:00:01+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess},
						},
						Changes: []Change[json.RawMessage]{
							{Kind: ChangeModify, ItemID: idZshrc, Reversible: true},
						},
					},
				},
				{
					Subject:    Subject{Name: "work"},
					Status:     StatusSuccess,
					Generation: &Generation{Profile: profileWork, After: intPtr(1)},
					StartedAt:  "2026-07-07T21:00:01+09:00",
					FinishedAt: "2026-07-07T21:00:02+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idInitLua, Kind: "entry", Label: "init.lua", Status: ItemSuccess},
						},
						Changes: []Change[json.RawMessage]{
							{Kind: ChangeAdd, ItemID: idInitLua, Reversible: true},
						},
					},
				},
			},
		},

		"generation-plan.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "plan",
			Status:      StatusSuccess,
			DryRun:      true,
			StartedAt:   "2026-07-07T21:10:00+09:00",
			FinishedAt:  "2026-07-07T21:10:01+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					Generation: &Generation{Profile: profileHome, Before: intPtr(41), After: intPtr(41)},
					StartedAt:  "2026-07-07T21:10:00+09:00",
					FinishedAt: "2026-07-07T21:10:01+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess},
						},
						Changes: []Change[json.RawMessage]{
							{Kind: ChangeModify, ItemID: idZshrc, Reversible: true, Info: rawMsg(changeZshrc)},
						},
					},
				},
			},
		},

		"multi-subject-partial-failure.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusError,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:35:04+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-05T12:34:56+09:00",
					FinishedAt: "2026-07-05T12:34:59+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Info: rawMsg(infoZshrc)},
						},
					},
				},
				{
					Subject:    Subject{Name: "work"},
					Status:     StatusError,
					StartedAt:  "2026-07-05T12:35:00+09:00",
					FinishedAt: "2026-07-05T12:35:04+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idInitLua, Kind: "entry", Label: "init.lua", Status: ItemFailed, Error: &Error{Code: "E_NPUT_COLLISION", Message: "target already exists"}},
						},
					},
				},
			},
		},

		"multi-subject-success.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusSuccess,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:35:04+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-05T12:34:56+09:00",
					FinishedAt: "2026-07-05T12:34:59+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Info: rawMsg(infoZshrc)},
						},
						Info: rawMsg(infoGen42),
					},
				},
				{
					Subject:    Subject{Name: "work"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-05T12:35:00+09:00",
					FinishedAt: "2026-07-05T12:35:04+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idInitLua, Kind: "entry", Label: "init.lua", Status: ItemSuccess, Info: rawMsg(infoInitLua)},
						},
						Info: rawMsg(infoGen7),
					},
				},
			},
		},

		"plan-dryrun.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "plan",
			Status:      StatusSuccess,
			DryRun:      true,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:34:58+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusSuccess,
					StartedAt:  "2026-07-05T12:34:56+09:00",
					FinishedAt: "2026-07-05T12:34:58+09:00",
					Result: Result[json.RawMessage, json.RawMessage, json.RawMessage]{
						Items: []Item[json.RawMessage]{
							{ID: idZshrc, Kind: "entry", Label: ".zshrc", Status: ItemSuccess, Info: rawMsg(infoZshrc)},
							{ID: idInitLua, Kind: "entry", Label: "init.lua", Status: ItemSuccess, Info: rawMsg(infoInitLua)},
						},
						Changes: []Change[json.RawMessage]{
							{Kind: ChangeModify, ItemID: idZshrc, Reversible: true, Info: rawMsg(changeZshrc)},
						},
						Info: rawMsg(infoGen42),
					},
				},
			},
		},

		// result.items を nil のまま構築する。MarshalJSON が [] へ正規化しなければ null になり失敗する。
		"subject-error-lock.json": {
			SpecVersion: 1,
			Tool:        Tool{Name: "nput", Version: "0.9.0"},
			Command:     "apply",
			Status:      StatusError,
			StartedAt:   "2026-07-05T12:34:56+09:00",
			FinishedAt:  "2026-07-05T12:34:58+09:00",
			Results: []SubjectResult[json.RawMessage, json.RawMessage, json.RawMessage]{
				{
					Subject:    Subject{Name: "home"},
					Status:     StatusError,
					StartedAt:  "2026-07-05T12:34:56+09:00",
					FinishedAt: "2026-07-05T12:34:58+09:00",
					Errors:     []Error{{Code: "E_LOCK", Message: "another apply is in progress"}},
				},
			},
		},
	}
}

func TestEncodeGolden(t *testing.T) {
	golden := goldenEnvelopes()

	paths, err := filepath.Glob(filepath.Join("..", "testdata", "v1", "valid", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no testdata found in ../testdata/v1/valid")
	}

	seen := map[string]bool{}
	for _, path := range paths {
		name := filepath.Base(path)
		seen[name] = true
		t.Run(name, func(t *testing.T) {
			env, ok := golden[name]
			if !ok {
				t.Fatalf("golden envelope not defined for %s (testdata に対応する構築が無い)", name)
			}
			got, err := json.Marshal(env)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, got, want)
		})
	}

	// golden 側に testdata から消えた（または名前を誤った）エントリが残っていないことを確認する。
	for name := range golden {
		if !seen[name] {
			t.Errorf("golden envelope %q に対応する testdata が存在しない", name)
		}
	}
}

// TestMarshalNormalizesNilArrays は必須配列の nil → [] 正規化を testdata に依らず直接検証する。
// golden 側の JSON 等価比較でも間接的に担保されるが、退行時の原因を一目で分かるよう明示する。
func TestMarshalNormalizesNilArrays(t *testing.T) {
	// Results / Items 双方を nil のまま構築する。
	env := rawEnvelope{
		SpecVersion: 1,
		Tool:        Tool{Name: "nput", Version: "0.9.0"},
		Command:     "apply",
		Status:      StatusSuccess,
		StartedAt:   "2026-07-05T12:34:56+09:00",
		FinishedAt:  "2026-07-05T12:34:56+09:00",
	}
	got, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	var decoded struct {
		Results *json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Results == nil {
		t.Fatalf("results フィールドが出力されていない: %s", got)
	}
	if string(*decoded.Results) != "[]" {
		t.Errorf("nil Results が [] に正規化されていない: got %s", *decoded.Results)
	}

	// Result.Items 単体も同様に [] へ正規化されること。
	res := Result[json.RawMessage, json.RawMessage, json.RawMessage]{}
	rgot, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var rdec struct {
		Items *json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(rgot, &rdec); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if rdec.Items == nil || string(*rdec.Items) != "[]" {
		t.Errorf("nil Items が [] に正規化されていない: got %s", rgot)
	}
}

// assertJSONEqual は 2 つの JSON バイト列を意味的（キー順・空白差を無視）に比較する。
func assertJSONEqual(t *testing.T, got, want []byte) {
	t.Helper()
	var gv, wv any
	if err := json.Unmarshal(got, &gv); err != nil {
		t.Fatalf("unmarshal got: %v\n%s", err, got)
	}
	if err := json.Unmarshal(want, &wv); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(gv, wv) {
		g, _ := json.MarshalIndent(gv, "", "  ")
		w, _ := json.MarshalIndent(wv, "", "  ")
		t.Errorf("JSON not equal\n--- got ---\n%s\n--- want ---\n%s", g, w)
	}
}
