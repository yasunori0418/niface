package niface

import "github.com/yasunori0418/niface/go/internal/spec"

// IDVectorsV1 は testdata/v1/id-vectors.json(正本とバイト同一)の生 bytes を返す。
// id 導出実装が全ベクタで期待値と一致することの検証(言語間互換の固定)に使う。
//
// デコード済み構造は提供しない。id-vectors の内部スキーマを外部契約として
// 凍結しないためで、consumer は必要な範囲だけを自前の型でデコードする。
// その際は json.Decoder の UseNumber を必ず有効にすること。標準の
// json.Unmarshal は JSON 数値を float64 へ落とすため、整数を含む identity の
// 表記判定(spec §5)が壊れ、DeriveID が全件エラーになる。
//
// 返り値は呼び出しごとに新しいスライスで、変更しても内部状態に影響しない。
func IDVectorsV1() []byte { return spec.IDVectorsV1() }
