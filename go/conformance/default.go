package conformance

import "github.com/yasunori0418/niface/go/internal/spec"

// NewDefaultChecker は embed 済みの正本 schema(schema/v1/envelope.schema.json と
// バイト同一)をコンパイルした Checker を返す。consumer は schema ファイルを
// 持ち込まずに Check できる(主 API)。呼び出しごとにコンパイルするため、
// 複数文書を検査するときは返り値の Checker を使い回す。
func NewDefaultChecker() (*Checker, error) {
	return NewChecker([]byte(spec.SchemaV1JSON))
}

// SchemaV1 は embed 済みの正本 schema の生 bytes を返す(補助 API)。schema を
// 自前のコンパイル設定で使いたい場合や配布物へ含めたい場合に使う。
// 返り値は呼び出しごとに新しいスライスで、変更しても内部状態に影響しない。
func SchemaV1() []byte { return []byte(spec.SchemaV1JSON) }
