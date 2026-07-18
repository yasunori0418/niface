// Package spec は正本(schema/v1・testdata/v1)の embed コピーを保持し、
// 公開 API(conformance.NewDefaultChecker / conformance.SchemaV1 /
// niface.IDVectorsV1)へ生 bytes を供給する。
//
// go module は github.com/yasunori0418/niface/go(サブディレクトリ module)で、
// go:embed は module ルートより上位のディレクトリを参照できない。正本を
// module 外(repo ルート)に置いたまま consumer へ届けるには、module 内への
// コピーの embed が必要になる(issue #42)。
//   - symlink は使えない(実測で確定): go:embed は symlink を irregular file
//     として拒否しコンパイルエラーになる。ディレクトリごと embed する場合は
//     symlink を黙ってスキップし、中身が空で埋め込まれる危険がある。
//   - 正本を go/ 配下へ移動する案も採らない: schema/ testdata/ が repo ルートに
//     並ぶ配置は README・spec・nix・CI 全部の前提であり、移動は規格側の
//     構造変更になる。コピー方式なら規格の所在を壊さない。
//
// コピーは正本から go generate で生成する(gen.go)。正本を変更したら
// go generate ./... を手動で実行して同期する(go build / go test は自動では
// 同期しない)。同期し忘れは spec_test.go のバイト完全一致検査が
// CI(nix flake check の checks.go)で検出して落とす。
package spec

import _ "embed"

// string で embed する。[]byte で embed すると、返したスライスへの呼び出し側の
// 書き込みが embed 実体まで汚染するため、string から呼び出しごとに新しい
// []byte を割り当てて返す(防御的コピー)。

//go:embed envelope.schema.json
var schemaV1 string

//go:embed id-vectors.json
var idVectorsV1 string

// SchemaV1 は schema/v1/envelope.schema.json(正本とバイト同一)の bytes を返す。
// 返り値は呼び出しごとに新しいスライスで、変更しても内部状態に影響しない。
func SchemaV1() []byte { return []byte(schemaV1) }

// IDVectorsV1 は testdata/v1/id-vectors.json(正本とバイト同一)の bytes を返す。
// 返り値は呼び出しごとに新しいスライスで、変更しても内部状態に影響しない。
func IDVectorsV1() []byte { return []byte(idVectorsV1) }
