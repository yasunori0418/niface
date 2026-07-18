package spec

// 正本 → embed コピーの同期。正本(schema/v1・testdata/v1)を変更したら
// go generate ./... を手動で実行して再生成する。go build / go test は
// トリガーしないため、同期し忘れは spec_test.go のバイト完全一致検査が
// CI で検出する。Windows は非対象のため cp 前提でよい(issue #42)。

//go:generate cp ../../../schema/v1/envelope.schema.json envelope.schema.json
//go:generate cp ../../../testdata/v1/id-vectors.json id-vectors.json
