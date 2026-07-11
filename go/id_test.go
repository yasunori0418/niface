package niface

// id-vectors.json の全ベクタ通過が適合条件。
// go test ./... を CI で必ず実行すること。

import (
	"encoding/json"
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

func loadVectors(t *testing.T) vectorFile {
	t.Helper()
	raw, err := os.ReadFile("../testdata/v1/id-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vf vectorFile
	if err := json.Unmarshal(raw, &vf); err != nil {
		t.Fatal(err)
	}
	return vf
}

func TestDeriveIDVectors(t *testing.T) {
	vf := loadVectors(t)
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
// id 導出時にエラーで拒否されることを検証する(spec §5)。
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
// JSON 由来の値は float64 になるため、ネイティブ整数経路はここで直接突く。
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
	outDomain := []Identity{
		{Kind: "n", Key: int64(maxSafe) + 1},
		{Kind: "n", Key: -int64(maxSafe) - 1},
	}
	for _, id := range outDomain {
		if _, err := DeriveID(id); err == nil {
			t.Errorf("out-of-domain integer %v: expected error, got nil", id.Key)
		}
	}
}
