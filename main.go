package main

import (
	"fmt"
	"os"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: webc [source_file]\n")
		os.Exit(1)
	}

	sourceCode, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read source file: %v\n", err)
		os.Exit(1)
	}

	lexer := lexer.New(string(sourceCode))
	tokens, err := lexer.ParseSource()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	parser := parser.New(tokens)
	program, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	module := codegen.NewModule(program)
	err = module.Generate()
	if err != nil {

		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	module.Save("output.wasm")

}
