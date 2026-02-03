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
)

const version string = "v0.1.0-beta"

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

	lexer := lexer.New(string(sourceCode))
	tokens := lexer.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	//fmt.Printf("tokens: %+v\n", tokens)

	parser := parser.New(tokens)
	program, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	//fmt.Printf("program: %+v\n", program)

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

		jsTemplate, err := templateFS.ReadFile("templates/browser.js")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		tmpl := template.Must(template.New("browser.js").Parse(string(jsTemplate)))

		jsFile, err := os.OpenFile(filepath.Join(outputDir, "index.js"), os.O_RDWR|os.O_CREATE, 0644)
		defer jsFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		tmpl.Execute(jsFile, map[string]any{
			"BinaryName": filepath.Base(*outputName),
		})
	} else {
		jsTemplate, err := templateFS.ReadFile("templates/server.js")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		tmpl := template.Must(template.New("server.js").Parse(string(jsTemplate)))

		jsFile, err := os.OpenFile(filepath.Join(outputDir, "index.js"), os.O_RDWR|os.O_CREATE, 0644)
		defer jsFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		tmpl.Execute(jsFile, map[string]any{
			"BinaryName": filepath.Base(*outputName),
		})

	}

}
