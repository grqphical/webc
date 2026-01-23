package main

import (
	"fmt"
	"os"

	"github.com/grqphical/webc/internal/lexer"
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

	fmt.Printf("tokens: %+v\n", tokens)
}
