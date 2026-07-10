// Command validate(niface-validate)は niface エンベロープを適合検証する。
//
// 使い方:
//
//	niface-validate [-schema PATH] [FILE ...]
//	niface-validate < envelope.json
//
// FILE を渡せば各ファイルを、省略すれば stdin を 1 文書として検証する。
// -schema 省略時は schema/v1/envelope.schema.json(CWD 基準)を使う。
// 診断は stderr に出す。全て適合すれば exit 0、違反があれば exit 1、
// schema / 入力の I/O エラーは exit 2。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yasunori0418/niface/go/conformance"
)

func main() {
	schemaPath := flag.String("schema", "schema/v1/envelope.schema.json", "envelope schema(JSON)のパス")
	flag.Parse()

	schemaJSON, err := os.ReadFile(*schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema 読み込み: %v\n", err)
		os.Exit(2)
	}
	chk, err := conformance.NewChecker(schemaJSON)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	type input struct {
		name string
		data []byte
	}
	var inputs []input
	if flag.NArg() == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stdin 読み込み: %v\n", err)
			os.Exit(2)
		}
		inputs = append(inputs, input{"<stdin>", data})
	} else {
		for _, p := range flag.Args() {
			data, err := os.ReadFile(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s 読み込み: %v\n", p, err)
				os.Exit(2)
			}
			inputs = append(inputs, input{p, data})
		}
	}

	conform := true
	for _, in := range inputs {
		findings := chk.Check(in.data)
		if len(findings) == 0 {
			fmt.Fprintf(os.Stderr, "OK   %s\n", in.name)
			continue
		}
		conform = false
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "FAIL %s: %s\n", in.name, f)
		}
	}
	if !conform {
		os.Exit(1)
	}
}
