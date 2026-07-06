package niface

// エンベロープと構成型(specVersion 1)。
// single / batch を問わずトップレベルは Results[] を持つ。
// ツール固有情報は Info の型パラメータで表現する。
// ツールを知らない消費側は json.RawMessage を渡して規格部分だけを扱える。

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

type Mode string

const (
	ModeSingle Mode = "single"
	ModeBatch  Mode = "batch"
)

type ItemStatus string

const (
	ItemSuccess ItemStatus = "success"
	ItemFailed  ItemStatus = "failed"
	ItemSkipped ItemStatus = "skipped"
)

type ChangeKind string

const (
	ChangeAdd    ChangeKind = "add"
	ChangeRemove ChangeKind = "remove"
	ChangeModify ChangeKind = "modify"
)

type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Subject は操作の主体を名指す弱い識別子。id 導出には関与しない。
type Subject struct {
	Name string `json:"name"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Detail  map[string]any `json:"detail,omitempty"`
}

type Item[T any] struct {
	ID       string     `json:"id"`
	Kind     string     `json:"kind"`
	Label    string     `json:"label,omitempty"`
	Status   ItemStatus `json:"status"`
	Error    *Error     `json:"error,omitempty"`
	Warnings []Error    `json:"warnings,omitempty"`
	Info     T          `json:"info,omitempty"`
}

type Change[T any] struct {
	Kind       ChangeKind `json:"kind"`
	ItemID     string     `json:"itemId"`
	Reversible bool       `json:"reversible"`
	Info       T          `json:"info,omitempty"`
}

type Result[TItem, TChange, TInfo any] struct {
	Items   []Item[TItem]     `json:"items"`
	Changes []Change[TChange] `json:"changes,omitempty"`
	Info    TInfo             `json:"info,omitempty"`
}

// SubjectResult は Results[] の要素。1 主体分の実行結果。
// Subject は single では任意(nil 可)、batch では必須。
type SubjectResult[TItem, TChange, TInfo any] struct {
	Subject    *Subject                      `json:"subject,omitempty"`
	Status     Status                        `json:"status"`
	StartedAt  string                        `json:"startedAt"`
	FinishedAt string                        `json:"finishedAt"`
	Errors     []Error                       `json:"errors,omitempty"`
	Result     Result[TItem, TChange, TInfo] `json:"result"`
}

type Envelope[TItem, TChange, TInfo any] struct {
	SpecVersion int                                    `json:"specVersion"`
	Tool        Tool                                   `json:"tool"`
	Command     string                                 `json:"command"`
	Mode        Mode                                   `json:"mode"`
	Status      Status                                 `json:"status"`
	DryRun      bool                                   `json:"dryRun"`
	StartedAt   string                                 `json:"startedAt"`
	FinishedAt  string                                 `json:"finishedAt"`
	Errors      []Error                                `json:"errors,omitempty"`
	Results     []SubjectResult[TItem, TChange, TInfo] `json:"results"`
}
