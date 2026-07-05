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
}

func TestDeriveIDVectors(t *testing.T) {
	raw, err := os.ReadFile("../testdata/v1/id-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vf vectorFile
	if err := json.Unmarshal(raw, &vf); err != nil {
		t.Fatal(err)
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
