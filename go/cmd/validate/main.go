// Command validate(niface-validate)は niface エンベロープを適合検証する。
//
// 使い方:
//
//	niface-validate [-schema PATH] [FILE ...]
//	niface-validate < envelope.json
//
// FILE を渡せば各ファイルを、省略すれば stdin を 1 文書として検証する。
// -schema 省略時は embed 済みの正本 schema を使う(リポジトリ外でも自立して
// 動く)。-schema 指定時はそのファイルを読む。
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
	schemaPath := flag.String("schema", "", "envelope schema(JSON)のパス(省略時は embed 済み正本 schema)")
	flag.Parse()
	os.Exit(run(*schemaPath, flag.Args(), os.Stdin, os.Stderr))
}

// run は検証本体。main から I/O(stdin / stderr)を注入可能にして、schema の
// 選択分岐と exit code 分類(適合 0 / 違反 1 / schema・入力の I/O エラー 2)を
// テストで固定できるようにする。
func run(schemaPath string, args []string, stdin io.Reader, stderr io.Writer) int {
	var chk *conformance.Checker
	var err error
	if schemaPath == "" {
		chk, err = conformance.NewDefaultChecker()
	} else {
		schemaJSON, readErr := os.ReadFile(schemaPath)
		if readErr != nil {
			fmt.Fprintf(stderr, "schema 読み込み: %v\n", readErr)
			return 2
		}
		chk, err = conformance.NewChecker(schemaJSON)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	type input struct {
		name string
		data []byte
	}
	var inputs []input
	if len(args) == 0 {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "stdin 読み込み: %v\n", err)
			return 2
		}
		inputs = append(inputs, input{"<stdin>", data})
	} else {
		for _, p := range args {
			data, err := os.ReadFile(p)
			if err != nil {
				fmt.Fprintf(stderr, "%s 読み込み: %v\n", p, err)
				return 2
			}
			inputs = append(inputs, input{p, data})
		}
	}

	conform := true
	for _, in := range inputs {
		findings := chk.Check(in.data)
		if len(findings) == 0 {
			fmt.Fprintf(stderr, "OK   %s\n", in.name)
			continue
		}
		conform = false
		for _, f := range findings {
			fmt.Fprintf(stderr, "FAIL %s: %s\n", in.name, f)
		}
	}
	if !conform {
		return 1
	}
	return 0
}
