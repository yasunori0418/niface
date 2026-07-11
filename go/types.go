package niface

import "encoding/json"

// エンベロープと構成型(specVersion 1)。
// トップレベルは常に Results[] を持ち、各要素は主体を Subject で名指す。
// ツール固有情報は Info の型パラメータで表現する。
// ツールを知らない消費側は json.RawMessage を渡して規格部分だけを扱える。

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
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

// Warning は警告。構造は Error と同形だが、code は W_ prefix に限定される（schema $defs/warning）。
type Warning struct {
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
	Warnings []Warning  `json:"warnings,omitempty"`
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

// Generation は profile 世代遷移の観測記録。
// Before / After は実行開始 / 終了時点で profile が指していた世代番号で、
// 観測できない場合は nil（初回実行の Before・profile 未作成の plan の After）。
type Generation struct {
	Profile string `json:"profile"`
	Before  *int   `json:"before,omitempty"`
	After   *int   `json:"after,omitempty"`
}

// SubjectResult は Results[] の要素。1 主体分の実行結果。
// Subject は single / batch を問わず常時必須。
type SubjectResult[TItem, TChange, TInfo any] struct {
	Subject    Subject                       `json:"subject"`
	Status     Status                        `json:"status"`
	Generation *Generation                   `json:"generation,omitempty"`
	StartedAt  string                        `json:"startedAt"`
	FinishedAt string                        `json:"finishedAt"`
	Errors     []Error                       `json:"errors,omitempty"`
	Warnings   []Warning                     `json:"warnings,omitempty"`
	Result     Result[TItem, TChange, TInfo] `json:"result"`
}

// Envelope の Info は実行全体のツール固有情報（主体に紐づかないもの）。
// 主体ごとのツール固有情報は Result.Info に置く。
type Envelope[TItem, TChange, TInfo, TEnvInfo any] struct {
	SpecVersion int                                    `json:"specVersion"`
	Tool        Tool                                   `json:"tool"`
	Command     string                                 `json:"command"`
	Status      Status                                 `json:"status"`
	DryRun      bool                                   `json:"dryRun"`
	StartedAt   string                                 `json:"startedAt"`
	FinishedAt  string                                 `json:"finishedAt"`
	Errors      []Error                                `json:"errors,omitempty"`
	Warnings    []Warning                              `json:"warnings,omitempty"`
	Info        TEnvInfo                               `json:"info,omitempty"`
	Results     []SubjectResult[TItem, TChange, TInfo] `json:"results"`
}

// MarshalJSON: 必須配列の nil を [] に正規化する。
//
// Envelope.Results / Result.Items は omitempty なしの必須配列(schema は type: array + required)。
// producer が nil slice のまま marshal すると null が出力され schema violation になる。
// この footgun を型レベルで塞ぐため、両型に MarshalJSON を実装して nil を空配列へ正規化する。
//
// generic type alias(type A[T] = B[T])は Go 1.24 要求のため使わず、method を持たない
// 別名の generic 定義型(シャドー)へ変換してから既定のフィールド直列化に委ねる shadow struct
// 方式で go 1.22 のまま実装する。シャドー型は Envelope / Result の method を継承しないため
// json.Marshal がそこで無限再帰しない。

// envelopeShadow は Envelope と同一フィールドだが MarshalJSON を持たないシャドー型。
type envelopeShadow[TItem, TChange, TInfo, TEnvInfo any] Envelope[TItem, TChange, TInfo, TEnvInfo]

func (e Envelope[TItem, TChange, TInfo, TEnvInfo]) MarshalJSON() ([]byte, error) {
	s := envelopeShadow[TItem, TChange, TInfo, TEnvInfo](e)
	if s.Results == nil {
		s.Results = []SubjectResult[TItem, TChange, TInfo]{}
	}
	return json.Marshal(s)
}

// resultShadow は Result と同一フィールドだが MarshalJSON を持たないシャドー型。
type resultShadow[TItem, TChange, TInfo any] Result[TItem, TChange, TInfo]

func (r Result[TItem, TChange, TInfo]) MarshalJSON() ([]byte, error) {
	s := resultShadow[TItem, TChange, TInfo](r)
	if s.Items == nil {
		s.Items = []Item[TItem]{}
	}
	return json.Marshal(s)
}
