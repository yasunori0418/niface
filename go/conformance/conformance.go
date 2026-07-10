// Package conformance は niface エンベロープの適合検証を行う。
//
// 適合は 2 層で判定する（→ spec/v1/spec.md §8, ADR-0021, ADR-0023）:
//   - schema 検証: schema/v1/envelope.schema.json（Draft 2020-12）への適合
//   - lint 検査: schema で表現しきれない MUST（§2 status 整合は schema 側、
//     §5 の itemId 参照整合・subject.name / item.id 一意性は本パッケージ）
package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// schemaURI は schema をコンパイラへ登録する内部 URI。schema の $id とは独立で、
// 外部 $ref を持たないため fragment 参照（#/$defs/...）の解決に影響しない。
const schemaURI = "mem:///niface/envelope.schema.json"

// Checker は schema を 1 度コンパイルして複数文書を検査する。
type Checker struct {
	schema *jsonschema.Schema
}

// NewChecker は schema JSON をコンパイルした Checker を返す。
func NewChecker(schemaJSON []byte) (*Checker, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		return nil, fmt.Errorf("schema の JSON parse: %w", err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaURI, doc); err != nil {
		return nil, fmt.Errorf("schema の登録: %w", err)
	}
	sch, err := c.Compile(schemaURI)
	if err != nil {
		return nil, fmt.Errorf("schema のコンパイル: %w", err)
	}
	return &Checker{schema: sch}, nil
}

// Check は 1 エンベロープ文書を schema + lint で検査し、違反メッセージを返す。
// 空スライスなら適合。
func (c *Checker) Check(docJSON []byte) []string {
	var findings []string

	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(docJSON))
	if err != nil {
		return []string{fmt.Sprintf("JSON parse: %v", err)}
	}
	if err := c.schema.Validate(inst); err != nil {
		findings = append(findings, "schema: "+err.Error())
	}

	var doc envelope
	if err := json.Unmarshal(docJSON, &doc); err != nil {
		findings = append(findings, fmt.Sprintf("lint: decode 不能: %v", err))
		return findings
	}
	findings = append(findings, lint(doc)...)
	return findings
}

// envelope は lint に必要な最小構造。規格の全フィールドは持たない。
type envelope struct {
	Results []struct {
		Subject struct {
			Name string `json:"name"`
		} `json:"subject"`
		Result struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Changes []struct {
				ItemID string `json:"itemId"`
			} `json:"changes"`
		} `json:"result"`
	} `json:"results"`
}

// lint は schema で表現しきれない §5 MUST を検査する。
func lint(doc envelope) []string {
	var f []string

	// subject.name の 1 エンベロープ内一意性（producer MUST）
	seenNames := map[string]bool{}
	for _, r := range doc.Results {
		if seenNames[r.Subject.Name] {
			f = append(f, fmt.Sprintf("subject.name が 1 エンベロープ内で重複: %q", r.Subject.Name))
		}
		seenNames[r.Subject.Name] = true
	}

	for i, r := range doc.Results {
		// item.id の 1 result 内一意性（§5）
		ids := map[string]bool{}
		for _, it := range r.Result.Items {
			if ids[it.ID] {
				f = append(f, fmt.Sprintf("results[%d] の item.id が result 内で重複: %q", i, it.ID))
			}
			ids[it.ID] = true
		}
		// changes[].itemId の同一 result 内参照整合（§5）
		for _, c := range r.Result.Changes {
			if !ids[c.ItemID] {
				f = append(f, fmt.Sprintf("results[%d] の changes.itemId が同一 result の item を指さない: %q", i, c.ItemID))
			}
		}
	}
	return f
}
