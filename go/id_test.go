package niface

// id-vectors.json の全ベクタ通過が適合条件。
// go test ./... を CI で必ず実行すること。

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"
)

type vectorFile struct {
	Vectors []struct {
		Identity struct {
			Kind string `json:"kind"`
			Key  any    `json:"key"`
		} `json:"identity"`
		Canonical string `json:"canonical"`
		Expected  string `json:"expected"`
	} `json:"vectors"`
	Rejected []struct {
		Identity struct {
			Kind string `json:"kind"`
			Key  any    `json:"key"`
		} `json:"identity"`
		Reason string `json:"reason"`
	} `json:"rejected"`
}

// loadVectors は id-vectors.json を UseNumber でデコードする。JSON 数値を
// json.Number(元の字句表現)として保持し、表記(1 と 1.0)を区別するため
// (spec §5・DeriveID の表記判定と揃える)。
func loadVectors(t *testing.T) vectorFile {
	t.Helper()
	raw, err := os.ReadFile("../testdata/v1/id-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var vf vectorFile
	if err := dec.Decode(&vf); err != nil {
		t.Fatal(err)
	}
	return vf
}

func TestDeriveIDVectors(t *testing.T) {
	vf := loadVectors(t)
	if len(vf.Vectors) == 0 {
		t.Fatal("no vectors found in id-vectors.json")
	}
	for i, v := range vf.Vectors {
		got, err := DeriveID(Identity{Kind: v.Identity.Kind, Key: v.Identity.Key})
		if err != nil {
			t.Errorf("vector %d: %v", i, err)
			continue
		}
		if got != v.Expected {
			t.Errorf("vector %d: got %s want %s (canonical: %s)", i, got, v.Expected, v.Canonical)
		}
	}
}

// TestDeriveIDRejectedVectors は域外 identity(rejected ベクタ)が
// id 導出時にエラーで拒否されることを検証する(spec §5・言語間で拒否を担保)。
func TestDeriveIDRejectedVectors(t *testing.T) {
	vf := loadVectors(t)
	if len(vf.Rejected) == 0 {
		t.Fatal("no rejected vectors found in id-vectors.json")
	}
	for i, v := range vf.Rejected {
		if _, err := DeriveID(Identity{Kind: v.Identity.Kind, Key: v.Identity.Key}); err == nil {
			t.Errorf("rejected vector %d: expected error but got none (reason: %s)", i, v.Reason)
		}
	}
}

// TestDeriveIDIntegerDomain は int / int64 経路の ±(2^53−1) ガードを検証する。
// JSON 由来の値は json.Number になるため、ネイティブ整数経路はここで直接突く。
func TestDeriveIDIntegerDomain(t *testing.T) {
	const maxSafe = 1<<53 - 1 // 2^53−1
	inDomain := []Identity{
		{Kind: "n", Key: 0},
		{Kind: "n", Key: 1},
		{Kind: "n", Key: int64(maxSafe)},
		{Kind: "n", Key: int64(-maxSafe)},
	}
	for _, id := range inDomain {
		if _, err := DeriveID(id); err != nil {
			t.Errorf("in-domain integer %v: unexpected error: %v", id.Key, err)
		}
	}
	// case int / case int64 の域外を両方突く。int の域外値は非定数の int64 変数を
	// 経由して作る。int(int64(maxSafe)+1) と定数式で書くと 32bit 環境では定数が
	// int の範囲を超えてコンパイルエラーになるため、実行時変換で回避する
	// (64bit 環境では int が 64bit のため域外値として有効に突ける)。
	over := int64(maxSafe) + 1
	bigInt := int(over)
	outDomain := []Identity{
		{Kind: "n", Key: int64(maxSafe) + 1},
		{Kind: "n", Key: -int64(maxSafe) - 1},
		{Kind: "n", Key: bigInt},
	}
	for _, id := range outDomain {
		_, err := DeriveID(id)
		if err == nil {
			t.Errorf("out-of-domain integer %v: expected error, got nil", id.Key)
			continue
		}
		if !errors.Is(err, errIntegerRange) {
			t.Errorf("out-of-domain integer %v: want errIntegerRange, got %v", id.Key, err)
		}
	}
}

// TestDeriveIDNumberNotation は json.Number 経路の表記判定を検証する(spec §5)。
// 整数表記のみ受理し、小数点・指数表記(値が整数でも)・範囲外は拒否する。
func TestDeriveIDNumberNotation(t *testing.T) {
	accept := []json.Number{"0", "1", "9007199254740991", "-9007199254740991"}
	for _, n := range accept {
		if _, err := DeriveID(Identity{Kind: "n", Key: n}); err != nil {
			t.Errorf("json.Number(%q): unexpected error: %v", n, err)
		}
	}
	// 拒否と分類(errors.Is)。
	reject := []struct {
		n    json.Number
		want error
	}{
		{"1.0", errNonInteger},
		{"1e3", errNonInteger},
		{"1E3", errNonInteger},
		{"10e-1", errNonInteger},
		{"9007199254740992", errIntegerRange},     // 2^53
		{"-9007199254740992", errIntegerRange},    // −2^53
		{"99999999999999999999", errIntegerRange}, // ParseInt overflow
	}
	for _, tc := range reject {
		_, err := DeriveID(Identity{Kind: "n", Key: tc.n})
		if err == nil {
			t.Errorf("json.Number(%q): expected error, got nil", tc.n)
			continue
		}
		if !errors.Is(err, tc.want) {
			t.Errorf("json.Number(%q): want %v, got %v", tc.n, tc.want, err)
		}
	}
}

// TestDeriveIDDomainErrors は各値域違反が対応する sentinel error に分類される
// ことを検証する(拒否理由の区別)。
func TestDeriveIDDomainErrors(t *testing.T) {
	cases := []struct {
		name string
		key  any
		want error
	}{
		{"float64 integer-valued", float64(1), errFloatType},
		{"float64 fractional", float64(1.5), errFloatType},
		{"json.Number fractional", json.Number("1.5"), errNonInteger},
		{"json.Number out of range", json.Number("9007199254740992"), errIntegerRange},
		{"native int64 out of range", int64(1 << 53), errIntegerRange},
		{"non-ASCII member name", map[string]any{"キー": "v"}, errNonASCIIKey},
		{"nested non-ASCII member", map[string]any{"items": []any{map[string]any{"🔑": "v"}}}, errNonASCIIKey},
		{"unsupported type", int32(1), errUnsupported},
	}
	for _, tc := range cases {
		_, err := DeriveID(Identity{Kind: "n", Key: tc.key})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
			continue
		}
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: want %v, got %v", tc.name, tc.want, err)
		}
	}
}
