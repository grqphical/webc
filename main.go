package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/grqphical/webc/internal/preprocessor"
)

const version string = "v0.2.0-alpha"

//go:embed templates/*
var templateFS embed.FS

func main() {
	compileForServer := flag.Bool("s", false, "Whether or not the runtime should be for the server instead of the browser")
	outputName := flag.String("o", "output.wasm", "Name/path of output binary")
	versionFlag := flag.Bool("v", false, "Prints the version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("webc %s\n", version)
		return
	}

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "missing input file\n")
		flag.Usage()
		os.Exit(1)
	}

	sourceCode, err := os.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read source file: %v\n", err)
		os.Exit(1)
	}

	preProcessor := preprocessor.New(templateFS)
	preProcessedSource, err := preProcessor.Parse(string(sourceCode))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while preprocessing file: %v\n", err)
		os.Exit(1)
	}

	lexer := lexer.New(preProcessedSource)

	parser := parser.New(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) != 0 {
		fmt.Println("Errors encountered while compiling:")
		for _, err := range parser.Errors() {
			fmt.Println(err.Error())
		}
		os.Exit(1)
	}

	module := codegen.NewModule(program)
	err = module.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	outputDir := filepath.Dir(*outputName)
	module.Save(*outputName)

	if !*compileForServer {
		htmlTemplate, err := templateFS.ReadFile("templates/index.html")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		err = os.WriteFile(filepath.Join(outputDir, "index.html"), htmlTemplate, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		tmpl, err := template.New("").ParseFS(templateFS, "templates/*.js", "templates/stdlib/lib/*.js")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		jsFile, err := os.OpenFile(filepath.Join(outputDir, "index.js"), os.O_RDWR|os.O_CREATE, 0644)
		defer jsFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		tmpl.ExecuteTemplate(jsFile, "browser-js", map[string]any{
			"BinaryName": filepath.Base(*outputName),
			"Server":     false,
		})
	} else {
		tmpl, err := template.New("").ParseFS(templateFS, "templates/*.js", "templates/stdlib/lib/*.js")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		jsFile, err := os.OpenFile(filepath.Join(outputDir, "index.js"), os.O_RDWR|os.O_CREATE, 0644)
		defer jsFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		tmpl.ExecuteTemplate(jsFile, "server-js", map[string]any{
			"BinaryName": filepath.Base(*outputName),
			"Server":     true,
		})

	}

}
